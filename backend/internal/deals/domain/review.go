package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID         uuid.UUID
	DealID     uuid.UUID
	ItemID     *uuid.UUID
	OfferID    *uuid.UUID
	AuthorID   uuid.UUID
	ProviderID uuid.UUID
	Rating     int
	Comment    *string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

func (r *Review) ToDTO() types.Review {
	dto := types.Review{
		Id:         r.ID,
		DealId:     r.DealID,
		AuthorId:   r.AuthorID,
		ProviderId: r.ProviderID,
		Rating:     r.Rating,
		Comment:    r.Comment,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}

	if r.OfferID != nil {
		dto.OfferRef = &types.OfferRef{OfferId: *r.OfferID}
	}

	if r.ItemID != nil {
		dto.ItemRef = &types.DealItemRef{
			DealId: r.DealID,
			ItemId: *r.ItemID,
		}
	}

	return dto
}
