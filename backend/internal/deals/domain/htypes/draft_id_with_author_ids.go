package htypes

import "github.com/google/uuid"

type DraftIDWithAuthorIDs struct {
	ID             uuid.UUID
	Name           *string
	ParticipantIDs []uuid.UUID
}
