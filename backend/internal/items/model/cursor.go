package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCreatedAtIsNil = errors.New("createdAt is nil")
	ErrViewsIsNil     = errors.New("views is nil")
)

type UniversalCursor struct {
	CreatedAt *time.Time `json:"createdAt"`
	Views     *int       `json:"views"`
	Id        uuid.UUID  `json:"id"`
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
