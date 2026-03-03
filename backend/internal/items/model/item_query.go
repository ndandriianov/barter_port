package model

import "github.com/google/uuid"

//go:generate enumer -type=SortType -json -text -sql
type SortType int

const (
	ByTime SortType = iota
	ByPopularity
)

type ItemQuery struct {
	NextCursor uuid.UUID `json:"next_cursor"`
	Limit      int       `json:"limit"`
	SortType   SortType  `json:"sort_type"`
}
