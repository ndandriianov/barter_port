package application

import (
	"barter-port/internal/auth/domain"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// --- TOKEN HELPERS ---

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

func getHashFromRawToken(rawToken string) (string, error) {
	trimmedToken := strings.TrimSpace(rawToken)
	if trimmedToken == "" {
		return "", domain.ErrInvalidEmailToken
	}
	return sha256Hex(rawToken), nil
}

func getHashFromToken(token string) string {
	return sha256Hex(token)
}

// --- CREDENTIAL VALIDATION ---

func (s *Service) validateCredentials(email, password string) error {
	if !s.re.MatchString(email) {
		return domain.ErrInvalidEmail
	}
	if len(password) < minPasswordLength {
		return domain.ErrPasswordTooShort
	}
	return nil
}

// --- EMAIL VERIFICATION ---

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
