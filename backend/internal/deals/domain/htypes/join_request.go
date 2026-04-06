package htypes

import "github.com/google/uuid"

type JoinRequest struct {
	UserID uuid.UUID
	DealID uuid.UUID
}

type JoinRequestWithVoters struct {
	UserID   uuid.UUID
	DealID   uuid.UUID
	VoterIDs []uuid.UUID
}
