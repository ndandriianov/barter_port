package htypes

import (
	"github.com/google/uuid"
)

type DraftIDWithAuthorIDs struct {
	ID             uuid.UUID
	ParticipantIDs []uuid.UUID
}
