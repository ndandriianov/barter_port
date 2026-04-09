package htypes

import (
	"barter-port/internal/deals/domain"

	"github.com/google/uuid"
)

type FailureVote struct {
	UserID uuid.UUID
	Vote   uuid.UUID
}

type FailureRecord struct {
	DealID           uuid.UUID
	UserID           *uuid.UUID
	ConfirmedByAdmin *bool
	AdminComment     *string
	PunishmentPoints *int
}

type FailureMaterials struct {
	Deal   domain.Deal
	ChatID *uuid.UUID
}
