package auth

import (
	"errors"
	"strings"
	"time"

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
	ErrInvalidEmail         = errors.New("invalid email")
	ErrPasswordTooShort     = errors.New("password too short")
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrEmailAlreadyVerified = errors.New("email already verified")
)

type UserRepo interface {
	Create(user model.User) error
	// GetByEmail(email string) (model.User, error)
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
}

func NewService(users UserRepo, tokens TokenRepo, mailer Mailer, frontendBaseURL string) *Service {
	return &Service{
		users:           users,
		tokens:          tokens,
		mailer:          mailer,
		frontendBaseURL: strings.TrimRight(frontendBaseURL, "/"),
	}
}

type RegisterResult struct {
	UserID string
	Email  string
}

// Register creates a new user, generates a verification token, and sends a verification email.
func (s *Service) Register(email, password string) (RegisterResult, error) {
	if err := validateCredentials(email, password); err != nil {
		return RegisterResult{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return RegisterResult{}, err
	}

	user := model.NewUser(newID(), email, string(hash))
	if err := s.users.Create(user); err != nil {
		return RegisterResult{}, err
	}

	// на всякий случай удаляются все старые токены для этого юзера, если они были
	s.tokens.DeleteAllForUser(user.ID)

	rawToken, err := generateToken(tokenLength)
	if err != nil {
		return RegisterResult{}, err
	}

	tokenHash := sha256Hex(rawToken)
	t := model.NewEmailVerificationToken(tokenHash, user.ID, time.Now().Add(tokenExpirationTime))

	if err = s.tokens.Save(t); err != nil {
		return RegisterResult{}, err
	}

	if err = s.mailer.Send(user.Email, subject, s.getEmailBody(rawToken)); err != nil {
		return RegisterResult{}, err
	}

	return RegisterResult{
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}

// VerifyEmail verifies a user's email using a raw token.
func (s *Service) VerifyEmail(rawToken string) error {
	tokenHash, err := getHashFromRawToken(rawToken)
	if err != nil {
		return err
	}

	t, err := s.tokens.GetByHash(tokenHash)
	if err != nil {
		return ErrInvalidToken
	}

	if t.Used {
		return nil
	}
	if time.Now().After(t.ExpiresAt) {
		return ErrTokenExpired
	}

	u, err := s.users.GetByID(t.UserID)
	if err != nil {
		return err
	}

	if u.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	if err = s.users.VerifyEmail(u.ID); err != nil {
		return err
	}

	if err = s.tokens.MarkUsed(tokenHash); err != nil {
		return err
	}

	return nil
}
