package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Draft struct {
	ID           uuid.UUID
	AuthorID     uuid.UUID
	Name         *string
	Description  *string
	OfferGroupID *uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	Offers       []OfferWithInfo
}

func (d *Draft) ToDTO() types.Draft {
	itemsDTO := make([]types.OfferWithInfo, len(d.Offers))
	for i, item := range d.Offers {
		itemsDTO[i] = item.Offer.ToDTOWithInfo(item.Info)
	}

	var offerGroupID *openapi_types.UUID
	if d.OfferGroupID != nil {
		converted := openapi_types.UUID(*d.OfferGroupID)
		offerGroupID = &converted
	}

	return types.Draft{
		Id:           d.ID,
		AuthorId:     d.AuthorID,
		Name:         d.Name,
		Description:  d.Description,
		OfferGroupId: offerGroupID,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		Offers:       itemsDTO,
	}
}

type OfferInfo struct {
	Quantity  int
	Confirmed *bool
}

type OfferIDAndInfo struct {
	ID   uuid.UUID
	Info OfferInfo
}

type OfferWithInfo struct {
	Offer Offer
	Info  OfferInfo
}
