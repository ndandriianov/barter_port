package domain

import "github.com/google/uuid"

type DraftDeal struct {
	Id          uuid.UUID `json:"id"`
	Items       []Offer   `json:"items"`
	Description string    `json:"description"`
}
