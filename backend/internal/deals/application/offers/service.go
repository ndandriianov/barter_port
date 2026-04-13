package offers

import (
	userspb "barter-port/contracts/grpc/users/v1"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	offersrep "barter-port/internal/deals/infrastructure/repository/offers"
	"barter-port/pkg/db"
	"barter-port/pkg/logger"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

var ErrOfferPhotoStorageNotConfigured = errors.New("offer photo storage is not configured")

type PhotoUpload struct {
	ContentType string
	Content     []byte
}

type PhotoStorage interface {
	UploadPhoto(ctx context.Context, offerID uuid.UUID, index int, contentType string, content []byte) (string, error)
	DeletePhoto(ctx context.Context, offerID uuid.UUID, index int) error
}

type Service struct {
	db             *pgxpool.Pool
	repo           *offersrep.Repository
	photoStorage   PhotoStorage
	usersClient    userspb.UsersServiceClient
	fallbackLogger *slog.Logger
}

func NewService(dbPool *pgxpool.Pool, offerRepository *offersrep.Repository, usersClient userspb.UsersServiceClient, photoStorage PhotoStorage, fallbackLogger *slog.Logger) *Service {
	return &Service{db: dbPool, repo: offerRepository, photoStorage: photoStorage, usersClient: usersClient, fallbackLogger: fallbackLogger}
}

func (s *Service) CreateOffer(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	itemType enums.ItemType,
	action enums.OfferAction,
	description string,
	photos []PhotoUpload,
) (*domain.Offer, error) {
	if name == "" {
		return nil, domain.ErrInvalidOfferName
	}
	if len(photos) > 0 && s.photoStorage == nil {
		return nil, ErrOfferPhotoStorageNotConfigured
	}

	log := logger.LogFrom(ctx, s.fallbackLogger).With(
		slog.String("user_id", userID.String()),
	)

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

	photoURLs := make([]string, 0, len(photos))
	for i, photo := range photos {
		photoURL, err := s.photoStorage.UploadPhoto(ctx, item.ID, i, photo.ContentType, photo.Content)
		if err != nil {
			s.cleanupUploadedPhotos(ctx, item.ID, i)
			return nil, err
		}
		log.Debug(
			"offer photo uploaded successfully",
			slog.String("offer_id", item.ID.String()),
			slog.Int("photo_index", i),
			slog.String("content_type", photo.ContentType),
			slog.Int("size_bytes", len(photo.Content)),
			slog.String("photo_url", photoURL),
		)
		photoURLs = append(photoURLs, photoURL)
	}
	item.PhotoUrls = photoURLs

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.repo.AddOffer(ctx, tx, item); err != nil {
			return err
		}
		return s.repo.AddOfferPhotos(ctx, tx, item.ID, photoURLs)
	})
	if err != nil {
		s.cleanupUploadedPhotos(ctx, item.ID, len(photoURLs))
		return nil, err
	}

	if len(photoURLs) > 0 {
		log.Debug(
			"offer with photos created successfully",
			slog.String("offer_id", item.ID.String()),
			slog.Int("photo_count", len(photoURLs)),
		)
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
		if offers[i].AuthorId.String() == info.Id {
			offers[i].AuthorName = &info.Name
		}
	}

	return offers, universalCursor, nil
}

// GetOfferByID retrieves a single offer by its ID, including the author name.
//
// Errors:
//   - domain.ErrOfferNotFound
func (s *Service) GetOfferByID(ctx context.Context, id uuid.UUID) (*domain.Offer, error) {
	offer, err := s.repo.GetOfferByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response, err := s.usersClient.GetUsersWithInfo(ctx, &userspb.GetUsersWithInfoRequest{Ids: []string{offer.AuthorId.String()}})
	if err == nil && len(response.Users) > 0 && response.Users[0] != nil {
		offer.AuthorName = &response.Users[0].Name
	}

	return offer, nil
}

func (s *Service) cleanupUploadedPhotos(ctx context.Context, offerID uuid.UUID, count int) {
	if s.photoStorage == nil || count == 0 {
		return
	}

	log := logger.LogFrom(ctx, s.fallbackLogger).With(slog.String("offer_id", offerID.String()))
	for i := 0; i < count; i++ {
		if err := s.photoStorage.DeletePhoto(ctx, offerID, i); err != nil {
			log.Warn("failed to cleanup uploaded offer photo", slog.Int("photo_index", i), slog.Any("error", err))
		}
	}
}
