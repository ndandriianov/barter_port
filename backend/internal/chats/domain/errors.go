package domain

import "errors"

var (
	ErrChatNotFound             = errors.New("chat not found")
	ErrForbidden                = errors.New("forbidden")
	ErrChatAlreadyExists        = errors.New("chat already exists")
	ErrChatWriteForbidden       = errors.New("sending messages in this chat is forbidden")
	ErrChatPendingFailureReview = errors.New("Нельзя отправить сообщение: по сделке ожидается решение администратора")
)

type UserMessageError struct {
	err error
	msg string
}

func NewUserMessageError(err error, msg string) error {
	return UserMessageError{err: err, msg: msg}
}

func (e UserMessageError) Error() string {
	return e.msg
}

func (e UserMessageError) Unwrap() error {
	return e.err
}
