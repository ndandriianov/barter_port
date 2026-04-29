package domain

import "errors"

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrUserAlreadyExists         = errors.New("user already exists")
	ErrForbidden                 = errors.New("forbidden")
	ErrAlreadySubscribed         = errors.New("already subscribed")
	ErrNotSubscribed             = errors.New("not subscribed")
	ErrCannotSubscribeToYourself = errors.New("cannot subscribe to yourself")
	ErrInvalidPhoneNumber        = errors.New("phone number must be in format +7 (999) 123-45-67")
)
