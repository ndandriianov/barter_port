package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReputationEvent struct {
	Id         uuid.UUID `db:"id"`
	SourceType string    `db:"source_type"`
	SourceID   uuid.UUID `db:"source_id"`
	Delta      int       `db:"delta"`
	CreatedAt  time.Time `db:"created_at"`
	Comment    *string   `db:"comment"`
}
