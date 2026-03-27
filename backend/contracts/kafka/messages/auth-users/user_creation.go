package auth_users

import (
	"time"

	"github.com/google/uuid"
)

const userCreationMessageType = "auth.user.created"

type UserCreationMessage struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (m UserCreationMessage) GetKey() string {
	return m.ID.String()
}

func (m UserCreationMessage) GetCreatedAt() time.Time {
	return m.CreatedAt
}

func (m UserCreationMessage) GetMessageType() string {
	return userCreationMessageType
}
