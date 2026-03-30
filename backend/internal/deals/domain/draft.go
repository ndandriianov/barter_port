package domain

import (
	"time"

	"github.com/google/uuid"
)

type Draft struct {
	ID          uuid.UUID
	AuthorID    uuid.UUID
	Name        *string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	Items       []Item
}

type ItemIDsAndQuantities struct {
	ID       uuid.UUID
	Quantity int
}
