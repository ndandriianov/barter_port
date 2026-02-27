package validators

import (
	"barter-port/internal/libs/authkit"
	"barter-port/internal/libs/jwt"
	"errors"

	"golang.org/x/net/context"
)

type LocalJWT struct {
	manager *jwt.Manager
}

func NewLocalJWT(manager *jwt.Manager) *LocalJWT {
	return &LocalJWT{manager: manager}
}

func (v *LocalJWT) ValidateAccess(_ context.Context, token string) (authkit.Principal, error) {
	claims, err := v.manager.ParseAccessToken(token)
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
		t := claims.IssuedAt.Time
		p.IssuedAt = &t
	}

	if claims.ExpiresAt != nil {
		t := claims.ExpiresAt.Time
		p.ExpiresAt = &t
	}

	return p, nil
}

func mapJWTError(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return errors.Join(authkit.ErrTokenExpired, err)

	case errors.Is(err, jwt.ErrInvalidToken),
		errors.Is(err, jwt.ErrInvalidTokenType),
		errors.Is(err, jwt.ErrUnexpectedSigningMethod):
		return errors.Join(authkit.ErrInvalidToken, err)

	default:
		return err
	}
}
