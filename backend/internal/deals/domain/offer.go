package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/enums"
	"time"

	"github.com/google/uuid"
)

type Offer struct {
	ID          uuid.UUID
	AuthorId    uuid.UUID
	Name        string
	Type        enums.ItemType
	Action      enums.OfferAction
	Description string
	CreatedAt   time.Time
	// TODO: добавить updated at
	Views int
}

func (i *Offer) ToDto() types.Offer {
	return types.Offer{
		Id:          i.ID,
		AuthorId:    i.AuthorId,
		Name:        i.Name,
		Type:        types.ItemType(i.Type.String()),
		Action:      types.OfferAction(i.Action.String()),
		Description: i.Description,
		CreatedAt:   i.CreatedAt,
		Views:       int64(i.Views),
	}
}

func (i *Offer) ToDTOWithInfo(info OfferInfo) types.OfferWithInfo {
	return types.OfferWithInfo{
		Action:      types.OfferAction(i.Action.String()),
		AuthorId:    i.AuthorId,
		CreatedAt:   i.CreatedAt,
		Description: i.Description,
		Id:          i.ID,
		Name:        i.Name,
		Quantity:    info.Quantity,
		Type:        types.ItemType(i.Type.String()),
		Views:       int64(i.Views),
	}
}
