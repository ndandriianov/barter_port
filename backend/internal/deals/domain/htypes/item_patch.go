package htypes

import "github.com/google/uuid"

type ItemPatch struct {
	// Content fields — only the item author can change these
	Name           *string
	Description    *string
	Quantity       *int
	DeletePhotoIds []uuid.UUID

	// Role fields — governed by claim/release rules
	ClaimProvider   bool
	ReleaseProvider bool
	ClaimReceiver   bool
	ReleaseReceiver bool
}
