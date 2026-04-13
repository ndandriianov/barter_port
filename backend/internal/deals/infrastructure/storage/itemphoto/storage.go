package itemphoto

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
		return nil, fmt.Errorf("initialize item photo storage: %w", err)
	}

	return &Storage{storage: storage}, nil
}

func (s *Storage) CopyPhoto(ctx context.Context, sourceURL string, itemID uuid.UUID, index int) (string, error) {
	key := s.objectKey(itemID, index)
	if err := s.storage.CopyManagedObject(ctx, sourceURL, key); err != nil {
		return "", fmt.Errorf("copy item photo object: %w", err)
	}

	return s.ManagedPhotoURL(itemID, index), nil
}

func (s *Storage) DeletePhoto(ctx context.Context, itemID uuid.UUID, index int) error {
	if err := s.storage.DeleteObject(ctx, s.objectKey(itemID, index)); err != nil {
		return fmt.Errorf("delete item photo object: %w", err)
	}

	return nil
}

func (s *Storage) ManagedPhotoURL(itemID uuid.UUID, index int) string {
	return s.storage.ManagedObjectURL(s.objectKey(itemID, index))
}

func (s *Storage) objectKey(itemID uuid.UUID, index int) string {
	return fmt.Sprintf("item-%s/photo-%d", itemID.String(), index)
}
