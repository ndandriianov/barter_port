package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"time"

	"github.com/google/uuid"
)

//go:generate enumer -type=ItemType -json -text -sql -transform=lower
type ItemType int

const (
	Good ItemType = iota
	Service
)

//go:generate enumer -type=OfferAction -json -text -sql -transform=lower
type OfferAction int

const (
	Give OfferAction = iota
	Take
)

type Offer struct {
	ID          uuid.UUID
	AuthorId    uuid.UUID
	Name        string
	Type        ItemType
	Action      OfferAction
	Description string
	CreatedAt   time.Time
	Views       int
}

func (i *Offer) ToDto() types.Item {
	return types.Item{
		Id:          i.ID,
		AuthorId:    i.AuthorId,
		Name:        i.Name,
		Type:        types.ItemType(i.Type.String()),
		Action:      types.ItemAction(i.Action.String()),
		Description: i.Description,
		CreatedAt:   i.CreatedAt,
		Views:       int64(i.Views),
	}
}

func (i *Offer) ToDTOWithInfo(info OfferInfo) types.ItemWithInfo {
	return types.ItemWithInfo{
		Action:      types.ItemAction(i.Action.String()),
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
