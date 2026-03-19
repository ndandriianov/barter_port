package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID `db:"id"`
	Email         string    `db:"email"`
	PasswordHash  string    `db:"password_hash"`
	EmailVerified bool      `db:"email_verified"`
	CreatedAt     time.Time `db:"created_at"`
}

func NewUser(id uuid.UUID, email string, passwordHash string) User {
	return User{
		ID:            id,
		Email:         email,
		PasswordHash:  passwordHash,
		EmailVerified: false,
		CreatedAt:     time.Now(),
	}
}
