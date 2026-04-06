package user

import (
	authpb "barter-port/contracts/grpc/auth/v1"
	"barter-port/internal/users/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type UsersRepository interface {
	GetUserById(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateName(ctx context.Context, id uuid.UUID, name string) error
	UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error
	GetNamesForUserIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*string, error)
}

var ErrAuthClientNotConfigured = errors.New("auth grpc client is not configured")

type Me struct {
	Id        uuid.UUID
	Name      *string
	Bio       *string
	Email     string
	CreatedAt time.Time
}

type Service struct {
	repository UsersRepository
	authClient authpb.AuthServiceClient
}

func NewService(repository UsersRepository, authClient authpb.AuthServiceClient) *Service {
	return &Service{repository: repository, authClient: authClient}
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repository.GetUserById(ctx, id)
}

func (s *Service) GetMe(ctx context.Context, id uuid.UUID) (Me, error) {
	u, err := s.repository.GetUserById(ctx, id)
	if err != nil {
		return Me{}, err
	}

	if s.authClient == nil {
		return Me{}, ErrAuthClientNotConfigured
	}

	authMe, err := s.authClient.GetMe(ctx, &authpb.GetMeRequest{Id: id.String()})
	if err != nil {
		return Me{}, err
	}

	var createdAt time.Time
	if ts := authMe.GetCreatedAt(); ts != nil {
		createdAt = ts.AsTime()
	}

	return Me{
		Id:        u.Id,
		Name:      u.Name,
		Bio:       u.Bio,
		Email:     authMe.GetEmail(),
		CreatedAt: createdAt,
	}, nil
}

// UpdateName updates users name by id.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (s *Service) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	return s.repository.UpdateName(ctx, id, name)
}

// UpdateBio updates users bio by id. Bio can be null.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (s *Service) UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error {
	return s.repository.UpdateBio(ctx, id, bio)
}

// GetNamesForUserIDs returns a map of user IDs to their corresponding names.
//
// No domain Errors
func (s *Service) GetNamesForUserIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*string, error) {
	names, err := s.repository.GetNamesForUserIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("repository.GetNamesForUserIDs: %w", err)
	}
	return names, nil
}
