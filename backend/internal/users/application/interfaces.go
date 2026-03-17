package application

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type UserService interface {
	AddUser(ctx context.Context, id uuid.UUID, name string) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	UpdateName(ctx context.Context, id uuid.UUID, name string) error
	UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error
}
