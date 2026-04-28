package domain

import "errors"

var (
	ErrInvalidEmail              = errors.New("invalid email")
	ErrPasswordTooShort          = errors.New("password too short")
	ErrEmailAlreadyInUse         = errors.New("email already in use")
	ErrInvalidOldCredentials     = errors.New("old credentials are invalid")
	ErrInvalidPasswordResetToken = errors.New("invalid password reset token")
	ErrPasswordResetTokenExpired = errors.New("password reset token expired")

	ErrInvalidEmailToken  = errors.New("invalid email_token")
	ErrEmailTokenExpired  = errors.New("email_token expired")
	ErrTokenAlreadyExists = errors.New("token already exists")
	ErrTokenNotFound      = errors.New("token not found")

	ErrUserNotFound       = errors.New("user not found")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrIncorrectPassword  = errors.New("incorrect password")

	ErrRefreshNotFound      = errors.New("refresh token not found")
	ErrRefreshAlreadyExists = errors.New("refresh token with this JTI already exists")
)
