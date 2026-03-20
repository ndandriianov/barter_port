package application

import (
	"barter-port/internal/auth/domain"
	"barter-port/internal/auth/infrastructure/repository/email_token"
	"barter-port/internal/auth/infrastructure/repository/outbox"
	"barter-port/internal/auth/infrastructure/repository/user"
	"barter-port/internal/libs/db"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

const (
	bcryptCost          = 12
	minPasswordLength   = 6
	tokenLength         = 32
	tokenExpirationTime = 24 * time.Hour
	tokenUrlPath        = "/verify-email?token="
	subject             = "Confirm your email"
)

var (
	ErrInvalidEmail      = errors.New("invalid email")
	ErrPasswordTooShort  = errors.New("password too short")
	ErrEmailAlreadyInUse = errors.New("email already in use")

	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidEmailToken = errors.New("invalid email_token")
	ErrEmailTokenExpired = errors.New("email_token expired")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrIncorrectPassword  = errors.New("incorrect password")
)

type UserRepo interface {
	Create(ctx context.Context, exec db.DB, user domain.User) error
	GetByEmail(ctx context.Context, exec db.DB, email string) (domain.User, error)
	GetByID(ctx context.Context, exec db.DB, id uuid.UUID) (domain.User, error)
	VerifyEmailIfNotVerified(ctx context.Context, exec db.DB, userID uuid.UUID) (changed bool, err error)
}

type TokenRepo interface {
	Save(ctx context.Context, exec db.DB, t domain.EmailVerificationToken) error
	GetByHashForUpdate(ctx context.Context, exec db.DB, tokenHash string) (domain.EmailVerificationToken, error)
	MarkUsed(ctx context.Context, exec db.DB, tokenHash string) error
	DeleteAllForUser(ctx context.Context, exec db.DB, userID uuid.UUID) error
}

type Mailer interface {
	Send(to, subject, body string) error
}

type Service struct {
	db     *pgxpool.Pool
	users  UserRepo
	tokens TokenRepo
	mailer Mailer
	logger *slog.Logger
	outbox *outbox.Repository

	frontendBaseURL string
	re              *regexp.Regexp
}

func NewService(
	db *pgxpool.Pool,
	users UserRepo,
	tokens TokenRepo,
	mailer Mailer,
	logger *slog.Logger,
	outbox *outbox.Repository,

	frontendBaseURL string,
	re *regexp.Regexp,
) *Service {
	if logger == nil {
		log.Fatal("logger is required")
	}

	return &Service{
		db:     db,
		users:  users,
		tokens: tokens,
		mailer: mailer,
		logger: logger,
		outbox: outbox,

		frontendBaseURL: strings.TrimRight(frontendBaseURL, "/"),
		re:              re,
	}
}

type RegisterResult struct {
	UserID uuid.UUID
	Email  string
}

// Register creates a new user, generates an email verification token,
// and sends a verification email.
//
// It returns the following domain errors:
//   - ErrInvalidEmail
//   - ErrPasswordTooShort
//   - ErrEmailAlreadyInUse
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) Register(ctx context.Context, email, password string) (RegisterResult, error) {
	if err := s.validateCredentials(email, password); err != nil {
		return RegisterResult{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("failed to hash password: %w", err)
	}

	u := domain.NewUser(uuid.New(), email, string(hash))
	var rawToken string

	err = db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.users.Create(ctx, tx, u); err != nil {
			if errors.Is(err, user.ErrEmailAlreadyInUse) {
				return ErrEmailAlreadyInUse
			}
			return fmt.Errorf("failed to create user: %w", err)
		}

		rawToken, err = generateToken(tokenLength)
		if err != nil {
			return fmt.Errorf("failed to generate email_token: %w", err)
		}

		tokenHash := getHashFromToken(rawToken)
		t := domain.NewEmailVerificationToken(tokenHash, u.ID, time.Now().Add(tokenExpirationTime))

		if err = s.tokens.Save(ctx, tx, t); err != nil {
			return fmt.Errorf("failed to save email_token: %w", err)
		}

		return nil
	})

	if err = s.mailer.Send(u.Email, subject, s.getEmailBody(rawToken)); err != nil {
		s.logger.Warn("failed to send email", slog.Any("error", err))
	}

	return RegisterResult{
		UserID: u.ID,
		Email:  u.Email,
	}, nil
}

// RetrySendVerificationEmail generates a new email verification token and sends a verification email if the user's email is not verified.
//
// It returns the following domain errors:
//   - ErrInvalidCredentials
//   - ErrEmailNotVerified
//   - ErrIncorrectPassword
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) RetrySendVerificationEmail(ctx context.Context, email, password string) error {
	u, err := s.verifyCredentials(ctx, email, password)
	if err != nil {
		return err
	}

	if u.EmailVerified {
		return nil
	}

	rawToken, err := generateToken(tokenLength)
	if err != nil {
		return fmt.Errorf("failed to generate email_token: %w", err)
	}

	tokenHash := getHashFromToken(rawToken)
	t := domain.NewEmailVerificationToken(tokenHash, u.ID, time.Now().Add(tokenExpirationTime))

	return db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		if err = s.tokens.Save(ctx, tx, t); err != nil {
			return fmt.Errorf("failed to save email_token: %w", err)
		}

		if err = s.mailer.Send(u.Email, subject, s.getEmailBody(rawToken)); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		return nil
	})
}

// VerifyEmail marks user's email as verified if the provided token is valid.
//
// It returns the following domain errors:
//   - ErrInvalidEmailToken
//   - ErrEmailTokenExpired
//   - ErrUserNotFound
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	tokenHash, err := getHashFromRawToken(rawToken)
	if err != nil {
		return fmt.Errorf("cannot get hash from raw email_token: %w", err)
	}

	err = db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		t, err := s.tokens.GetByHashForUpdate(ctx, tx, tokenHash)
		if err != nil {
			if errors.Is(err, email_token.ErrTokenNotFound) {
				return ErrInvalidEmailToken
			}
			return fmt.Errorf("failed to get email_token by hash: %w", err)
		}

		if t.Used {
			return nil
		}
		if time.Now().After(t.ExpiresAt) {
			return ErrEmailTokenExpired
		}

		changed, err := s.users.VerifyEmailIfNotVerified(ctx, tx, t.UserID)
		if err != nil {
			if errors.Is(err, user.ErrUserNotFound) {
				return fmt.Errorf("failed to verify email: %w", ErrUserNotFound)
			}
			return fmt.Errorf("failed to verify email: %w", err)
		}

		if changed {
			err = s.outbox.WriteUserCreationEvent(ctx, tx, domain.UserCreationEvent{
				ID:        uuid.New(),
				UserID:    t.UserID,
				CreatedAt: time.Now(),
			})
			if err != nil {
				return err
			}
			s.logger.Debug("email verified", slog.Any("user", t.UserID))

		} else {
			s.logger.Debug("email already verified", slog.Any("user", t.UserID))
		}

		if err = s.tokens.MarkUsed(ctx, tx, tokenHash); err != nil {
			if errors.Is(err, email_token.ErrTokenNotFound) {
				s.logger.Warn("failed to mark used email_token as used")
			}
		}

		return nil
	})

	return nil
}

type LoginResult struct {
	AccessToken string
}

// Login checks the provided credentials and returns a JWT if they are valid.
//
// It returns the following domain errors:
//   - ErrInvalidCredentials
//   - ErrEmailNotVerified
//   - ErrIncorrectPassword
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) Login(ctx context.Context, email, password string) (uuid.UUID, error) {
	u, err := s.verifyCredentials(ctx, email, password)
	if err != nil {
		return uuid.Nil, err
	}

	if !u.EmailVerified {
		return uuid.Nil, ErrEmailNotVerified
	}

	return u.ID, nil
}

func (s *Service) verifyCredentials(ctx context.Context, email, password string) (domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if err := s.validateCredentials(email, password); err != nil {
		return domain.User{}, ErrInvalidCredentials
	}

	u, err := s.users.GetByEmail(ctx, s.db, email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return domain.User{}, fmt.Errorf("user not found: %w", ErrInvalidCredentials)
		}
		return domain.User{}, fmt.Errorf("failed to get user by email: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, ErrIncorrectPassword
	}

	return u, nil
}
