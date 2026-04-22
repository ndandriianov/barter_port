package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/enums"
	"time"

	"github.com/google/uuid"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

type Item struct {
	ID          uuid.UUID
	OfferID     *uuid.UUID
	AuthorID    uuid.UUID
	ProviderID  *uuid.UUID
	ReceiverID  *uuid.UUID
	Name        string
	Description string
	Type        enums.ItemType
	UpdatedAt   *time.Time
	Quantity    int
	PhotoIDs    []uuid.UUID
	PhotoURLs   []string
}

func (i *Item) ToDTO() types.Item {
	var photoIDs *[]openapitypes.UUID
	if len(i.PhotoIDs) > 0 {
		copied := make([]openapitypes.UUID, 0, len(i.PhotoIDs))
		for _, id := range i.PhotoIDs {
			copied = append(copied, id)
		}
		photoIDs = &copied
	}

	var photoURLs *[]string
	if len(i.PhotoURLs) > 0 {
		photoURLs = new(append([]string(nil), i.PhotoURLs...))
	}

	return types.Item{
		Id:          i.ID,
		OfferId:     i.OfferID,
		AuthorId:    i.AuthorID,
		ProviderId:  i.ProviderID,
		ReceiverId:  i.ReceiverID,
		Name:        i.Name,
		Description: i.Description,
		Type:        types.ItemType(i.Type.String()),
		UpdatedAt:   i.UpdatedAt,
		Quantity:    i.Quantity,
		PhotoIds:    photoIDs,
		PhotoUrls:   photoURLs,
	}
}
