package model

import (
	"time"

	"github.com/google/uuid"
)

//go:generate enumer -type=ItemType -json -text -sql
type ItemType int

const (
	Good ItemType = iota
	Service
)

//go:generate enumer -type=ItemAction -json -text -sql
type ItemAction int

const (
	Give ItemAction = iota
	Take
)

type Item struct {
	ID          uuid.UUID
	Name        string
	Type        ItemType
	Action      ItemAction
	Description string
	CreatedAt   time.Time
}
