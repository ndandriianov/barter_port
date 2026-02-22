package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
	GTILength    int       = 16
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
	UserID string    `json:"user_id"`
	Type   TokenType `json:"type"`
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

func (m *Manager) GenerateAccessToken(userID string) (string, error) {
	token, _, err := m.generateToken(userID, AccessToken, m.cfg.AccessSecret, m.cfg.AccessTTL)
	return token, err
}

func (m *Manager) GenerateRefreshToken(userID string) (string, Claims, error) {
	return m.generateToken(userID, RefreshToken, m.cfg.RefreshSecret, m.cfg.RefreshTTL)
}

func (m *Manager) generateToken(userID string, tokenType TokenType, secret string, ttl time.Duration) (string, Claims, error) {
	id, err := generateGTI()
	if err != nil {
		return "", Claims{}, fmt.Errorf("failed to generate GTI: %w", err)
	}
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
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

func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	return m.parseToken(tokenStr, AccessToken, m.cfg.AccessSecret)
}

func (m *Manager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	return m.parseToken(tokenStr, RefreshToken, m.cfg.RefreshSecret)
}

func (m *Manager) parseToken(tokenStr string, expectedType TokenType, secret string) (*Claims, error) {
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

func generateGTI() (string, error) {
	b := make([]byte, GTILength)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
