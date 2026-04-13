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
	key := s.objectKey(userID, uuid.NewString())
	if err := s.storage.PutObject(ctx, key, contentType, content); err != nil {
		return "", fmt.Errorf("put avatar object: %w", err)
	}

	return s.storage.ManagedObjectURL(key), nil
}

func (s *Storage) DeleteAvatar(ctx context.Context, avatarURL string) error {
	if err := s.storage.DeleteManagedObject(ctx, avatarURL); err != nil {
		return fmt.Errorf("delete avatar object: %w", err)
	}

	return nil
}

func (s *Storage) IsManagedAvatarURL(avatarURL string) bool {
	_, ok := s.storage.ObjectKeyFromManagedURL(avatarURL)
	return ok
}

func (s *Storage) objectKey(userID uuid.UUID, objectID string) string {
	return "user-" + userID.String() + "/avatar-" + objectID
}
