package htypes

import "github.com/google/uuid"

type DealIDWithParticipantIDs struct {
	ID             uuid.UUID
	ParticipantIDs []uuid.UUID
}
