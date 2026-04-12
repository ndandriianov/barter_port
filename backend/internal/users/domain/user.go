package domain

import (
	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `db:"id"`
	Name      *string   `db:"name"`
	Bio       *string   `db:"bio"`
	AvatarURL *string   `db:"avatar_url"`
}
