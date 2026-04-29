package domain

import "errors"

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrUserAlreadyExists         = errors.New("user already exists")
	ErrForbidden                 = errors.New("forbidden")
	ErrAlreadySubscribed         = errors.New("already subscribed")
	ErrNotSubscribed             = errors.New("not subscribed")
	ErrCannotSubscribeToYourself = errors.New("cannot subscribe to yourself")
	ErrCannotHideYourself        = errors.New("cannot hide yourself")
	ErrCannotHideSubscribedUser  = errors.New("cannot hide subscribed user")
	ErrHiddenUserSubscription    = errors.New("cannot subscribe to hidden user")
	ErrInvalidPhoneNumber        = errors.New("phone number must be in format +7 (999) 123-45-67")
)
