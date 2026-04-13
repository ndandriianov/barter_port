package domain

import "github.com/google/uuid"

type OfferPhoto struct {
	ID       uuid.UUID `db:"id"`
	OfferID  uuid.UUID `db:"offer_id"`
	URL      string    `db:"url"`
	Position int       `db:"position"`
}
