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

// GetItems retrieves items based on the provided query parameters.
// It supports pagination through the nextCursor and limit parameters, and sorting based on the sortType.
//
// Errors:
// - ErrInvalidSortType: returned when an unsupported sort type is provided.
func (s *ItemService) GetItems(
	ctx context.Context,
	nextCursor uuid.UUID,
	limit int,
	sortType model.SortType,
) ([]model.Item, error) {

	switch sortType {
	case model.ByTime:
		return s.repo.GetItemsOrderByTime(ctx, nextCursor, limit)
	case model.ByPopularity:
		return s.repo.GetItemsOrderByPopularity(ctx, nextCursor, limit)
	default:
		return nil, ErrInvalidSortType
	}
}

// TODO: hide
// TODO: unhide
