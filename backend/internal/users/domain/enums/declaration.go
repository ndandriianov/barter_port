package enums

//go:generate enumer -type=SourceType -json -text -sql -transform=lower -trimprefix=SourceType
type SourceType int

const (
	SourceTypeOfferReport SourceType = iota
)
