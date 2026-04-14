package offers

import (
	userspb "barter-port/contracts/grpc/users/v1"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	offersrep "barter-port/internal/deals/infrastructure/repository/offers"
	"barter-port/pkg/authkit"
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
var ErrOfferPhotoLimitExceeded = errors.New("offer photo limit exceeded")
var ErrOfferPhotoNotFound = errors.New("offer photo not found")

const maxOfferPhotoCount = 10

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
	adminChecker   *authkit.AdminChecker
	fallbackLogger *slog.Logger
}

func NewService(
	dbPool *pgxpool.Pool,
	offerRepository *offersrep.Repository,
	usersClient userspb.UsersServiceClient,
	photoStorage PhotoStorage,
	adminChecker *authkit.AdminChecker,
	fallbackLogger *slog.Logger,
) *Service {
	return &Service{
		db:             dbPool,
		repo:           offerRepository,
		photoStorage:   photoStorage,
		usersClient:    usersClient,
		adminChecker:   adminChecker,
		fallbackLogger: fallbackLogger,
	}
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

	uploadedPhotos := make([]domain.OfferPhoto, 0, len(photos))
	uploadedPositions := make([]int, 0, len(photos))
	for i, photo := range photos {
		photoURL, err := s.photoStorage.UploadPhoto(ctx, item.ID, i, photo.ContentType, photo.Content)
		if err != nil {
			s.cleanupUploadedPhotos(ctx, item.ID, uploadedPositions)
			return nil, err
		}

		uploadedPhotos = append(uploadedPhotos, domain.OfferPhoto{
			ID:       uuid.New(),
			OfferID:  item.ID,
			URL:      photoURL,
			Position: i,
		})
		uploadedPositions = append(uploadedPositions, i)
		log.Debug(
			"offer photo uploaded successfully",
			slog.String("offer_id", item.ID.String()),
			slog.Int("photo_index", i),
			slog.String("content_type", photo.ContentType),
			slog.Int("size_bytes", len(photo.Content)),
			slog.String("photo_url", photoURL),
		)
	}
	for _, photo := range uploadedPhotos {
		item.PhotoIds = append(item.PhotoIds, photo.ID)
		item.PhotoUrls = append(item.PhotoUrls, photo.URL)
	}

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.repo.AddOffer(ctx, tx, item); err != nil {
			return err
		}
		return s.repo.AddOfferPhotos(ctx, tx, uploadedPhotos)
	})
	if err != nil {
		s.cleanupUploadedPhotos(ctx, item.ID, uploadedPositions)
		return nil, err
	}

	if len(uploadedPhotos) > 0 {
		log.Debug(
			"offer with photos created successfully",
			slog.String("offer_id", item.ID.String()),
			slog.Int("photo_count", len(uploadedPhotos)),
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
	requesterID *uuid.UUID,
) ([]domain.Offer, *domain.UniversalCursor, error) {

	log := logger.LogFrom(ctx, s.fallbackLogger)

	var universalCursor *domain.UniversalCursor
	var offers []domain.Offer
	isAdmin := false
	if requesterID != nil {
		var err error
		isAdmin, err = s.adminChecker.IsAdmin(ctx, *requesterID)
		if err != nil {
			return nil, nil, fmt.Errorf("check admin: %w", err)
		}
	}

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

		offers, timeCursor, err = s.repo.GetOffersOrderByTimeForRequester(ctx, timeCursor, limit, authorID, requesterID, isAdmin)
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

		offers, popularityCursor, err = s.repo.GetOffersOrderByPopularityForRequester(ctx, popularityCursor, limit, authorID, requesterID, isAdmin)
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
// Returns ErrOfferNotFound if the offer is hidden and the requester is not the author.
//
// Errors:
//   - domain.ErrOfferNotFound
func (s *Service) GetOfferByID(ctx context.Context, id uuid.UUID, requesterID *uuid.UUID) (*domain.Offer, error) {
	offer, err := s.repo.GetOfferByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if offer.IsHidden {
		if requesterID == nil {
			return nil, domain.ErrOfferNotFound
		}
		if offer.AuthorId != *requesterID {
			isAdmin, adminErr := s.adminChecker.IsAdmin(ctx, *requesterID)
			if adminErr != nil {
				return nil, fmt.Errorf("check admin: %w", adminErr)
			}
			if !isAdmin {
				return nil, domain.ErrOfferNotFound
			}
		}
	}

	response, err := s.usersClient.GetUsersWithInfo(ctx, &userspb.GetUsersWithInfoRequest{Ids: []string{offer.AuthorId.String()}})
	if err == nil && len(response.Users) > 0 && response.Users[0] != nil {
		offer.AuthorName = &response.Users[0].Name
	}

	return offer, nil
}

// ViewOfferByID increments offer views.
//
// Errors:
//   - domain.ErrOfferNotFound
func (s *Service) ViewOfferByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.ViewOffer(ctx, s.db, id)
}

// UpdateOffer updates an offer. Only the author can update it.
//
// Domain errors:
//   - domain.ErrOfferNotFound
//   - domain.ErrForbidden
//   - domain.ErrInvalidOfferName
func (s *Service) UpdateOffer(
	ctx context.Context,
	userID uuid.UUID,
	offerID uuid.UUID,
	patch htypes.OfferPatch,
	newPhotos []PhotoUpload,
) (*domain.Offer, error) {
	if patch.Name != nil && *patch.Name == "" {
		return nil, domain.ErrInvalidOfferName
	}
	if len(newPhotos) > 0 && s.photoStorage == nil {
		return nil, ErrOfferPhotoStorageNotConfigured
	}

	var (
		offer             *domain.Offer
		deletedPhotos     []domain.OfferPhoto
		uploadedPositions []int
	)

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		currentOffer, err := s.repo.GetOffer(ctx, tx, offerID)
		if err != nil {
			return err
		}
		if currentOffer.AuthorId != userID {
			return domain.ErrForbidden
		}
		if currentOffer.ModificationBlocked {
			return domain.ErrModificationBlocked
		}

		currentPhotos, err := s.repo.GetOfferPhotos(ctx, tx, offerID)
		if err != nil {
			return err
		}

		photosByID := make(map[uuid.UUID]domain.OfferPhoto, len(currentPhotos))
		for _, photo := range currentPhotos {
			photosByID[photo.ID] = photo
		}

		deleteIDs := make([]uuid.UUID, 0, len(patch.DeletePhotoIds))
		seenDeleteIDs := make(map[uuid.UUID]struct{}, len(patch.DeletePhotoIds))
		for _, photoID := range patch.DeletePhotoIds {
			if _, seen := seenDeleteIDs[photoID]; seen {
				continue
			}
			photo, ok := photosByID[photoID]
			if !ok {
				return ErrOfferPhotoNotFound
			}
			seenDeleteIDs[photoID] = struct{}{}
			deleteIDs = append(deleteIDs, photoID)
			deletedPhotos = append(deletedPhotos, photo)
		}

		remainingCount := len(currentPhotos) - len(deleteIDs) + len(newPhotos)
		if remainingCount > maxOfferPhotoCount {
			return ErrOfferPhotoLimitExceeded
		}

		nextPosition := 0
		for _, photo := range currentPhotos {
			if photo.Position >= nextPosition {
				nextPosition = photo.Position + 1
			}
		}

		uploadedPhotos := make([]domain.OfferPhoto, 0, len(newPhotos))
		for i, photo := range newPhotos {
			position := nextPosition + i
			photoURL, uploadErr := s.photoStorage.UploadPhoto(ctx, offerID, position, photo.ContentType, photo.Content)
			if uploadErr != nil {
				return uploadErr
			}

			uploadedPositions = append(uploadedPositions, position)
			uploadedPhotos = append(uploadedPhotos, domain.OfferPhoto{
				ID:       uuid.New(),
				OfferID:  offerID,
				URL:      photoURL,
				Position: position,
			})
		}

		if err = s.repo.DeleteOfferPhotos(ctx, tx, offerID, deleteIDs); err != nil {
			return err
		}
		if err = s.repo.AddOfferPhotos(ctx, tx, uploadedPhotos); err != nil {
			return err
		}

		updatedOffer, err := s.repo.UpdateOffer(ctx, tx, offerID, userID, patch)
		if err != nil {
			return err
		}

		offer = &updatedOffer
		return nil
	})
	if err != nil {
		s.cleanupUploadedPhotos(ctx, offerID, uploadedPositions)
		return nil, err
	}

	if s.photoStorage != nil {
		for _, photo := range deletedPhotos {
			if err = s.photoStorage.DeletePhoto(ctx, offerID, photo.Position); err != nil {
				logger.LogFrom(ctx, s.fallbackLogger).Warn(
					"failed to delete offer photo from storage after update",
					slog.String("offer_id", offerID.String()),
					slog.String("photo_id", photo.ID.String()),
					slog.Int("photo_position", photo.Position),
					slog.Any("error", err),
				)
			}
		}
	}

	response, err := s.usersClient.GetUsersWithInfo(ctx, &userspb.GetUsersWithInfoRequest{Ids: []string{offer.AuthorId.String()}})
	if err == nil && len(response.Users) > 0 && response.Users[0] != nil {
		offer.AuthorName = &response.Users[0].Name
	}

	return offer, nil
}

// DeleteOffer deletes an offer. Only the author can delete it.
//
// Domain errors:
//   - domain.ErrOfferNotFound
//   - domain.ErrForbidden
//   - domain.ErrModificationBlocked
func (s *Service) DeleteOffer(ctx context.Context, userID uuid.UUID, offerID uuid.UUID) error {
	offer, err := s.repo.GetOffer(ctx, s.db, offerID)
	if err != nil {
		return err
	}
	if offer.AuthorId != userID {
		return domain.ErrForbidden
	}
	if offer.ModificationBlocked {
		return domain.ErrModificationBlocked
	}
	return s.repo.DeleteOffer(ctx, s.db, offerID, userID)
}

func (s *Service) cleanupUploadedPhotos(ctx context.Context, offerID uuid.UUID, positions []int) {
	if s.photoStorage == nil || len(positions) == 0 {
		return
	}

	log := logger.LogFrom(ctx, s.fallbackLogger).With(slog.String("offer_id", offerID.String()))
	for _, position := range positions {
		if err := s.photoStorage.DeletePhoto(ctx, offerID, position); err != nil {
			log.Warn("failed to cleanup uploaded offer photo", slog.Int("photo_index", position), slog.Any("error", err))
		}
	}
}
