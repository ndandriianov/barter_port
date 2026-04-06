package uc_status

//go:generate enumer -type=Enum -json -text -sql
type Enum int

const (
	New Enum = iota
	Success
	Failed
)
