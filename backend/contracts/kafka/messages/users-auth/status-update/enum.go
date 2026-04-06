package status_update

//go:generate enumer -type=Enum -json -text -sql
type Enum int

const (
	Success Enum = iota
	Failed
)
