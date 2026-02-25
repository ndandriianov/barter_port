package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	Email         string
	PasswordHash  string
	EmailVerified bool
	CreatedAt     time.Time
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
