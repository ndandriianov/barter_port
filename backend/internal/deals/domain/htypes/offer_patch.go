package htypes

import "barter-port/internal/deals/domain/enums"

type OfferPatch struct {
	Name        *string
	Description *string
	Type        *enums.ItemType
	Action      *enums.OfferAction
}
