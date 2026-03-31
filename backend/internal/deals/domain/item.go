package domain

import (
	"barter-port/internal/deals/domain/enums"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID          uuid.UUID
	AuthorID    uuid.UUID
	ProviderID  *uuid.UUID
	ReceiverID  *uuid.UUID
	Name        string
	Description string
	Type        enums.ItemType
	UpdatedAt   *time.Time
}
