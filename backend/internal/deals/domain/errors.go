package domain

import "errors"

var (
	ErrInvalidOfferName = errors.New("invalid offer name")
	ErrDraftNotFound    = errors.New("draft not found")
	ErrNoOffers         = errors.New("no offers")
)
