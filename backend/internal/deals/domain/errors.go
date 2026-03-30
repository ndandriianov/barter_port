package domain

import "errors"

var (
	ErrInvalidItemName = errors.New("invalid item name")
	ErrDraftNotFound   = errors.New("draft not found")
)
