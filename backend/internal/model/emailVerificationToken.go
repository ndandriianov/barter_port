package model

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerificationToken struct {
	TokenHash string
	UserID    uuid.UUID
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

func NewEmailVerificationToken(tokenHash string, userID uuid.UUID, expiresAt time.Time) EmailVerificationToken {
	return EmailVerificationToken{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
	}
}
