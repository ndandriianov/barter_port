package auth_users

import (
	"time"

	"github.com/google/uuid"
)

type UserCreationMessage struct {
	ID        uuid.UUID `json:"id" db:"id"`
	EventID   uuid.UUID `json:"event_id" db:"event_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
