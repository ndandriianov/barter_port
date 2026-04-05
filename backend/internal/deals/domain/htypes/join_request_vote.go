package htypes

import "github.com/google/uuid"

type JoinRequestVote struct {
	UserID  uuid.UUID
	DealID  uuid.UUID
	VoterID uuid.UUID
	Vote    bool
}
