package model

import "time"

type RefreshToken struct {
	TokenHash string
	UserID    string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}
