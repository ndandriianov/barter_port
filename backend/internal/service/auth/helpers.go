package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
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
	raw, _ := generateToken(12)
	return raw
}

func validateEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	return strings.Contains(email, "@") && len(email) >= 5
}
