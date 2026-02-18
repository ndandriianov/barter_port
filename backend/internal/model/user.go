package model

import "time"

type User struct {
	ID            string
	Email         string
	PasswordHash  string
	EmailVerified bool
	CreatedAt     time.Time
}
