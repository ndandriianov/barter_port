package enums

//go:generate enumer -type=ItemType -json -text -sql -transform=lower
type ItemType int

const (
	ItemTypeGood ItemType = iota
	ItemTypeService
)

//go:generate enumer -type=OfferAction -json -text -sql -transform=lower
type OfferAction int

const (
	OfferActionGive OfferAction = iota
	OfferActionTake
)

//go:generate enumer -type=SortType -json -text -sql
type SortType int

const (
	SortTypeByTime SortType = iota
	SortTypeByPopularity
)
