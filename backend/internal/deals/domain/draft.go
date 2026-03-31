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
	Items       []ItemWithInfo
}

func (d *Draft) ToDTO() types.Draft {
	itemsDTO := make([]types.ItemWithInfo, len(d.Items))
	for i, item := range d.Items {
		itemsDTO[i] = item.Item.ToDTOWithInfo(item.Info)
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

type ItemInfo struct {
	Quantity   int
	ReceiverID *uuid.UUID
}

type ItemIDsAndInfo struct {
	ID   uuid.UUID
	Info ItemInfo
}

type ItemWithInfo struct {
	Item Item
	Info ItemInfo
}
