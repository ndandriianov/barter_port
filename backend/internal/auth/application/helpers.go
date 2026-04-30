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

func getHashFromRawTokenWithError(rawToken string, emptyErr error) (string, error) {
	trimmedToken := strings.TrimSpace(rawToken)
	if trimmedToken == "" {
		return "", emptyErr
	}
	return sha256Hex(rawToken), nil
}

func getHashFromRawToken(rawToken string) (string, error) {
	return getHashFromRawTokenWithError(rawToken, domain.ErrInvalidEmailToken)
}

func getHashFromToken(token string) string {
	return sha256Hex(token)
}

// --- CREDENTIAL VALIDATION ---

func (s *Service) validateCredentials(email, password string) error {
	if !s.re.MatchString(email) {
		return domain.ErrInvalidEmail
	}

	return s.validatePassword(password)
}

func (s *Service) validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return domain.ErrPasswordTooShort
	}

	return nil
}

func (s *Service) shouldAutoVerifyEmail(email string) bool {
	if s.emailBypassMode {
		return true
	}

	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	return strings.HasSuffix(normalizedEmail, "@barterport.local")
}

// --- EMAIL VERIFICATION ---

func (s *Service) getVerifyURL(token string) string {
	return s.frontendBaseURL + verifyEmailTokenURLPath + token
}

func (s *Service) getVerifyEmailBody(token string) string {
	body := "Hello!\n\n" +
		"Please confirm your email by clicking the link:\n\n" +
		s.getVerifyURL(token) + "\n\n" +
		"If you didn't register, ignore this email."
	return body
}

// --- PASSWORD RESET ---

func (s *Service) getPasswordResetURL(token string) string {
	return s.frontendBaseURL + passwordResetTokenURLPath + token
}

func (s *Service) getPasswordResetEmailBody(token string) string {
	body := "Hello!\n\n" +
		"To reset your password, click the link below:\n\n" +
		s.getPasswordResetURL(token) + "\n\n" +
		"If you didn't request a password reset, ignore this email."
	return body
}
