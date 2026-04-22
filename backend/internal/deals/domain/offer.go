package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/enums"
	"time"

	"github.com/google/uuid"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

type Offer struct {
	ID                  uuid.UUID         `db:"id"`
	AuthorId            uuid.UUID         `db:"author_id"`
	AuthorName          *string           `db:"-"`
	IsFavorite          *bool             `db:"-"`
	Name                string            `db:"name"`
	Tags                []string          `db:"tags"`
	PhotoIds            []uuid.UUID       `db:"photo_ids"`
	PhotoUrls           []string          `db:"photo_urls"`
	Type                enums.ItemType    `db:"type"`
	Action              enums.OfferAction `db:"action"`
	Description         string            `db:"description"`
	CreatedAt           time.Time         `db:"created_at"`
	UpdatedAt           *time.Time        `db:"updated_at"`
	Views               int               `db:"views"`
	IsHidden            bool              `db:"is_hidden"`
	ModificationBlocked bool              `db:"modification_blocked"`
}

func (i *Offer) ToDto() types.Offer {
	var photoIDs *[]openapitypes.UUID
	if len(i.PhotoIds) > 0 {
		copied := make([]openapitypes.UUID, 0, len(i.PhotoIds))
		for _, id := range i.PhotoIds {
			copied = append(copied, id)
		}
		photoIDs = &copied
	}

	var photoURLs *[]string
	if len(i.PhotoUrls) > 0 {
		copied := append([]string(nil), i.PhotoUrls...)
		photoURLs = &copied
	}

	return types.Offer{
		Id:                  i.ID,
		AuthorId:            i.AuthorId,
		AuthorName:          i.AuthorName,
		Name:                i.Name,
		Tags:                append([]types.TagName(nil), i.tagsToDTO()...),
		PhotoIds:            photoIDs,
		PhotoUrls:           photoURLs,
		Type:                types.ItemType(i.Type.String()),
		Action:              types.OfferAction(i.Action.String()),
		Description:         i.Description,
		CreatedAt:           i.CreatedAt,
		UpdatedAt:           i.UpdatedAt,
		Views:               int64(i.Views),
		IsFavorite:          i.IsFavorite,
		IsHidden:            boolPtr(i.IsHidden),
		ModificationBlocked: boolPtr(i.ModificationBlocked),
	}
}

func (i *Offer) ToDTOWithInfo(info OfferInfo) types.OfferWithInfo {
	var photoIDs *[]openapitypes.UUID
	if len(i.PhotoIds) > 0 {
		copied := make([]openapitypes.UUID, 0, len(i.PhotoIds))
		for _, id := range i.PhotoIds {
			copied = append(copied, id)
		}
		photoIDs = &copied
	}

	var photoURLs *[]string
	if len(i.PhotoUrls) > 0 {
		copied := append([]string(nil), i.PhotoUrls...)
		photoURLs = &copied
	}

	return types.OfferWithInfo{
		Action:              types.OfferAction(i.Action.String()),
		AuthorId:            i.AuthorId,
		AuthorName:          i.AuthorName,
		CreatedAt:           i.CreatedAt,
		Description:         i.Description,
		Id:                  i.ID,
		IsFavorite:          i.IsFavorite,
		IsHidden:            boolPtr(i.IsHidden),
		ModificationBlocked: boolPtr(i.ModificationBlocked),
		Name:                i.Name,
		PhotoIds:            photoIDs,
		PhotoUrls:           photoURLs,
		Quantity:            info.Quantity,
		Tags:                append([]types.TagName(nil), i.tagsToDTO()...),
		Type:                types.ItemType(i.Type.String()),
		UpdatedAt:           i.UpdatedAt,
		Views:               int64(i.Views),
		Confirmed:           info.Confirmed,
	}
}

func (i *Offer) tagsToDTO() []types.TagName {
	if len(i.Tags) == 0 {
		return []types.TagName{}
	}

	tags := make([]types.TagName, 0, len(i.Tags))
	for _, tag := range i.Tags {
		tags = append(tags, types.TagName(tag))
	}

	return tags
}

func boolPtr(v bool) *bool {
	return &v
}
