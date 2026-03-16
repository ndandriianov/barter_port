package model

import "github.com/google/uuid"

type DraftDeal struct {
	Id          uuid.UUID `json:"id"`
	Items       []Item    `json:"items"`
	Description string    `json:"description"`
}
