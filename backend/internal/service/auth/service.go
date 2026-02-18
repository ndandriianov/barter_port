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
}

type TokenRepo interface {
	Save(t model.EmailVerificationToken) error
	DeleteAllForUser(userID string)
}

type RegisterResult struct {
	UserID    string
	Email     string
	VerifyURL string
}

type Service struct {
	users  UserRepo
	tokens TokenRepo

	FrontendBaseURL string
}

func NewService(users UserRepo, tokens TokenRepo, frontendBaseURL string) *Service {
	return &Service{
		users:           users,
		tokens:          tokens,
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

	// В реальном проекте тут отправка письма.
	// Пока просто вернём ссылку, чтобы можно было тестить.
	return RegisterResult{
		UserID:    user.ID,
		Email:     user.Email,
		VerifyURL: verifyURL,
	}, nil
}
