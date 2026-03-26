package users_auth

import (
	"time"

	"github.com/google/uuid"
)

const ucResultMessageType = "users.auth.uc_result"

type UCResultMessage struct {
	ID        uuid.UUID `json:"id" db:"id"`
	EventID   uuid.UUID `json:"event_id" db:"event_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (m UCResultMessage) GetKey() string {
	return m.ID.String()
}

func (m UCResultMessage) GetCreatedAt() time.Time {
	return m.CreatedAt
}

func (m UCResultMessage) GetMessageType() string {
	return ucResultMessageType
}
