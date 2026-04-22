package validators

import (
	"barter-port/pkg/authkit"
	"errors"
	"fmt"

	"golang.org/x/net/context"
)

type LocalJWT struct {
	secret string
}

func NewLocalJWT(secret string) *LocalJWT {
	return &LocalJWT{secret: secret}
}

// ValidateAccess validates the access token and returns the principal if valid.
//
// Errors: it returns
//   - authkit.ErrTokenExpired if the token is expired
//   - authkit.ErrInvalidToken if the token is invalid for any reason, also wrapping the original error for more context
func (v *LocalJWT) ValidateAccess(_ context.Context, token string) (authkit.Principal, error) {
	claims, err := authkit.ParseToken(token, authkit.AccessToken, v.secret)
	if err != nil {
		return authkit.Principal{}, mapJWTError(err)
	}

	p := authkit.Principal{
		UserID:  claims.UserID,
		Type:    claims.Type,
		Subject: claims.Subject,
		JTI:     claims.ID,
	}

	if claims.IssuedAt != nil {
		p.IssuedAt = new(claims.IssuedAt.Time)
	}

	if claims.ExpiresAt != nil {
		p.ExpiresAt = new(claims.ExpiresAt.Time)
	}

	return p, nil
}

func mapJWTError(err error) error {
	switch {
	case errors.Is(err, authkit.ErrTokenExpired):
		return err

	case errors.Is(err, authkit.ErrInvalidToken),
		errors.Is(err, authkit.ErrInvalidTokenType),
		errors.Is(err, authkit.ErrUnexpectedSigningMethod):
		return fmt.Errorf("%w: %v", authkit.ErrInvalidToken, err)

	default:
		return err
	}
}
