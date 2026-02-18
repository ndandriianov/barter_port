package model

import "time"

type EmailVerificationToken struct {
	TokenHash string
	UserID    string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

func NewEmailVerificationToken(tokenHash, userID string, expiresAt time.Time) EmailVerificationToken {
	return EmailVerificationToken{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
	}
}
