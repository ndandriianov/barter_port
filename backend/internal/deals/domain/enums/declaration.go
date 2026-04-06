package enums

//go:generate enumer -type=ItemType -json -text -sql -transform=lower -trimprefix=ItemType
type ItemType int

const (
	ItemTypeGood ItemType = iota
	ItemTypeService
)

//go:generate enumer -type=OfferAction -json -text -sql -transform=lower -trimprefix=OfferAction
type OfferAction int

const (
	OfferActionGive OfferAction = iota
	OfferActionTake
)

//go:generate enumer -type=SortType -json -text -sql -trimprefix=SortType
type SortType int

const (
	SortTypeByTime SortType = iota
	SortTypeByPopularity
)

//go:generate enumer -type=DealStatus -json -text -sql -trimprefix=DealStatus
type DealStatus int

const (
	DealStatusLookingForParticipants DealStatus = iota
	DealStatusDiscussion
	DealStatusConfirmed
	DealStatusCompleted
	DealStatusCancelled
	DealStatusFailed
)
