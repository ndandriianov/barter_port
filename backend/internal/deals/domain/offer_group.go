package domain

import "github.com/google/uuid"

type OfferGroup struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Units       []OfferGroupUnit
}

type OfferGroupUnit struct {
	ID     uuid.UUID
	Offers []Offer
}

type OfferGroupUnitCreateInput struct {
	OfferIDs []uuid.UUID
}
