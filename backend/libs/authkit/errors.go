package authkit

import "errors"

var (
	ErrMissingToken        = errors.New("missing token")
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrInternalServerError = errors.New("internal server error")
	ErrUnavailable         = errors.New("auth unavailable") // На будущее для авторизации через auth сервис
)
