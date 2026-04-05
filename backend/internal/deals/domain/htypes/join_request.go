package htypes

import "github.com/google/uuid"

type JoinRequest struct {
	UserID uuid.UUID
	DealID uuid.UUID
}
