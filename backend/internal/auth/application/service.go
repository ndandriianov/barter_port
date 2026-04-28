package application

import (
	authusers "barter-port/contracts/kafka/messages/auth-users"
	"barter-port/internal/auth/domain"
	ucstatus "barter-port/internal/auth/domain/uc-status"
	ucoutbox "barter-port/internal/auth/infrastructure/repository/uc-outbox"
	"barter-port/pkg/db"
	"context"
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
)

const (
	bcryptCost                   = 12
	minPasswordLength            = 6
	tokenLength                  = 32
	verifyEmailTokenExpiration   = 24 * time.Hour
	passwordResetTokenExpiration = 2 * time.Hour
	verifyEmailTokenURLPath      = "/verify-email?token="
	passwordResetTokenURLPath    = "/reset-password?token="
	verifyEmailSubject           = "Confirm your email"
	passwordResetSubject         = "Reset your password"
)

type UserRepo interface {
	Create(ctx context.Context, exec db.DB, user domain.User) error
	GetByEmail(ctx context.Context, exec db.DB, email string) (domain.User, error)
	GetByID(ctx context.Context, exec db.DB, id uuid.UUID) (domain.User, error)
	GetStatistics(ctx context.Context, exec db.DB) (totalRegistered int, verifiedEmails int, err error)
	VerifyEmailIfNotVerified(ctx context.Context, exec db.DB, userID uuid.UUID) (changed bool, err error)
	UpdatePasswordHash(ctx context.Context, exec db.DB, userID uuid.UUID, passwordHash string) error
}

type EmailTokenRepo interface {
	Save(ctx context.Context, exec db.DB, t domain.EmailVerificationToken) error
	GetByHashForUpdate(ctx context.Context, exec db.DB, tokenHash string) (domain.EmailVerificationToken, error)
	MarkUsed(ctx context.Context, exec db.DB, tokenHash string) error
	DeleteAllForUser(ctx context.Context, exec db.DB, userID uuid.UUID) error
}

type PasswordResetTokenRepo interface {
	Save(ctx context.Context, exec db.DB, t domain.PasswordResetToken) error
	GetByHashForUpdate(ctx context.Context, exec db.DB, tokenHash string) (domain.PasswordResetToken, error)
	MarkUsed(ctx context.Context, exec db.DB, tokenHash string) error
	DeleteAllForUser(ctx context.Context, exec db.DB, userID uuid.UUID) error
}

type RefreshTokenRepo interface {
	DeleteAllForUser(ctx context.Context, exec db.DB, userID uuid.UUID) error
}

type UserCreationEventRepository interface {
	Add(ctx context.Context, exec db.DB, event domain.UserCreationEvent) error
	GetByUserID(ctx context.Context, exec db.DB, userID uuid.UUID) (*domain.UserCreationEvent, error)
	SetStatus(ctx context.Context, exec db.DB, userID uuid.UUID, status string) error
}

type Mailer interface {
	Send(to, subject, body string) error
}

type Service struct {
	db              *pgxpool.Pool
	users           UserRepo
	ucEventRepo     UserCreationEventRepository
	emailTokens     EmailTokenRepo
	passwordTokens  PasswordResetTokenRepo
	refreshTokens   RefreshTokenRepo
	mailer          Mailer
	logger          *slog.Logger
	outbox          *ucoutbox.Repository
	emailBypassMode bool

	frontendBaseURL string
	adminEmail      string
	re              *regexp.Regexp
}

func NewService(
	db *pgxpool.Pool,
	users UserRepo,
	ucEventRepo UserCreationEventRepository,
	emailTokens EmailTokenRepo,
	passwordTokens PasswordResetTokenRepo,
	refreshTokens RefreshTokenRepo,
	mailer Mailer,
	logger *slog.Logger,
	outbox *ucoutbox.Repository,
	emailBypassMode bool,

	frontendBaseURL string,
	adminEmail string,
	re *regexp.Regexp,
) *Service {
	if logger == nil {
		log.Fatal("logger is required")
	}

	return &Service{
		db:              db,
		users:           users,
		ucEventRepo:     ucEventRepo,
		emailTokens:     emailTokens,
		passwordTokens:  passwordTokens,
		refreshTokens:   refreshTokens,
		mailer:          mailer,
		logger:          logger,
		outbox:          outbox,
		emailBypassMode: emailBypassMode,

		frontendBaseURL: strings.TrimRight(frontendBaseURL, "/"),
		adminEmail:      strings.TrimSpace(strings.ToLower(adminEmail)),
		re:              re,
	}
}

type RegisterResult struct {
	UserID uuid.UUID
	Email  string
}

const typeName = "auth.service"

// Register creates a new user, generates an email verification token,
// and sends a verification email.
//
// It returns the following domain errors:
//   - domain.ErrInvalidEmail
//   - domain.ErrPasswordTooShort
//   - domain.ErrEmailAlreadyInUse
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) Register(ctx context.Context, email, password string) (RegisterResult, error) {
	const funcName = "Register"

	log := s.logger.With(slog.String("func", funcName), slog.String("type", typeName))

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
			if errors.Is(err, domain.ErrEmailAlreadyInUse) {
				return domain.ErrEmailAlreadyInUse
			}
			return fmt.Errorf("failed to create user: %w", err)
		}

		if s.emailBypassMode {
			if err = s.createUser(ctx, tx, u.ID); err != nil {
				return err
			}

			if _, err = s.users.VerifyEmailIfNotVerified(ctx, tx, u.ID); err != nil {
				return fmt.Errorf("failed to verify email: %w", err)
			}

			log.Info("user created", slog.String("email", u.Email), slog.Any("user_id", u.ID))
			return nil
		}

		rawToken, err = generateToken(tokenLength)
		if err != nil {
			return fmt.Errorf("failed to generate email_token: %w", err)
		}

		tokenHash := getHashFromToken(rawToken)
		t := domain.NewEmailVerificationToken(tokenHash, u.ID, time.Now().Add(verifyEmailTokenExpiration))

		if err = s.emailTokens.Save(ctx, tx, t); err != nil {
			return fmt.Errorf("failed to save email_token: %w", err)
		}

		return nil
	})
	if err != nil {
		return RegisterResult{}, err
	}

	if err = s.mailer.Send(u.Email, verifyEmailSubject, s.getVerifyEmailBody(rawToken)); err != nil {
		log.Warn("failed to send email", slog.Any("error", err))
	}

	return RegisterResult{
		UserID: u.ID,
		Email:  u.Email,
	}, nil
}

func (s *Service) CreateAdmin(ctx context.Context, email, password string) (RegisterResult, error) {
	const funcName = "CreateAdmin"

	log := s.logger.With(slog.String("func", funcName), slog.String("type", typeName))

	email = strings.TrimSpace(strings.ToLower(email))

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("failed to hash password: %w", err)
	}

	u := domain.NewUser(uuid.New(), email, string(hash))
	u.EmailVerified = true
	result := RegisterResult{UserID: u.ID, Email: u.Email}

	err = db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		err = s.users.Create(ctx, tx, u)
		if err != nil {
			if !errors.Is(err, domain.ErrEmailAlreadyInUse) {
				return err
			}

			return fmt.Errorf("failed to create user: %w", err)
		}

		if err = s.createUser(ctx, tx, u.ID); err != nil {
			return fmt.Errorf("failed to create admin user event: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyInUse) {
			log.Warn("admin already exists", slog.String("email", email))
			return result, nil
		}

		return RegisterResult{}, err
	}

	log.Info("admin ensured", slog.String("email", result.Email), slog.Any("user_id", result.UserID))

	return result, nil
}

// RetrySendVerificationEmail generates a new email verification token and sends a verification email if the user's email is not verified.
//
// It returns the following domain errors:
//   - domain.ErrInvalidCredentials
//   - domain.ErrEmailNotVerified
//   - domain.ErrIncorrectPassword
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
	t := domain.NewEmailVerificationToken(tokenHash, u.ID, time.Now().Add(verifyEmailTokenExpiration))

	return db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		if err = s.emailTokens.Save(ctx, tx, t); err != nil {
			return fmt.Errorf("failed to save email_token: %w", err)
		}

		if err = s.mailer.Send(u.Email, verifyEmailSubject, s.getVerifyEmailBody(rawToken)); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		return nil
	})
}

// VerifyEmail marks user's email as verified if the provided token is valid.
//
// It returns the following domain errors:
//   - domain.ErrInvalidEmailToken
//   - domain.ErrEmailTokenExpired
//   - domain.ErrUserNotFound
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	const funcName = "VerifyEmail"

	log := s.logger.With(slog.String("func", funcName), slog.String("type", typeName))

	tokenHash, err := getHashFromRawToken(rawToken)
	if err != nil {
		return fmt.Errorf("cannot get hash from raw email_token: %w", err)
	}

	err = db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		t, err := s.emailTokens.GetByHashForUpdate(ctx, tx, tokenHash)
		if err != nil {
			if errors.Is(err, domain.ErrTokenNotFound) {
				return domain.ErrInvalidEmailToken
			}
			return fmt.Errorf("failed to get email_token by hash: %w", err)
		}

		if t.Used {
			return nil
		}
		if time.Now().After(t.ExpiresAt) {
			return domain.ErrEmailTokenExpired
		}

		changed, err := s.users.VerifyEmailIfNotVerified(ctx, tx, t.UserID)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				return fmt.Errorf("failed to verify email: %w", domain.ErrUserNotFound)
			}
			return fmt.Errorf("failed to verify email: %w", err)
		}

		if changed {
			if err = s.createUser(ctx, tx, t.UserID); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
			log.Debug("email verified", slog.Any("user", t.UserID))

		} else {
			log.Debug("email already verified", slog.Any("user", t.UserID))
		}

		if err = s.emailTokens.MarkUsed(ctx, tx, tokenHash); err != nil {
			return fmt.Errorf("failed to mark used email_token as used: %w", err)
		}

		return nil
	})

	return err
}

// GetMe retrieves the user's information by their ID.
//
// Errors:
//   - domain.ErrUserNotFound
func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	return s.users.GetByID(ctx, s.db, userID)
}

// IsAdmin reports whether the user matches the configured admin account.
//
// Errors:
//   - domain.ErrUserNotFound
func (s *Service) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := s.GetMe(ctx, userID)
	if err != nil {
		return false, err
	}

	return s.isAdminEmail(user.Email), nil
}

func (s *Service) IsAdminEmail(email string) bool {
	return s.isAdminEmail(email)
}

func (s *Service) GetAdminPlatformStatistics(ctx context.Context, requesterID uuid.UUID) (AdminPlatformStatistics, error) {
	isAdmin, err := s.IsAdmin(ctx, requesterID)
	if err != nil {
		return AdminPlatformStatistics{}, err
	}
	if !isAdmin {
		return AdminPlatformStatistics{}, domain.ErrForbidden
	}

	totalRegistered, verifiedEmails, err := s.users.GetStatistics(ctx, s.db)
	if err != nil {
		return AdminPlatformStatistics{}, fmt.Errorf("failed to get auth platform statistics: %w", err)
	}

	return AdminPlatformStatistics{
		Users: AdminPlatformUsersStatistics{
			TotalRegistered: totalRegistered,
			VerifiedEmails:  verifiedEmails,
		},
	}, nil
}

func (s *Service) GetAdminUserStatistics(
	ctx context.Context,
	requesterID uuid.UUID,
	targetUserID uuid.UUID,
) (AdminUserStatistics, error) {
	isAdmin, err := s.IsAdmin(ctx, requesterID)
	if err != nil {
		return AdminUserStatistics{}, err
	}
	if !isAdmin {
		return AdminUserStatistics{}, domain.ErrForbidden
	}

	user, err := s.users.GetByID(ctx, s.db, targetUserID)
	if err != nil {
		return AdminUserStatistics{}, err
	}

	return AdminUserStatistics{
		UserID:        targetUserID,
		RegisteredAt:  user.CreatedAt,
		EmailVerified: user.EmailVerified,
	}, nil
}

func (s *Service) isAdminEmail(email string) bool {
	if s.adminEmail == "" {
		return false
	}

	return strings.TrimSpace(strings.ToLower(email)) == s.adminEmail
}

func (s *Service) createUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	const funcName = "createUser"

	log := s.logger.With(slog.String("func", funcName), slog.String("type", typeName))

	event := domain.UserCreationEvent{
		UserID:    userID,
		CreatedAt: time.Now(),
		Status:    ucstatus.New,
	}

	err := s.ucEventRepo.Add(ctx, tx, event)
	if err != nil {
		return fmt.Errorf("failed to add user creation: %w", err)
	}

	err = s.outbox.WriteUserCreationMessage(ctx, tx, authusers.UserCreationMessage{
		ID:        uuid.New(),
		UserID:    event.UserID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to write user creation message to outbox: %w", err)
	}

	log.Info("user created", slog.Any("user_id", event.UserID))

	return nil
}

type LoginResult struct {
	AccessToken string
}

type AdminPlatformStatistics struct {
	Users AdminPlatformUsersStatistics
}

type AdminPlatformUsersStatistics struct {
	TotalRegistered int
	VerifiedEmails  int
}

type AdminUserStatistics struct {
	UserID        uuid.UUID
	RegisteredAt  time.Time
	EmailVerified bool
}

func (s *Service) GetUserCreationStatus(ctx context.Context, userID uuid.UUID) (string, error) {
	event, err := s.ucEventRepo.GetByUserID(ctx, s.db, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user creation event: %w", err)
	}

	return event.Status.String(), nil
}

// Login checks the provided credentials and returns a JWT if they are valid.
//
// It returns the following domain errors:
//   - domain.ErrInvalidCredentials
//   - domain.ErrEmailNotVerified
//   - domain.ErrIncorrectPassword
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) Login(ctx context.Context, email, password string) (uuid.UUID, error) {
	u, err := s.verifyCredentials(ctx, email, password)
	if err != nil {
		return uuid.Nil, err
	}

	if !u.EmailVerified {
		return uuid.Nil, domain.ErrEmailNotVerified
	}

	return u.ID, nil
}

// RequestPasswordReset generates a password reset token and sends a recovery link to the user's email.
//
// It returns the following domain errors:
//   - domain.ErrInvalidEmail
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if !s.re.MatchString(email) {
		return domain.ErrInvalidEmail
	}

	u, err := s.users.GetByEmail(ctx, s.db, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get user by email: %w", err)
	}

	rawToken, err := generateToken(tokenLength)
	if err != nil {
		return fmt.Errorf("failed to generate password reset token: %w", err)
	}

	tokenHash := getHashFromToken(rawToken)
	t := domain.NewPasswordResetToken(tokenHash, u.ID, time.Now().Add(passwordResetTokenExpiration))

	err = db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.passwordTokens.DeleteAllForUser(ctx, tx, u.ID); err != nil {
			return fmt.Errorf("failed to delete previous password reset tokens: %w", err)
		}

		if err := s.passwordTokens.Save(ctx, tx, t); err != nil {
			return fmt.Errorf("failed to save password reset token: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err := s.mailer.Send(u.Email, passwordResetSubject, s.getPasswordResetEmailBody(rawToken)); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// ResetPassword validates a password reset token and replaces the user's password.
//
// It returns the following domain errors:
//   - domain.ErrPasswordTooShort
//   - domain.ErrInvalidPasswordResetToken
//   - domain.ErrPasswordResetTokenExpired
//   - domain.ErrUserNotFound
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	tokenHash, err := getHashFromRawTokenWithError(rawToken, domain.ErrInvalidPasswordResetToken)
	if err != nil {
		return fmt.Errorf("cannot get hash from raw password reset token: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		t, err := s.passwordTokens.GetByHashForUpdate(ctx, tx, tokenHash)
		if err != nil {
			if errors.Is(err, domain.ErrTokenNotFound) {
				return domain.ErrInvalidPasswordResetToken
			}
			return fmt.Errorf("failed to get password reset token by hash: %w", err)
		}

		if t.Used {
			return domain.ErrInvalidPasswordResetToken
		}
		if time.Now().After(t.ExpiresAt) {
			return domain.ErrPasswordResetTokenExpired
		}

		if err := s.users.UpdatePasswordHash(ctx, tx, t.UserID, string(hash)); err != nil {
			return fmt.Errorf("failed to update password hash: %w", err)
		}

		if err := s.refreshTokens.DeleteAllForUser(ctx, tx, t.UserID); err != nil {
			return fmt.Errorf("failed to delete refresh tokens for user: %w", err)
		}

		if err := s.passwordTokens.MarkUsed(ctx, tx, tokenHash); err != nil {
			return fmt.Errorf("failed to mark password reset token as used: %w", err)
		}

		return nil
	})
}

// ChangePassword updates the authenticated user's password after re-checking the old credentials.
//
// It returns the following domain errors:
//   - domain.ErrInvalidOldCredentials
//   - domain.ErrPasswordTooShort
//   - domain.ErrUserNotFound
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, oldEmail, oldPassword, newPassword string) error {
	oldEmail = strings.TrimSpace(strings.ToLower(oldEmail))

	if oldEmail == "" || strings.TrimSpace(oldPassword) == "" {
		return domain.ErrInvalidOldCredentials
	}

	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	u, err := s.verifyCredentials(ctx, oldEmail, oldPassword)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrIncorrectPassword):
			return domain.ErrInvalidOldCredentials
		default:
			return fmt.Errorf("failed to verify old credentials: %w", err)
		}
	}

	if u.ID != userID {
		return domain.ErrInvalidOldCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.users.UpdatePasswordHash(ctx, s.db, userID, string(hash)); err != nil {
		return fmt.Errorf("failed to update password hash: %w", err)
	}

	if err := s.refreshTokens.DeleteAllForUser(ctx, s.db, userID); err != nil {
		return fmt.Errorf("failed to delete refresh tokens for user: %w", err)
	}

	return nil
}

func (s *Service) verifyCredentials(ctx context.Context, email, password string) (domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if err := s.validateCredentials(email, password); err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}

	u, err := s.users.GetByEmail(ctx, s.db, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.User{}, fmt.Errorf("user not found: %w", domain.ErrInvalidCredentials)
		}
		return domain.User{}, fmt.Errorf("failed to get user by email: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, domain.ErrIncorrectPassword
	}

	return u, nil
}
