package jwt

import (
	"barter-port/libs/authkit"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	GTILength int = 16
)

var (
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid jwt")
	ErrInvalidTokenType        = errors.New("invalid jwt type")
	ErrTokenExpired            = errors.New("token expired")
)

type Config struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type Claims struct {
	UserID uuid.UUID         `json:"user_id"`
	Type   authkit.TokenType `json:"type"`
	jwt.RegisteredClaims
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
// and returns both the token string and its claims.
//
// Errors: it returns only internal errors related to token generation and signing,
// as the input parameters are controlled by the application logic.
func (m *Manager) GenerateRefreshToken(userID uuid.UUID) (string, Claims, error) {
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
) (string, Claims, error) {
	id, err := generateJTI()
	if err != nil {
		return "", Claims{}, fmt.Errorf("failed to generate GTI: %w", err)
	}
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        id,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", Claims{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, claims, nil
}

//
// === PARSE TOKENS ===
//

// ParseAccessToken parses and validates an access token string, returning the claims if valid.
//
// Errors:
//   - ErrUnexpectedSigningMethod: if the token's signing method is not HS256.
//   - ErrTokenExpired: if the token has expired.
//   - ErrInvalidToken: if the token is invalid for any reason (e.g., malformed, signature mismatch).
//   - ErrInvalidTokenType: if the token's type does not match "access".
func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	return m.parseToken(tokenStr, authkit.AccessToken, m.cfg.AccessSecret)
}

// ParseRefreshToken parses and validates a refresh token string, returning the claims if valid.
//
// Errors:
//   - ErrUnexpectedSigningMethod: if the token's signing method is not HS256.
//   - ErrTokenExpired: if the token has expired.
//   - ErrInvalidToken: if the token is invalid for any reason (e.g., malformed, signature mismatch).
//   - ErrInvalidTokenType: if the token's type does not match "refresh".
func (m *Manager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	return m.parseToken(tokenStr, authkit.RefreshToken, m.cfg.RefreshSecret)
}

// parseToken is a helper method to parse and validate a JWT token.
// It checks the signing method, validates the token, and ensures the token type matches the expected type.
//
// Errors:
//   - ErrUnexpectedSigningMethod: if the token's signing method is not HS256.
//   - ErrTokenExpired: if the token has expired.
//   - ErrInvalidToken: if the token is invalid for any reason (e.g., malformed, signature mismatch).
//   - ErrInvalidTokenType: if the token's type does not match the expected type.
func (m *Manager) parseToken(tokenStr string, expectedType authkit.TokenType, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrUnexpectedSigningMethod
			}
			return []byte(secret), nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Type != expectedType {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
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
