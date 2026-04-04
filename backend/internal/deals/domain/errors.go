package domain

import "errors"

var (
	ErrInvalidOfferName = errors.New("invalid offer name")
	ErrDraftNotFound    = errors.New("draft not found")
	ErrDealNotFound     = errors.New("deal not found")
	ErrItemNotFound     = errors.New("item not found")
	ErrNoOffers         = errors.New("no offers")
	ErrOfferNotFound    = errors.New("offer not found")
	ErrUserNotInDraft   = errors.New("user not in draft")
	ErrForbidden        = errors.New("forbidden")
	ErrRoleAlreadyTaken = errors.New("role already taken by another user")
	ErrNotRoleHolder    = errors.New("user does not hold this role")
)
