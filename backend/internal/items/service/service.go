package service

import (
	"barter-port/internal/items/model"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

var (
	ErrInvalidItemName = errors.New("invalid item name")
	ErrInvalidSortType = errors.New("invalid sort type")
)

type Repository interface {
	AddItem(ctx context.Context, item model.Item) error
	GetItemsOrderByTime(ctx context.Context, nextCursor uuid.UUID, limit int) ([]model.Item, error)
	GetItemsOrderByPopularity(ctx context.Context, nextCursor uuid.UUID, limit int) ([]model.Item, error)
}

type ItemService struct {
	repo Repository
}

func NewItemService(itemRepository Repository) *ItemService {
	return &ItemService{repo: itemRepository}
}

func (s *ItemService) CreateItem(
	ctx context.Context,
	name string,
	itemType model.ItemType,
	action model.ItemAction,
	description string,
) error {
	if name == "" {
		return ErrInvalidItemName
	}

	item := model.Item{
		ID:          uuid.New(),
		Name:        name,
		Type:        itemType,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}

	err := s.repo.AddItem(ctx, item)
	if err != nil {
		return err
	}

	return nil
}

func (s *ItemService) GetItems(ctx context.Context, query model.ItemQuery) ([]model.Item, error) {
	switch query.SortType {
	case model.ByTime:
		return s.repo.GetItemsOrderByTime(ctx, query.NextCursor, query.Limit)
	case model.ByPopularity:
		return s.repo.GetItemsOrderByPopularity(ctx, query.NextCursor, query.Limit)
	default:
		return nil, ErrInvalidSortType
	}
}

// TODO: hide
// TODO: unhide
