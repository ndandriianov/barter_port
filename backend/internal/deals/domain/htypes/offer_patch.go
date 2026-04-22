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
}
