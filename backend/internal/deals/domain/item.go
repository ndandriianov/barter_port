package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/enums"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID          uuid.UUID
	AuthorID    uuid.UUID
	ProviderID  *uuid.UUID
	ReceiverID  *uuid.UUID
	Name        string
	Description string
	Type        enums.ItemType
	UpdatedAt   *time.Time
}

func (i *Item) ToDTO() types.Item {
	return types.Item{
		Id:          i.ID,
		AuthorId:    i.AuthorID,
		ProviderId:  i.ProviderID,
		ReceiverId:  i.ReceiverID,
		Name:        i.Name,
		Description: i.Description,
		Type:        types.ItemType(i.Type.String()),
		UpdatedAt:   i.UpdatedAt,
	}
}
