package domain

import "errors"

var (
	ErrInvalidOfferName        = errors.New("invalid offer name")
	ErrInvalidQuantity         = errors.New("invalid quantity")
	ErrDraftNotFound           = errors.New("draft not found")
	ErrDealNotFound            = errors.New("deal not found")
	ErrItemNotFound            = errors.New("item not found")
	ErrFailureNotFound         = errors.New("failure not found")
	ErrNoOffers                = errors.New("no offers")
	ErrOfferNotFound           = errors.New("offer not found")
	ErrUserNotInDraft          = errors.New("user not in draft")
	ErrForbidden               = errors.New("forbidden")
	ErrAdminOnly               = errors.New("admin only")
	ErrRoleAlreadyTaken        = errors.New("role already taken by another user")
	ErrNotRoleHolder           = errors.New("user does not hold this role")
	ErrDuplicateRole           = errors.New("user already holds this item with other role")
	ErrInvalidDealStatus       = errors.New("deal is not in the expected status")
	ErrFailureReviewRequired   = errors.New("deal is pending failure review")
	ErrFailureAlreadyResolved  = errors.New("failure already resolved")
	ErrInvalidFailureDecision  = errors.New("invalid failure decision")
	ErrJoinRequestNotFound     = errors.New("join request not found")
	ErrDealParticipantsUnready = errors.New("deal participants unready")
)
