package user

import (
	"barter-port/internal/users/model"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type UsersRepository interface {
	AddUser(ctx context.Context, user model.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	UpdateName(ctx context.Context, id uuid.UUID, name string) error
	UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error
}

type Service struct {
	repository UsersRepository
}

func NewService(repository UsersRepository) *Service {
	return &Service{repository}
}

func (s *Service) AddUser(ctx context.Context, id uuid.UUID, name string) error {
	user := model.User{
		Id:   id,
		Name: name,
	}

	return s.repository.AddUser(ctx, user)
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteUser(ctx, id)
}

func (s *Service) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	return s.repository.UpdateName(ctx, id, name)
}

func (s *Service) UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error {
	return s.repository.UpdateBio(ctx, id, bio)
}
