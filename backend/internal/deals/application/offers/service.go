package offers

import (
	"barter-port/internal/deals/domain"
	offersrep "barter-port/internal/deals/infrastructure/repository/offers"
	"barter-port/pkg/logger"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type Service struct {
	repo           *offersrep.Repository
	fallbackLogger *slog.Logger
}

func NewService(offerRepository *offersrep.Repository, fallbackLogger *slog.Logger) *Service {
	return &Service{repo: offerRepository, fallbackLogger: fallbackLogger}
}

func (s *Service) CreateOffer(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	itemType domain.ItemType,
	action domain.OfferAction,
	description string,
) (*domain.Offer, error) {
	if name == "" {
		return nil, domain.ErrInvalidOfferName
	}

	item := domain.Offer{
		ID:          uuid.New(),
		AuthorId:    userID,
		Name:        name,
		Type:        itemType,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
		Views:       0,
	}

	err := s.repo.AddOffer(ctx, item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// GetOffers retrieves items based on the provided query parameters.
// It supports pagination through the nextCursor and limit parameters, and sorting based on the sortType.
//
// Errors: only internal
func (s *Service) GetOffers(
	ctx context.Context,
	sortType domain.SortType,
	cursor *domain.UniversalCursor,
	limit int,
) ([]domain.Offer, *domain.UniversalCursor, error) {

	log := logger.LogFrom(ctx, s.fallbackLogger)

	switch sortType {
	case domain.ByTime:
		var timeCursor *domain.TimeCursor
		var err error

		if cursor != nil {
			timeCursor, err = cursor.ToTimeCursor()
		}
		if err != nil || timeCursor == nil {
			log.Debug("time cursor is not specified, starting from the beginning", slog.Any("error", err))
		}

		offers, newCursor, err := s.repo.GetOffersOrderByTime(ctx, timeCursor, limit)
		if err != nil {
			return nil, nil, err
		}

		var universalCursor *domain.UniversalCursor
		if newCursor != nil {
			universalCursor = newCursor.ToUniversalCursor()
		}

		return offers, universalCursor, nil

	case domain.ByPopularity:
		var popularityCursor *domain.PopularityCursor
		var err error

		if cursor != nil {
			popularityCursor, err = cursor.ToPopularityCursor()
		}
		if err != nil || popularityCursor == nil {
			log.Debug("popularity cursor is not specified, starting from the beginning", slog.Any("error", err))
		}

		offers, newCursor, err := s.repo.GetItemsOrderByPopularity(ctx, popularityCursor, limit)
		if err != nil {
			return nil, nil, err
		}

		var universalCursor *domain.UniversalCursor
		if newCursor != nil {
			universalCursor = newCursor.ToUniversalCursor()
		}

		return offers, universalCursor, nil

	default:
		return nil, nil, fmt.Errorf("invalid sort type: %v", sortType)
	}
}

// TODO: hide
// TODO: unhide
