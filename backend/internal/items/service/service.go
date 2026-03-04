package service

import (
	"barter-port/internal/items/model"
	"barter-port/internal/libs/platform/logger"
	"errors"
	"log/slog"
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
		Views:       0,
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
	sortType model.SortType,
	cursor model.UniversalCursor,
	limit int,
) ([]model.Item, model.UniversalCursor, error) {

	log := logger.LogFrom(ctx, s.fallbackLogger)

	switch sortType {
	case model.ByTime:
		timeCursor, err := cursor.ToTimeCursor()
		if err != nil {
			log.Debug(
				"time cursor is not specified, starting from the beginning",
				slog.String("error", err.Error()),
			)
		}

		items, newCursor, err := s.repo.GetItemsOrderByTime(ctx, timeCursor, limit)
		if err != nil {
			return nil, model.UniversalCursor{}, err
		}

		return items, *newCursor.ToUniversalCursor(), nil

	case model.ByPopularity:
		popularityCursor, err := cursor.ToPopularityCursor()
		if err != nil {
			log.Debug(
				"popularity cursor is not specified, starting from the beginning",
				slog.String("error", err.Error()),
			)
		}

		items, newCursor, err := s.repo.GetItemsOrderByPopularity(ctx, popularityCursor, limit)
		if err != nil {
			return nil, model.UniversalCursor{}, err
		}

		return items, *newCursor.ToUniversalCursor(), nil

	default:
		return nil, model.UniversalCursor{}, ErrInvalidSortType
	}
}

// TODO: hide
// TODO: unhide
