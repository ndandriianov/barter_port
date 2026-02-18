package model

import "time"

type User struct {
	ID            string
	Email         string
	PasswordHash  string
	EmailVerified bool
	CreatedAt     time.Time
}

func NewUser(id, email, passwordHash string) User {
	return User{
		ID:            id,
		Email:         email,
		PasswordHash:  passwordHash,
		EmailVerified: false,
		CreatedAt:     time.Now(),
	}
}
