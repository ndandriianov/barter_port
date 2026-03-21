package auth_users

import (
	"time"

	"github.com/google/uuid"
)

type UserCreationEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
