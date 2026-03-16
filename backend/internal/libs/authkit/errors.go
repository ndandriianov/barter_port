package authkit

import "errors"

var (
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrMissingToken            = errors.New("missing token")
	ErrInvalidTokenType        = errors.New("invalid jwt type")
	ErrInvalidToken            = errors.New("invalid token")
	ErrTokenExpired            = errors.New("token expired")
	ErrUnavailable             = errors.New("auth unavailable") // На будущее для авторизации через auth сервис
)
