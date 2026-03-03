package jwt

import (
	"barter-port/internal/libs/authkit"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	GTILength int = 16
)

type Config struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type Manager struct {
	cfg Config
}

func NewManager(cfg Config) *Manager {
	return &Manager{cfg: cfg}
}

//
// === GENERATE TOKENS ===
//

// GenerateAccessToken generates a signed JWT access token for the given user ID.
//
// Errors: it returns only internal errors related to token generation and signing,
// as the input parameters are controlled by the application logic.
func (m *Manager) GenerateAccessToken(userID uuid.UUID) (string, error) {
	token, _, err := m.generateToken(userID, authkit.AccessToken, m.cfg.AccessSecret, m.cfg.AccessTTL)
	return token, err
}

// GenerateRefreshToken generates a signed JWT refresh token for the given user ID
// and returns both the token string and its authkit.Claims.
//
// Errors: it returns only internal errors related to token generation and signing,
// as the input parameters are controlled by the application logic.
func (m *Manager) GenerateRefreshToken(userID uuid.UUID) (string, authkit.Claims, error) {
	return m.generateToken(userID, authkit.RefreshToken, m.cfg.RefreshSecret, m.cfg.RefreshTTL)
}

// generateToken is a helper method to create a JWT token with the specified user ID,
// token type, secret, and time-to-live (TTL).
// It generates a unique JTI, sets the issued and expiration times, and signs the token using HS256.
//
// Errors: it returns only internal errors related to token generation and signing,
// as the input parameters are controlled by the application logic.
func (m *Manager) generateToken(
	userID uuid.UUID,
	tokenType authkit.TokenType,
	secret string,
	ttl time.Duration,
) (string, authkit.Claims, error) {
	id, err := generateJTI()
	if err != nil {
		return "", authkit.Claims{}, fmt.Errorf("failed to generate GTI: %w", err)
	}
	now := time.Now()

	Claims := authkit.Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        id,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", authkit.Claims{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, Claims, nil
}

//
// === PARSE TOKENS ===
//

// ParseAccessToken parses and validates an access token string, returning the authkit.Claims if valid.
//
// Errors:
//   - authkit.ErrUnexpectedSigningMethod: if the token's signing method is not HS256.
//   - authkit.ErrTokenExpired: if the token has expired.
//   - authkit.ErrInvalidToken: if the token is invalid for any reason (e.g., malformed, signature mismatch).
//   - authkit.ErrInvalidTokenType: if the token's type does not match "access".
func (m *Manager) ParseAccessToken(tokenStr string) (*authkit.Claims, error) {
	return authkit.ParseToken(tokenStr, authkit.AccessToken, m.cfg.AccessSecret)
}

// ParseRefreshToken parses and validates a refresh token string, returning the authkit.Claims if valid.
//
// Errors:
//   - authkit.ErrUnexpectedSigningMethod: if the token's signing method is not HS256.
//   - authkit.ErrTokenExpired: if the token has expired.
//   - authkit.ErrInvalidToken: if the token is invalid for any reason (e.g., malformed, signature mismatch).
//   - authkit.ErrInvalidTokenType: if the token's type does not match "refresh".
func (m *Manager) ParseRefreshToken(tokenStr string) (*authkit.Claims, error) {
	return authkit.ParseToken(tokenStr, authkit.RefreshToken, m.cfg.RefreshSecret)
}

//
// === HELPERS ===
//

func generateJTI() (string, error) {
	b := make([]byte, GTILength)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
