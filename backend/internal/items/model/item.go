package model

import (
	"barter-port/contracts/openapi/items/types"
	"time"

	"github.com/google/uuid"
)

//go:generate enumer -type=ItemType -json -text -sql -transform=lower
type ItemType int

const (
	Good ItemType = iota
	Service
)

//go:generate enumer -type=ItemAction -json -text -sql -transform=lower
type ItemAction int

const (
	Give ItemAction = iota
	Take
)

type Item struct {
	ID          uuid.UUID
	AuthorId    uuid.UUID
	Name        string
	Type        ItemType
	Action      ItemAction
	Description string
	CreatedAt   time.Time
	Views       int
}

func (i Item) ToDto() types.Item {
	return types.Item{
		Action:      types.ItemAction(i.Action.String()),
		AuthorId:    i.AuthorId,
		CreatedAt:   i.CreatedAt,
		Description: i.Description,
		Id:          i.ID,
		Name:        i.Name,
		Type:        types.ItemType(i.Type.String()),
		Views:       int64(i.Views),
	}
}
