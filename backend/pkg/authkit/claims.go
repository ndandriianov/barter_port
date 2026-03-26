package authkit

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}
