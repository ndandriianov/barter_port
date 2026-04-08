package domain

import "errors"

var (
	ErrChatNotFound      = errors.New("chat not found")
	ErrForbidden         = errors.New("forbidden")
	ErrChatAlreadyExists = errors.New("chat already exists")
)
