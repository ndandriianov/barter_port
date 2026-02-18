package errors

import "errors"

var (
	// User Repo errors
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyInUse = errors.New("email already in use")

	// Token Repo errors
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenAlreadyExists = errors.New("token already exists")

	// Auth Service errors
	ErrInvalidEmail         = errors.New("invalid email")
	ErrPasswordTooShort     = errors.New("password too short")
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrTokenAlreadyUsed     = errors.New("token already used")
	ErrEmailAlreadyVerified = errors.New("email already verified")
)
