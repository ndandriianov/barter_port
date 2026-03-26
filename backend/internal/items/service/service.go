package service

import (
	"barter-port/internal/items/model"
	"barter-port/pkg/logger"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

var (
	ErrInvalidItemName = errors.New("invalid item name")
)

type Repository interface {
	AddItem(ctx context.Context, item model.Item) error

	GetItemsOrderByTime(
		ctx context.Context,
		cursor *model.TimeCursor,
		limit int,
	) ([]model.Item, *model.TimeCursor, error)

	GetItemsOrderByPopularity(
		ctx context.Context,
		cursor *model.PopularityCursor,
		limit int,
	) ([]model.Item, *model.PopularityCursor, error)
}

type ItemService struct {
	repo           Repository
	fallbackLogger *slog.Logger
}

func NewItemService(itemRepository Repository, fallbackLogger *slog.Logger) *ItemService {
	return &ItemService{repo: itemRepository, fallbackLogger: fallbackLogger}
}

func (s *ItemService) CreateItem(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	itemType model.ItemType,
	action model.ItemAction,
	description string,
) (*model.Item, error) {
	if name == "" {
		return nil, ErrInvalidItemName
	}

	item := model.Item{
		ID:          uuid.New(),
		AuthorId:    userID,
		Name:        name,
		Type:        itemType,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
		Views:       0,
	}

	err := s.repo.AddItem(ctx, item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// GetItems retrieves items based on the provided query parameters.
// It supports pagination through the nextCursor and limit parameters, and sorting based on the sortType.
//
// Errors: only internal
func (s *ItemService) GetItems(
	ctx context.Context,
	sortType model.SortType,
	cursor *model.UniversalCursor,
	limit int,
) ([]model.Item, *model.UniversalCursor, error) {

	log := logger.LogFrom(ctx, s.fallbackLogger)

	switch sortType {
	case model.ByTime:
		var timeCursor *model.TimeCursor
		var err error

		if cursor != nil {
			timeCursor, err = cursor.ToTimeCursor()
		}
		if err != nil || timeCursor == nil {
			log.Debug("time cursor is not specified, starting from the beginning", slog.Any("error", err))
		}

		items, newCursor, err := s.repo.GetItemsOrderByTime(ctx, timeCursor, limit)
		if err != nil {
			return nil, nil, err
		}

		var universalCursor *model.UniversalCursor
		if newCursor != nil {
			universalCursor = newCursor.ToUniversalCursor()
		}

		return items, universalCursor, nil

	case model.ByPopularity:
		var popularityCursor *model.PopularityCursor
		var err error

		if cursor != nil {
			popularityCursor, err = cursor.ToPopularityCursor()
		}
		if err != nil || popularityCursor == nil {
			log.Debug("popularity cursor is not specified, starting from the beginning", slog.Any("error", err))
		}

		items, newCursor, err := s.repo.GetItemsOrderByPopularity(ctx, popularityCursor, limit)
		if err != nil {
			return nil, nil, err
		}

		var universalCursor *model.UniversalCursor
		if newCursor != nil {
			universalCursor = newCursor.ToUniversalCursor()
		}

		return items, universalCursor, nil

	default:
		return nil, nil, fmt.Errorf("invalid sort type: %v", sortType)
	}
}

// TODO: hide
// TODO: unhide
