package htypes

import (
	"barter-port/internal/deals/domain/enums"

	"github.com/google/uuid"
)

type OfferPatch struct {
	Name           *string
	Description    *string
	Type           *enums.ItemType
	Action         *enums.OfferAction
	Tags           *[]string
	DeletePhotoIds []uuid.UUID
	// UpdateLocation=true means the location fields should be applied (even if both are nil = clear).
	UpdateLocation bool
	Latitude       *float64
	Longitude      *float64
}
