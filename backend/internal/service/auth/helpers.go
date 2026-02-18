package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/ndandriianov/barter_port/backend/internal/errors"
)

func generateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func newID() string {
	// максимально простая заглушка
	raw, _ := generateToken(bcryptCost)
	return raw
}

func validateEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	return strings.Contains(email, "@") && len(email) >= 5
}

func validateCredentials(email, password string) error {
	if !validateEmail(email) {
		return errors.ErrInvalidEmail
	}
	if len(password) < minPasswordLength {
		return errors.ErrPasswordTooShort
	}
	return nil
}

func (s *Service) getVerifyURL(token string) string {
	return s.frontendBaseURL + tokenUrlPath + token
}

func (s *Service) getEmailBody(token string) string {
	body := "Hello!\n\n" +
		"Please confirm your email by clicking the link:\n\n" +
		s.getVerifyURL(token) + "\n\n" +
		"If you didn't register, ignore this email."
	return body
}
