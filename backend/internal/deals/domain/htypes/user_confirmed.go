package htypes

import "github.com/google/uuid"

type UserConfirmed struct {
	UserID    uuid.UUID
	Confirmed bool
}
