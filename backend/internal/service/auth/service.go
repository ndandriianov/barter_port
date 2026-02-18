package auth

import (
	"strings"
	"time"

	"github.com/ndandriianov/barter_port/backend/internal/errors"
	"github.com/ndandriianov/barter_port/backend/internal/model"

	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	Create(user model.User) error
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

type RegisterResult struct {
	UserID string
	Email  string
}

type Service struct {
	users  UserRepo
	tokens TokenRepo
	mailer Mailer

	FrontendBaseURL string
}

func NewService(users UserRepo, tokens TokenRepo, mailer Mailer, frontendBaseURL string) *Service {
	return &Service{
		users:           users,
		tokens:          tokens,
		mailer:          mailer,
		FrontendBaseURL: strings.TrimRight(frontendBaseURL, "/"),
	}
}

func (s *Service) Register(email, password string) (RegisterResult, error) {
	if !validateEmail(email) {
		return RegisterResult{}, errors.ErrInvalidEmail
	}
	if len(password) < 6 {
		return RegisterResult{}, errors.ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return RegisterResult{}, err
	}

	user := model.User{
		ID:            newID(),
		Email:         email,
		PasswordHash:  string(hash),
		EmailVerified: false,
		CreatedAt:     time.Now(),
	}

	if err := s.users.Create(user); err != nil {
		return RegisterResult{}, err
	}

	// на всякий случай удалим все старые токены для этого юзера, если они были
	s.tokens.DeleteAllForUser(user.ID)

	rawToken, err := generateToken(32)
	if err != nil {
		return RegisterResult{}, err
	}

	tokenHash := sha256Hex(rawToken)

	t := model.EmailVerificationToken{
		TokenHash: tokenHash,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}

	if err := s.tokens.Save(t); err != nil {
		return RegisterResult{}, err
	}

	verifyURL := s.FrontendBaseURL + "/verify-email?token=" + rawToken

	subject := "Confirm your email"
	body := "Hello!\n\nPlease confirm your email by clicking the link:\n\n" + verifyURL + "\n\nIf you didn't register, ignore this email."

	if err := s.mailer.Send(user.Email, subject, body); err != nil {
		return RegisterResult{}, err
	}

	return RegisterResult{
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}

func (s *Service) VerifyEmail(rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return errors.ErrInvalidToken
	}

	tokenHash := sha256Hex(rawToken)

	t, err := s.tokens.GetByHash(tokenHash)
	if err != nil {
		return errors.ErrInvalidToken
	}

	if t.Used {
		return nil
	}
	if time.Now().After(t.ExpiresAt) {
		return errors.ErrTokenExpired
	}

	u, err := s.users.GetByID(t.UserID)
	if err != nil {
		return err
	}

	if u.EmailVerified {
		return errors.ErrEmailAlreadyVerified
	}

	if err = s.users.VerifyEmail(u.ID); err != nil {
		return err
	}

	if err = s.tokens.MarkUsed(tokenHash); err != nil {
		return err
	}

	return nil
}
