package auth

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/repository/token"
	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/repository/user"
	"github.com/ndandriianov/barter_port/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
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

	ErrUserNotFound = errors.New("user not found")
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
)

type UserRepo interface {
	Create(user model.User) error
	GetByEmail(email string) (model.User, error)
	GetByID(id string) (model.User, error)
	VerifyEmail(userID string) error
}

type TokenRepo interface {
	Save(t model.EmailVerificationToken) error
	GetByHash(tokenHash string) (model.EmailVerificationToken, error)
	MarkUsed(tokenHash string) error
	DeleteAllForUser(userID string)
}

type Mailer interface {
	Send(to, subject, body string) error
}

type Service struct {
	users  UserRepo
	tokens TokenRepo
	mailer Mailer

	frontendBaseURL string

	jwtSecret []byte
	jwtTTL    time.Duration

	re *regexp.Regexp
}

func NewService(
	users UserRepo,
	tokens TokenRepo,
	mailer Mailer,
	frontendBaseURL string,
	jwtSecret string,
	jwtTTL time.Duration,
	re *regexp.Regexp,
) *Service {
	return &Service{
		users:  users,
		tokens: tokens,
		mailer: mailer,

		frontendBaseURL: strings.TrimRight(frontendBaseURL, "/"),

		jwtSecret: []byte(jwtSecret),
		jwtTTL:    jwtTTL,

		re: re,
	}
}

type RegisterResult struct {
	UserID string
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
func (s *Service) Register(email, password string) (RegisterResult, error) {
	if err := s.validateCredentials(email, password); err != nil {
		return RegisterResult{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12) // TODO: заменить на константу
	if err != nil {
		return RegisterResult{}, fmt.Errorf("failed to hash password: %w", err)
	}

	u := model.NewUser(newID(), email, string(hash))
	if err := s.users.Create(u); err != nil {
		if errors.Is(err, user.ErrEmailAlreadyInUse) {
			return RegisterResult{}, ErrEmailAlreadyInUse
		}
		return RegisterResult{}, fmt.Errorf("failed to create user: %w", err)
	}

	// на всякий случай удаляются все старые токены для этого юзера, если они были
	s.tokens.DeleteAllForUser(u.ID)

	rawToken, err := generateToken(tokenLength)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash := getHashFromToken(rawToken)
	t := model.NewEmailVerificationToken(tokenHash, u.ID, time.Now().Add(tokenExpirationTime))

	if err = s.tokens.Save(t); err != nil {
		return RegisterResult{}, fmt.Errorf("failed to save token: %w", err)
	}

	if err = s.mailer.Send(u.Email, subject, s.getEmailBody(rawToken)); err != nil {
		return RegisterResult{}, fmt.Errorf("failed to send email: %w", err)
	}

	return RegisterResult{
		UserID: u.ID,
		Email:  u.Email,
	}, nil
}

// VerifyEmail marks user's email as verified if the provided token is valid.
//
// It returns the following domain errors:
//   - ErrInvalidToken
//   - ErrTokenExpired
//   - ErrUserNotFound
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) VerifyEmail(rawToken string) error {
	tokenHash, err := getHashFromRawToken(rawToken)
	if err != nil {
		return fmt.Errorf("cannot get hash from raw token: %w", err)
	}

	t, err := s.tokens.GetByHash(tokenHash)
	if err != nil {
		if errors.Is(err, token.ErrTokenNotFound) {
			return ErrInvalidToken
		}
		return fmt.Errorf("failed to get token by hash: %w", err)
	}

	if t.Used {
		return nil
	}
	if time.Now().After(t.ExpiresAt) {
		return ErrTokenExpired
	}

	u, err := s.users.GetByID(t.UserID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user by id: %w", err)
	}

	if u.EmailVerified {
		return nil
	}

	if err = s.users.VerifyEmail(u.ID); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return fmt.Errorf("failed to verify email: %w", ErrUserNotFound)
		}
		return fmt.Errorf("failed to verify email: %w", err)
	}

	if err = s.tokens.MarkUsed(tokenHash); err != nil {
		if errors.Is(err, token.ErrTokenNotFound) {
			// TODO: логировать эту ошибку, но не возвращать её пользователю, так как верификация уже прошла успешно
		}
	}

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
//
// All other errors are treated as internal and returned wrapped.
func (s *Service) Login(email, password string) (LoginResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if !s.validateEmail(email) {
		return LoginResult{}, fmt.Errorf("invalid email: %w", ErrInvalidCredentials)
	}
	if !validatePassword(password) {
		return LoginResult{}, fmt.Errorf("invalid password: %w", ErrInvalidCredentials)
	}

	u, err := s.users.GetByEmail(email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return LoginResult{}, fmt.Errorf("user not found: %w", ErrInvalidCredentials)
		}
		return LoginResult{}, fmt.Errorf("failed to get user by email: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return LoginResult{}, fmt.Errorf("incorrect password: %w", err)
	}

	if !u.EmailVerified {
		return LoginResult{}, ErrEmailNotVerified
	}

	jwtToken, err := s.generateJWT(u)
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate jwt: %w", err)
	}

	return LoginResult{AccessToken: jwtToken}, nil
}
