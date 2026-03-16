package model

//go:generate enumer -type=SortType -json -text -sql
type SortType int

const (
	ByTime SortType = iota
	ByPopularity
)
