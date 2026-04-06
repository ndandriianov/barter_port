package htypes

type ItemPatch struct {
	// Content fields — only the item author can change these
	Name        *string
	Description *string
	Quantity    *int

	// Role fields — governed by claim/release rules
	ClaimProvider   bool
	ReleaseProvider bool
	ClaimReceiver   bool
	ReleaseReceiver bool
}
