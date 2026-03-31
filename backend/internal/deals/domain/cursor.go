package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

var (
	ErrCreatedAtIsNil = errors.New("createdAt is nil")
	ErrViewsIsNil     = errors.New("views is nil")
	ErrInvalidId      = errors.New("invalid id")
)

type UniversalCursor struct {
	CreatedAt *time.Time `json:"createdAt" example:"2026-03-04T12:00:00Z"`
	Views     *int       `json:"views"  example:"120"`
	Id        uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

func NewUniversalCursor(createdAtStr, viewsStr, idStr string) (*UniversalCursor, error) {
	var createdAtPtr *time.Time
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		createdAtPtr = nil
	} else {
		createdAtPtr = &createdAt
	}

	var viewsPtr *int
	views, err := strconv.Atoi(viewsStr)
	if err != nil {
		viewsPtr = nil
	} else {
		viewsPtr = &views
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, ErrInvalidId
	}

	return &UniversalCursor{
		CreatedAt: createdAtPtr,
		Views:     viewsPtr,
		Id:        id,
	}, nil
}

func (c *UniversalCursor) ToDto() types.OffersCursor {
	var views *int64
	if c.Views != nil {
		views = new(int64(*c.Views))
	}

	return types.OffersCursor{
		CreatedAt: c.CreatedAt,
		Id:        openapitypes.UUID{},
		Views:     views,
	}
}

type TimeCursor struct {
	CreatedAt time.Time
	Id        uuid.UUID
}

type PopularityCursor struct {
	Views int
	Id    uuid.UUID
}

func (c *UniversalCursor) ToTimeCursor() (*TimeCursor, error) {
	if c.CreatedAt == nil {
		return nil, ErrCreatedAtIsNil
	}
	return &TimeCursor{
		CreatedAt: *c.CreatedAt,
		Id:        c.Id,
	}, nil
}

func (c *UniversalCursor) ToPopularityCursor() (*PopularityCursor, error) {
	if c.Views == nil {
		return nil, ErrViewsIsNil
	}
	return &PopularityCursor{
		Views: *c.Views,
		Id:    c.Id,
	}, nil
}

func (c *TimeCursor) ToUniversalCursor() *UniversalCursor {
	return &UniversalCursor{
		CreatedAt: &c.CreatedAt,
		Id:        c.Id,
	}
}

func (c *PopularityCursor) ToUniversalCursor() *UniversalCursor {
	return &UniversalCursor{
		Views: &c.Views,
		Id:    c.Id,
	}
}
