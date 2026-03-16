package authkit

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ParseToken is a helper method to parse and validate a JWT token.
// It checks the signing method, validates the token, and ensures the token type matches the expected type.
//
// Errors:
//   - ErrUnexpectedSigningMethod: if the token's signing method is not HS256.
//   - ErrTokenExpired: if the token has expired.
//   - ErrInvalidToken: if the token is invalid for any reason (e.g., malformed, signature mismatch).
//   - ErrInvalidTokenType: if the token's type does not match the expected type.
func ParseToken(tokenStr string, expectedType TokenType, secret string) (*Claims, error) {
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
