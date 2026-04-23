package domain

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	TokenHash string    `db:"token_hash"`
	UserID    uuid.UUID `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
	Used      bool      `db:"used"`
	CreatedAt time.Time `db:"created_at"`
}

func NewPasswordResetToken(tokenHash string, userID uuid.UUID, expiresAt time.Time) PasswordResetToken {
	return PasswordResetToken{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
	}
}
