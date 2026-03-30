package domain

import "errors"

var (
	ErrInvalidEmail      = errors.New("invalid email")
	ErrPasswordTooShort  = errors.New("password too short")
	ErrEmailAlreadyInUse = errors.New("email already in use")

	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidEmailToken = errors.New("invalid email_token")
	ErrEmailTokenExpired = errors.New("email_token expired")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrIncorrectPassword  = errors.New("incorrect password")
)
