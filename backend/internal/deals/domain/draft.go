package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"time"

	"github.com/google/uuid"
)

type Draft struct {
	ID          uuid.UUID
	AuthorID    uuid.UUID
	Name        *string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	Items       []OfferWithInfo
}

func (d *Draft) ToDTO() types.Draft {
	itemsDTO := make([]types.ItemWithInfo, len(d.Items))
	for i, item := range d.Items {
		itemsDTO[i] = item.Offer.ToDTOWithInfo(item.Info)
	}

	return types.Draft{
		Id:          d.ID,
		AuthorId:    d.AuthorID,
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		Items:       itemsDTO,
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
