package avatar

import (
	"context"
	"fmt"

	"barter-port/pkg/storage/s3storage"

	"github.com/google/uuid"
)

type Config = s3storage.Config

type Storage struct {
	storage *s3storage.Storage
}

func NewStorage(cfg Config) (*Storage, error) {
	storage, err := s3storage.NewStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize avatar storage: %w", err)
	}

	return &Storage{storage: storage}, nil
}

func (s *Storage) UploadAvatar(ctx context.Context, userID uuid.UUID, contentType string, content []byte) (string, error) {
	key := s.objectKey(userID)
	if err := s.storage.PutObject(ctx, key, contentType, content); err != nil {
		return "", fmt.Errorf("put avatar object: %w", err)
	}

	return s.ManagedAvatarURL(userID), nil
}

func (s *Storage) DeleteAvatar(ctx context.Context, userID uuid.UUID) error {
	if err := s.storage.DeleteObject(ctx, s.objectKey(userID)); err != nil {
		return fmt.Errorf("delete avatar object: %w", err)
	}

	return nil
}

func (s *Storage) ManagedAvatarURL(userID uuid.UUID) string {
	return s.storage.ManagedObjectURL(s.objectKey(userID))
}

func (s *Storage) objectKey(userID uuid.UUID) string {
	return "user-" + userID.String() + "/avatar"
}
