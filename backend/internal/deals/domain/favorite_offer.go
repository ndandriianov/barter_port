package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"time"

	"github.com/google/uuid"
)

type FavoritedOffer struct {
	Offer
	FavoritedAt time.Time `db:"favorited_at"`
}

func (o *FavoritedOffer) ToDto() types.FavoritedOffer {
	offerDTO := o.Offer.ToDto()

	return types.FavoritedOffer{
		Action:              offerDTO.Action,
		AuthorId:            offerDTO.AuthorId,
		AuthorName:          offerDTO.AuthorName,
		CreatedAt:           offerDTO.CreatedAt,
		Description:         offerDTO.Description,
		FavoritedAt:         o.FavoritedAt,
		Id:                  offerDTO.Id,
		IsFavorite:          offerDTO.IsFavorite,
		IsHidden:            offerDTO.IsHidden,
		ModificationBlocked: offerDTO.ModificationBlocked,
		Name:                offerDTO.Name,
		PhotoIds:            offerDTO.PhotoIds,
		PhotoUrls:           offerDTO.PhotoUrls,
		Tags:                offerDTO.Tags,
		Type:                offerDTO.Type,
		UpdatedAt:           offerDTO.UpdatedAt,
		Views:               offerDTO.Views,
	}
}

type FavoriteOffersCursor struct {
	FavoritedAt time.Time
	Id          uuid.UUID
}

func (c *FavoriteOffersCursor) ToDto() types.FavoriteOffersCursor {
	return types.FavoriteOffersCursor{
		FavoritedAt: c.FavoritedAt,
		Id:          c.Id,
	}
}
