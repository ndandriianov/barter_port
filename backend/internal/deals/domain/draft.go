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
	Items       []Item
}

func (d *Draft) ToDTO() types.Draft {
	itemsDTO := make([]types.Item, len(d.Items))
	for i, item := range d.Items {
		itemsDTO[i] = item.ToDto()
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

type ItemIDsAndQuantities struct {
	ID       uuid.UUID
	Quantity int
}
