package htypes

import (
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"

	"github.com/google/uuid"
)

type DealIDWithParticipantIDs struct {
	ID             uuid.UUID
	Status         enums.DealStatus
	Name           *string
	ParticipantIDs []uuid.UUID
	Items          []domain.Item
}
