package htypes

import (
	"barter-port/internal/deals/domain/enums"

	"github.com/google/uuid"
)

type DealIDWithParticipantIDs struct {
	ID             uuid.UUID
	Status         enums.DealStatus
	ParticipantIDs []uuid.UUID
}
