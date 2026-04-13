package offerphoto

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
		return nil, fmt.Errorf("initialize offer photo storage: %w", err)
	}

	return &Storage{storage: storage}, nil
}

func (s *Storage) UploadPhoto(ctx context.Context, offerID uuid.UUID, index int, contentType string, content []byte) (string, error) {
	key := s.objectKey(offerID, index)
	if err := s.storage.PutObject(ctx, key, contentType, content); err != nil {
		return "", fmt.Errorf("put offer photo object: %w", err)
	}

	return s.ManagedPhotoURL(offerID, index), nil
}

func (s *Storage) DeletePhoto(ctx context.Context, offerID uuid.UUID, index int) error {
	if err := s.storage.DeleteObject(ctx, s.objectKey(offerID, index)); err != nil {
		return fmt.Errorf("delete offer photo object: %w", err)
	}

	return nil
}

func (s *Storage) ManagedPhotoURL(offerID uuid.UUID, index int) string {
	return s.storage.ManagedObjectURL(s.objectKey(offerID, index))
}

func (s *Storage) objectKey(offerID uuid.UUID, index int) string {
	return fmt.Sprintf("offer-%s/photo-%d", offerID.String(), index)
}
