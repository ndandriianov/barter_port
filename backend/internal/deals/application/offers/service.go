package offers

import (
	userspb "barter-port/contracts/grpc/users/v1"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
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
	usersClient    userspb.UsersServiceClient
	fallbackLogger *slog.Logger
}

func NewService(offerRepository *offersrep.Repository, usersClient userspb.UsersServiceClient, fallbackLogger *slog.Logger) *Service {
	return &Service{repo: offerRepository, usersClient: usersClient, fallbackLogger: fallbackLogger}
}

func (s *Service) CreateOffer(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	itemType enums.ItemType,
	action enums.OfferAction,
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
	sortType enums.SortType,
	cursor *domain.UniversalCursor,
	limit int,
	authorID *uuid.UUID,
) ([]domain.Offer, *domain.UniversalCursor, error) {

	log := logger.LogFrom(ctx, s.fallbackLogger)

	var universalCursor *domain.UniversalCursor
	var offers []domain.Offer

	switch sortType {
	case enums.SortTypeByTime:
		var timeCursor *domain.TimeCursor
		var err error

		if cursor != nil {
			timeCursor, err = cursor.ToTimeCursor()
		}
		if err != nil || timeCursor == nil {
			log.Debug("time cursor is not specified, starting from the beginning", slog.Any("error", err))
		}

		offers, timeCursor, err = s.repo.GetOffersOrderByTime(ctx, timeCursor, limit, authorID)
		if err != nil {
			return nil, nil, err
		}

		if timeCursor != nil {
			universalCursor = timeCursor.ToUniversalCursor()
		}

	case enums.SortTypeByPopularity:
		var popularityCursor *domain.PopularityCursor
		var err error

		if cursor != nil {
			popularityCursor, err = cursor.ToPopularityCursor()
		}
		if err != nil || popularityCursor == nil {
			log.Debug("popularity cursor is not specified, starting from the beginning", slog.Any("error", err))
		}

		offers, popularityCursor, err = s.repo.GetOffersOrderByPopularity(ctx, popularityCursor, limit, authorID)
		if err != nil {
			return nil, nil, err
		}

		if popularityCursor != nil {
			universalCursor = popularityCursor.ToUniversalCursor()
		}

	default:
		return nil, nil, fmt.Errorf("invalid sort type: %v", sortType)
	}

	ids := make([]string, len(offers))
	for i, o := range offers {
		ids[i] = o.AuthorId.String()
	}

	response, err := s.usersClient.GetUsersWithInfo(ctx, &userspb.GetUsersWithInfoRequest{Ids: ids})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get author names: %w", err)
	}

	for i, info := range response.Users {
		if info == nil {
			continue // буду считать что пользователь с неуказанным именем
		}
		if offers[i].ID.String() == info.Id {
			offers[i].Name = info.Name
		}
	}

	return offers, universalCursor, nil
}

// TODO: hide
// TODO: unhide
