package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ndandriianov/barter_port/backend/internal/model"
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type JWTService struct {
	jwtSecret []byte
	jwtTTL    time.Duration
}

var (
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrAccessJWTExpired        = errors.New("token expired")
	ErrInvalidAccessJWT        = errors.New("invalid token")
)

func NewJWTService(secret []byte, ttl time.Duration) *JWTService {
	return &JWTService{
		jwtSecret: secret,
		jwtTTL:    ttl,
	}
}

func (s *JWTService) generateAccessToken(u model.User) (string, error) {
	now := time.Now()

	claims := Claims{
		Email: u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *JWTService) ParseToken(raw string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSigningMethod
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrAccessJWTExpired
		}
		return nil, ErrInvalidAccessJWT
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Subject == "" {
		return nil, ErrInvalidAccessJWT
	}

	return claims, nil
}
