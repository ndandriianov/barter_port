package model

import "time"

type EmailVerificationToken struct {
	TokenHash string
	UserID    string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}
