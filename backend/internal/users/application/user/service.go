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

// AddUser adds user if not exist. Unique check by id.
//
// Errors:
//   - model.ErrUserAlreadyExists: Occurs if a user with the same id already exists in the repository.
func (s *Service) AddUser(ctx context.Context, id uuid.UUID, name string) error {
	user := model.User{
		Id:   id,
		Name: name,
	}

	return s.repository.AddUser(ctx, user)
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// TODO: implement with transaction
	panic("implement me")
}

// UpdateName updates users name by id.
//
// Errors:
//   - model.ErrUserNotFound: Occurs if no user is found with the given id.
func (s *Service) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	return s.repository.UpdateName(ctx, id, name)
}

// UpdateBio updates users bio by id. Bio can be null.
//
// Errors:
//   - model.ErrUserNotFound: Occurs if no user is found with the given id.
func (s *Service) UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error {
	return s.repository.UpdateBio(ctx, id, bio)
}
