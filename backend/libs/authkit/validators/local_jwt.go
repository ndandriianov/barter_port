package validators

import (
	authkit2 "barter-port/libs/authkit"
	"barter-port/libs/jwt"
	"errors"

	"golang.org/x/net/context"
)

type LocalJWT struct {
	manager *jwt.Manager
}

func NewLocalJWT(manager *jwt.Manager) *LocalJWT {
	return &LocalJWT{manager: manager}
}

func (v *LocalJWT) ValidateAccess(_ context.Context, token string) (authkit2.Principal, error) {
	claims, err := v.manager.ParseAccessToken(token)
	if err != nil {
		return authkit2.Principal{}, mapJWTError(err)
	}

	p := authkit2.Principal{
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
		return errors.Join(authkit2.ErrTokenExpired, err)

	case errors.Is(err, jwt.ErrInvalidToken),
		errors.Is(err, jwt.ErrInvalidTokenType),
		errors.Is(err, jwt.ErrUnexpectedSigningMethod):
		return errors.Join(authkit2.ErrInvalidToken, err)

	default:
		return err
	}
}
