package domain

import (
	"barter-port/internal/deals/domain/enums"
	"time"

	"github.com/google/uuid"
)

type Deal struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	Status      enums.DealStatus
	Items       []Item
}
