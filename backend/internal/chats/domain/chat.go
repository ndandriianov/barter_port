package domain

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID           uuid.UUID
	DealID       *uuid.UUID
	Participants []uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type Message struct {
	ID        uuid.UUID
	ChatID    uuid.UUID
	SenderID  uuid.UUID
	Content   string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
