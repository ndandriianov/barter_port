package domain

import "github.com/google/uuid"

type ItemPhoto struct {
	ID       uuid.UUID `db:"id"`
	ItemID   uuid.UUID `db:"item_id"`
	URL      string    `db:"url"`
	Position int       `db:"position"`
}
