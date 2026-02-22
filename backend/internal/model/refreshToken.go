package model

import "time"

type RefreshToken struct {
	JTI       string
	UserID    string
	ExpiresAt time.Time
	Revoked   bool
}
