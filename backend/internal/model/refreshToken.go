package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	JTI       string
	UserID    uuid.UUID
	ExpiresAt time.Time
	Revoked   bool
}
