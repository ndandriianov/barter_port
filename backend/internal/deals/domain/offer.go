package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/enums"
	"time"

	"github.com/google/uuid"
)

type Offer struct {
	ID          uuid.UUID         `db:"id"`
	AuthorId    uuid.UUID         `db:"author_id"`
	AuthorName  *string           `db:"-"`
	Name        string            `db:"name"`
	PhotoUrls   []string          `db:"photo_urls"`
	Type        enums.ItemType    `db:"type"`
	Action      enums.OfferAction `db:"action"`
	Description string            `db:"description"`
	CreatedAt   time.Time         `db:"created_at"`
	// TODO: добавить updated at
	Views int `db:"views"`
}

func (i *Offer) ToDto() types.Offer {
	var photoURLs *[]string
	if len(i.PhotoUrls) > 0 {
		copied := append([]string(nil), i.PhotoUrls...)
		photoURLs = &copied
	}

	return types.Offer{
		Id:          i.ID,
		AuthorId:    i.AuthorId,
		AuthorName:  i.AuthorName,
		Name:        i.Name,
		PhotoUrls:   photoURLs,
		Type:        types.ItemType(i.Type.String()),
		Action:      types.OfferAction(i.Action.String()),
		Description: i.Description,
		CreatedAt:   i.CreatedAt,
		Views:       int64(i.Views),
	}
}

func (i *Offer) ToDTOWithInfo(info OfferInfo) types.OfferWithInfo {
	var photoURLs *[]string
	if len(i.PhotoUrls) > 0 {
		copied := append([]string(nil), i.PhotoUrls...)
		photoURLs = &copied
	}

	return types.OfferWithInfo{
		Action:      types.OfferAction(i.Action.String()),
		AuthorId:    i.AuthorId,
		AuthorName:  i.AuthorName,
		CreatedAt:   i.CreatedAt,
		Description: i.Description,
		Id:          i.ID,
		Name:        i.Name,
		PhotoUrls:   photoURLs,
		Quantity:    info.Quantity,
		Type:        types.ItemType(i.Type.String()),
		Views:       int64(i.Views),
		Confirmed:   info.Confirmed,
	}
}
