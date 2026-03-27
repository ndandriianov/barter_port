package domain

import (
	ucstatus "barter-port/internal/auth/domain/uc-status"
	"time"

	"github.com/google/uuid"
)

type UserCreationEvent struct {
	UserID    uuid.UUID     `json:"user_id" db:"user_id"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
	Status    ucstatus.Enum `json:"status" db:"status"`
}
