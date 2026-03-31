package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"time"

	"github.com/google/uuid"
)

type Deal struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	Items       []Item
}

func (d *Deal) ToDTO() types.Deal {
	itemsDTO := make([]types.Item, len(d.Items))
	for i, item := range d.Items {
		itemsDTO[i] = item.ToDTO()
	}
	return types.Deal{
		Id:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		Items:       itemsDTO,
	}
}
