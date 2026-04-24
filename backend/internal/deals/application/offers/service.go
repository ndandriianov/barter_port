package offers

import (
	userspb "barter-port/contracts/grpc/users/v1"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	offersrep "barter-port/internal/deals/infrastructure/repository/offers"
	"barter-port/pkg/authkit"
	"barter-port/pkg/db"
	"barter-port/pkg/geo"
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
	tags []string,
	photos []PhotoUpload,
	latitude *float64,
	longitude *float64,
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
		Tags:        append([]string(nil), tags...),
		Type:        itemType,
		Action:      action,
		Description: description,
		Latitude:    latitude,
		Longitude:   longitude,
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
		if err := s.repo.ReplaceOfferTags(ctx, tx, item.ID, tags); err != nil {
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

// ================================================================================
// ПОЛУЧИТЬ СПИСОК OFFERS
// ================================================================================

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
	tagFilter *[]string,
	requestLocation *domain.Location,
) ([]domain.Offer, *domain.UniversalCursor, error) {

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
		timeCursor, err := getTimeCursor(cursor)
		if err != nil {
			return nil, nil, err
		}

		offers, newTimeCursor, err := s.getOffersByTime(ctx, timeCursor, limit, authorID, isAdmin, tagFilter)
		if err != nil {
			return nil, nil, err
		}

		var newUniversalCursor *domain.UniversalCursor
		if newTimeCursor != nil {
			newUniversalCursor = newTimeCursor.ToUniversalCursor()
		}

		offers, err = s.addAuthorNameToOffers(ctx, offers)
		if err != nil {
			return nil, nil, err
		}
		if requesterID != nil {
			offers, err = s.addFavoriteFlagToOffers(ctx, *requesterID, offers)
			if err != nil {
				return nil, nil, err
			}
			offers, err = s.addDistanceToOffers(ctx, *requesterID, offers)
			if err != nil {
				return nil, nil, err
			}
		}

		return offers, newUniversalCursor, nil

	case enums.SortTypeByPopularity:
		popularityCursor, err := getPopularityCursor(cursor)
		if err != nil {
			return nil, nil, err
		}

		offers, newPopularityCursor, err := s.getOffersByPopularity(ctx, popularityCursor, limit, authorID, isAdmin, tagFilter)
		if err != nil {
			return nil, nil, err
		}

		var newUniversalCursor *domain.UniversalCursor
		if newPopularityCursor != nil {
			newUniversalCursor = newPopularityCursor.ToUniversalCursor()
		}

		offers, err = s.addAuthorNameToOffers(ctx, offers)
		if err != nil {
			return nil, nil, err
		}
		if requesterID != nil {
			offers, err = s.addFavoriteFlagToOffers(ctx, *requesterID, offers)
			if err != nil {
				return nil, nil, err
			}
			offers, err = s.addDistanceToOffers(ctx, *requesterID, offers)
			if err != nil {
				return nil, nil, err
			}
		}

		return offers, newUniversalCursor, nil

	case enums.SortTypeByDistance:
		if requestLocation == nil {
			return nil, nil, fmt.Errorf("distance sort requires request location")
		}

		distanceCursor, err := getDistanceCursor(cursor)
		if err != nil {
			return nil, nil, err
		}

		offers, newDistanceCursor, err := s.getOffersByDistance(ctx, distanceCursor, limit, authorID, isAdmin, tagFilter, *requestLocation)
		if err != nil {
			return nil, nil, err
		}

		var newUniversalCursor *domain.UniversalCursor
		if newDistanceCursor != nil {
			newUniversalCursor = newDistanceCursor.ToUniversalCursor()
		}

		offers, err = s.addAuthorNameToOffers(ctx, offers)
		if err != nil {
			return nil, nil, err
		}
		if requesterID != nil {
			offers, err = s.addFavoriteFlagToOffers(ctx, *requesterID, offers)
			if err != nil {
				return nil, nil, err
			}
		}

		return offers, newUniversalCursor, nil

	default:
		return nil, nil, fmt.Errorf("invalid sort type: %v", sortType)
	}
}

func (s *Service) GetSubscribedOffers(
	ctx context.Context,
	requesterID uuid.UUID,
	sortType enums.SortType,
	cursor *domain.UniversalCursor,
	limit int,
	requestLocation *domain.Location,
) ([]domain.Offer, *domain.UniversalCursor, error) {
	isAdmin, err := s.adminChecker.IsAdmin(ctx, requesterID)
	if err != nil {
		return nil, nil, fmt.Errorf("check admin: %w", err)
	}

	subscriptionsResponse, err := s.usersClient.ListSubscriptions(ctx, &userspb.ListSubscriptionsRequest{
		UserId: requesterID.String(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("list subscriptions: %w", err)
	}

	authorIDs := make([]uuid.UUID, 0, len(subscriptionsResponse.Subscriptions))
	for _, subscription := range subscriptionsResponse.Subscriptions {
		if subscription == nil {
			continue
		}

		authorID, parseErr := uuid.Parse(subscription.Id)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("parse subscription id %q: %w", subscription.Id, parseErr)
		}
		authorIDs = append(authorIDs, authorID)
	}

	if len(authorIDs) == 0 {
		return []domain.Offer{}, nil, nil
	}

	switch sortType {
	case enums.SortTypeByTime:
		timeCursor, err := getTimeCursor(cursor)
		if err != nil {
			return nil, nil, err
		}

		offers, newTimeCursor, err := s.getSubscribedOffersByTime(ctx, timeCursor, limit, authorIDs, isAdmin)
		if err != nil {
			return nil, nil, err
		}

		var newUniversalCursor *domain.UniversalCursor
		if newTimeCursor != nil {
			newUniversalCursor = newTimeCursor.ToUniversalCursor()
		}

		offers, err = s.addAuthorNameToOffers(ctx, offers)
		if err != nil {
			return nil, nil, err
		}
		offers, err = s.addFavoriteFlagToOffers(ctx, requesterID, offers)
		if err != nil {
			return nil, nil, err
		}
		offers, err = s.addDistanceToOffers(ctx, requesterID, offers)
		if err != nil {
			return nil, nil, err
		}

		return offers, newUniversalCursor, nil

	case enums.SortTypeByPopularity:
		popularityCursor, err := getPopularityCursor(cursor)
		if err != nil {
			return nil, nil, err
		}

		offers, newPopularityCursor, err := s.getSubscribedOffersByPopularity(ctx, popularityCursor, limit, authorIDs, isAdmin)
		if err != nil {
			return nil, nil, err
		}

		var newUniversalCursor *domain.UniversalCursor
		if newPopularityCursor != nil {
			newUniversalCursor = newPopularityCursor.ToUniversalCursor()
		}

		offers, err = s.addAuthorNameToOffers(ctx, offers)
		if err != nil {
			return nil, nil, err
		}
		offers, err = s.addFavoriteFlagToOffers(ctx, requesterID, offers)
		if err != nil {
			return nil, nil, err
		}
		offers, err = s.addDistanceToOffers(ctx, requesterID, offers)
		if err != nil {
			return nil, nil, err
		}

		return offers, newUniversalCursor, nil

	case enums.SortTypeByDistance:
		if requestLocation == nil {
			return nil, nil, fmt.Errorf("distance sort requires request location")
		}

		distanceCursor, err := getDistanceCursor(cursor)
		if err != nil {
			return nil, nil, err
		}

		offers, newDistanceCursor, err := s.getSubscribedOffersByDistance(ctx, distanceCursor, limit, authorIDs, isAdmin, *requestLocation)
		if err != nil {
			return nil, nil, err
		}

		var newUniversalCursor *domain.UniversalCursor
		if newDistanceCursor != nil {
			newUniversalCursor = newDistanceCursor.ToUniversalCursor()
		}

		offers, err = s.addAuthorNameToOffers(ctx, offers)
		if err != nil {
			return nil, nil, err
		}
		offers, err = s.addFavoriteFlagToOffers(ctx, requesterID, offers)
		if err != nil {
			return nil, nil, err
		}

		return offers, newUniversalCursor, nil

	default:
		return nil, nil, fmt.Errorf("invalid sort type: %v", sortType)
	}
}

func getTimeCursor(cursor *domain.UniversalCursor) (*domain.TimeCursor, error) {
	if cursor == nil {
		return nil, nil
	}
	return cursor.ToTimeCursor()
}

func (s *Service) getOffersByTime(
	ctx context.Context,
	cursor *domain.TimeCursor,
	limit int,
	authorID *uuid.UUID,
	isAdmin bool,
	tagFilter *[]string,
) ([]domain.Offer, *domain.TimeCursor, error) {
	tags, tagsFilterPresent := extractTagFilter(tagFilter)

	switch { // наличие курсора
	case cursor != nil:
		switch { // my
		case authorID != nil:
			return s.repo.GetMyOffersOrderByTime(ctx, *cursor, *authorID, limit, tags, tagsFilterPresent)
		default:
			return s.repo.GetOffersOrderByTime(ctx, limit, *cursor, isAdmin, tags, tagsFilterPresent)
		}
	default:
		switch { // my
		case authorID != nil:
			return s.repo.GetMyOffersOrderByTimeNoCursor(ctx, *authorID, limit, tags, tagsFilterPresent)
		default:
			return s.repo.GetOffersOrderByTimeNoCursor(ctx, limit, isAdmin, tags, tagsFilterPresent)
		}
	}
}

func (s *Service) getSubscribedOffersByTime(
	ctx context.Context,
	cursor *domain.TimeCursor,
	limit int,
	authorIDs []uuid.UUID,
	isAdmin bool,
) ([]domain.Offer, *domain.TimeCursor, error) {
	if cursor != nil {
		return s.repo.GetSubscribedOffersOrderByTime(ctx, limit, *cursor, authorIDs, isAdmin)
	}

	return s.repo.GetSubscribedOffersOrderByTimeNoCursor(ctx, limit, authorIDs, isAdmin)
}

func getPopularityCursor(cursor *domain.UniversalCursor) (*domain.PopularityCursor, error) {
	if cursor == nil {
		return nil, nil
	}
	return cursor.ToPopularityCursor()
}

func getDistanceCursor(cursor *domain.UniversalCursor) (*domain.DistanceCursor, error) {
	if cursor == nil {
		return nil, nil
	}
	return cursor.ToDistanceCursor()
}

func (s *Service) getOffersByPopularity(
	ctx context.Context,
	cursor *domain.PopularityCursor,
	limit int,
	authorID *uuid.UUID,
	isAdmin bool,
	tagFilter *[]string,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	tags, tagsFilterPresent := extractTagFilter(tagFilter)

	switch { // наличие курсора
	case cursor != nil:
		switch { // my
		case authorID != nil:
			return s.repo.GetMyOffersOrderByPopularity(ctx, limit, *cursor, *authorID, tags, tagsFilterPresent)
		default:
			return s.repo.GetOffersOrderByPopularity(ctx, limit, *cursor, isAdmin, tags, tagsFilterPresent)
		}
	default:
		switch { // my
		case authorID != nil:
			return s.repo.GetMyOffersOrderByPopularityNoCursor(ctx, limit, *authorID, tags, tagsFilterPresent)
		default:
			return s.repo.GetOffersOrderByPopularityNoCursor(ctx, limit, isAdmin, tags, tagsFilterPresent)
		}
	}
}

func (s *Service) getSubscribedOffersByPopularity(
	ctx context.Context,
	cursor *domain.PopularityCursor,
	limit int,
	authorIDs []uuid.UUID,
	isAdmin bool,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	if cursor != nil {
		return s.repo.GetSubscribedOffersOrderByPopularity(ctx, limit, *cursor, authorIDs, isAdmin)
	}

	return s.repo.GetSubscribedOffersOrderByPopularityNoCursor(ctx, limit, authorIDs, isAdmin)
}

func (s *Service) getOffersByDistance(
	ctx context.Context,
	cursor *domain.DistanceCursor,
	limit int,
	authorID *uuid.UUID,
	isAdmin bool,
	tagFilter *[]string,
	location domain.Location,
) ([]domain.Offer, *domain.DistanceCursor, error) {
	tags, tagsFilterPresent := extractTagFilter(tagFilter)

	switch {
	case cursor != nil:
		switch {
		case authorID != nil:
			return s.repo.GetMyOffersOrderByDistance(ctx, limit, *cursor, *authorID, location.Lat, location.Lon, tags, tagsFilterPresent)
		default:
			return s.repo.GetOffersOrderByDistance(ctx, limit, *cursor, isAdmin, location.Lat, location.Lon, tags, tagsFilterPresent)
		}
	default:
		switch {
		case authorID != nil:
			return s.repo.GetMyOffersOrderByDistanceNoCursor(ctx, limit, *authorID, location.Lat, location.Lon, tags, tagsFilterPresent)
		default:
			return s.repo.GetOffersOrderByDistanceNoCursor(ctx, limit, isAdmin, location.Lat, location.Lon, tags, tagsFilterPresent)
		}
	}
}

func (s *Service) getSubscribedOffersByDistance(
	ctx context.Context,
	cursor *domain.DistanceCursor,
	limit int,
	authorIDs []uuid.UUID,
	isAdmin bool,
	location domain.Location,
) ([]domain.Offer, *domain.DistanceCursor, error) {
	if cursor != nil {
		return s.repo.GetSubscribedOffersOrderByDistance(ctx, limit, *cursor, authorIDs, isAdmin, location.Lat, location.Lon)
	}

	return s.repo.GetSubscribedOffersOrderByDistanceNoCursor(ctx, limit, authorIDs, isAdmin, location.Lat, location.Lon)
}

// addDistanceToOffers populates DistanceMeters on each offer that has coordinates, using the
// requester's saved location from the users service. Silently skips if requester has no location.
func (s *Service) addDistanceToOffers(ctx context.Context, requesterID uuid.UUID, offers []domain.Offer) ([]domain.Offer, error) {
	resp, err := s.usersClient.GetUserLocation(ctx, &userspb.GetUserLocationRequest{UserId: requesterID.String()})
	if err != nil {
		return offers, nil // best-effort: no distance if gRPC fails
	}
	if resp.Latitude == nil || resp.Longitude == nil {
		return offers, nil
	}

	userLat := resp.GetLatitude()
	userLon := resp.GetLongitude()

	for i := range offers {
		o := &offers[i]
		if o.Latitude == nil || o.Longitude == nil {
			continue
		}
		d := geo.HaversineDistance(userLat, userLon, *o.Latitude, *o.Longitude)
		o.DistanceMeters = &d
	}

	return offers, nil
}

func (s *Service) addDistanceToFavoritedOffers(
	ctx context.Context,
	requesterID uuid.UUID,
	offers []domain.FavoritedOffer,
) ([]domain.FavoritedOffer, error) {
	resp, err := s.usersClient.GetUserLocation(ctx, &userspb.GetUserLocationRequest{UserId: requesterID.String()})
	if err != nil {
		return offers, nil // best-effort: no distance if gRPC fails
	}
	if resp.Latitude == nil || resp.Longitude == nil {
		return offers, nil
	}

	userLat := resp.GetLatitude()
	userLon := resp.GetLongitude()

	for i := range offers {
		o := &offers[i]
		if o.Latitude == nil || o.Longitude == nil {
			continue
		}
		d := geo.HaversineDistance(userLat, userLon, *o.Latitude, *o.Longitude)
		o.DistanceMeters = &d
	}

	return offers, nil
}

func (s *Service) addAuthorNameToOffers(ctx context.Context, offers []domain.Offer) ([]domain.Offer, error) {
	if len(offers) == 0 {
		return offers, nil
	}

	ids := make([]string, len(offers))
	for i, o := range offers {
		ids[i] = o.AuthorId.String()
	}

	response, err := s.usersClient.GetUsersWithInfo(ctx, &userspb.GetUsersWithInfoRequest{Ids: ids})
	if err != nil {
		return nil, fmt.Errorf("failed to get author names: %w", err)
	}

	for i, info := range response.Users {
		if info == nil {
			continue // буду считать что пользователь с неуказанным именем
		}
		if offers[i].AuthorId.String() == info.Id {
			offers[i].AuthorName = &info.Name
		}
	}

	return offers, nil
}

func (s *Service) addAuthorNameToFavoritedOffers(ctx context.Context, offers []domain.FavoritedOffer) ([]domain.FavoritedOffer, error) {
	if len(offers) == 0 {
		return offers, nil
	}

	ids := make([]string, len(offers))
	for i, offer := range offers {
		ids[i] = offer.AuthorId.String()
	}

	response, err := s.usersClient.GetUsersWithInfo(ctx, &userspb.GetUsersWithInfoRequest{Ids: ids})
	if err != nil {
		return nil, fmt.Errorf("failed to get author names: %w", err)
	}

	for i, info := range response.Users {
		if info == nil {
			continue
		}
		if offers[i].AuthorId.String() == info.Id {
			offers[i].AuthorName = &info.Name
		}
	}

	return offers, nil
}

func (s *Service) addFavoriteFlagToOffers(ctx context.Context, userID uuid.UUID, offers []domain.Offer) ([]domain.Offer, error) {
	if len(offers) == 0 {
		return offers, nil
	}

	ids := make([]uuid.UUID, len(offers))
	for i, offer := range offers {
		ids[i] = offer.ID
	}

	favoriteIDs, err := s.repo.GetFavoriteOfferIDs(ctx, userID, ids)
	if err != nil {
		return nil, fmt.Errorf("get favorite offer ids: %w", err)
	}

	for i := range offers {
		offers[i].IsFavorite = new(favoriteIDs[offers[i].ID])
	}

	return offers, nil
}

func (s *Service) ensureOfferVisible(ctx context.Context, offer *domain.Offer, requesterID *uuid.UUID) error {
	if !offer.IsHidden {
		return nil
	}
	if requesterID == nil {
		return domain.ErrOfferNotFound
	}
	if offer.AuthorId == *requesterID {
		return nil
	}

	isAdmin, err := s.adminChecker.IsAdmin(ctx, *requesterID)
	if err != nil {
		return fmt.Errorf("check admin: %w", err)
	}
	if !isAdmin {
		return domain.ErrOfferNotFound
	}

	return nil
}

// ================================================================================
// КОНЕЦ СЕКЦИИ получить список offers
// ================================================================================

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

	if err := s.ensureOfferVisible(ctx, offer, requesterID); err != nil {
		return nil, err
	}

	response, err := s.usersClient.GetUsersWithInfo(ctx, &userspb.GetUsersWithInfoRequest{Ids: []string{offer.AuthorId.String()}})
	if err == nil && len(response.Users) > 0 && response.Users[0] != nil {
		offer.AuthorName = &response.Users[0].Name
	}
	if requesterID != nil {
		offersWithFavoriteFlag, favoriteErr := s.addFavoriteFlagToOffers(ctx, *requesterID, []domain.Offer{*offer})
		if favoriteErr != nil {
			return nil, favoriteErr
		}
		offer = &offersWithFavoriteFlag[0]

		offerSlice, distErr := s.addDistanceToOffers(ctx, *requesterID, []domain.Offer{*offer})
		if distErr != nil {
			return nil, distErr
		}
		offer = &offerSlice[0]
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

func (s *Service) AddOfferToFavorites(ctx context.Context, userID uuid.UUID, offerID uuid.UUID) error {
	offer, err := s.repo.GetOfferByID(ctx, offerID)
	if err != nil {
		return err
	}
	if err := s.ensureOfferVisible(ctx, offer, &userID); err != nil {
		return err
	}

	return s.repo.AddOfferToFavorites(ctx, s.db, userID, offerID)
}

func (s *Service) RemoveOfferFromFavorites(ctx context.Context, userID uuid.UUID, offerID uuid.UUID) error {
	return s.repo.RemoveOfferFromFavorites(ctx, s.db, userID, offerID)
}

func (s *Service) GetFavoriteOffers(
	ctx context.Context,
	userID uuid.UUID,
	cursor *domain.FavoriteOffersCursor,
	limit int,
) ([]domain.FavoritedOffer, *domain.FavoriteOffersCursor, error) {
	isAdmin, err := s.adminChecker.IsAdmin(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("check admin: %w", err)
	}

	var offers []domain.FavoritedOffer
	var nextCursor *domain.FavoriteOffersCursor
	if cursor != nil {
		offers, nextCursor, err = s.repo.GetFavoriteOffers(ctx, userID, *cursor, limit, isAdmin)
	} else {
		offers, nextCursor, err = s.repo.GetFavoriteOffersNoCursor(ctx, userID, limit, isAdmin)
	}
	if err != nil {
		return nil, nil, err
	}

	offers, err = s.addAuthorNameToFavoritedOffers(ctx, offers)
	if err != nil {
		return nil, nil, err
	}

	for i := range offers {
		offers[i].IsFavorite = new(true)
	}

	offers, err = s.addDistanceToFavoritedOffers(ctx, userID, offers)
	if err != nil {
		return nil, nil, err
	}

	return offers, nextCursor, nil
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
		if patch.Tags != nil {
			if err = s.repo.ReplaceOfferTags(ctx, tx, offerID, *patch.Tags); err != nil {
				return err
			}
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
	return db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		offer, err := s.repo.GetOffer(ctx, tx, offerID)
		if err != nil {
			return err
		}
		if offer.AuthorId != userID {
			return domain.ErrForbidden
		}
		if offer.ModificationBlocked {
			return domain.ErrModificationBlocked
		}
		if err := s.repo.DeleteOffer(ctx, tx, offerID, userID); err != nil {
			return err
		}
		return s.repo.DeleteUnusedTags(ctx, tx)
	})
}

func (s *Service) ListTags(ctx context.Context) ([]string, error) {
	return s.repo.ListTags(ctx)
}

func (s *Service) DeleteTag(ctx context.Context, requesterID uuid.UUID, rawName string) error {
	normalized, err := domain.NormalizeTag(rawName)
	if err != nil {
		return err
	}

	isAdmin, err := s.adminChecker.IsAdmin(ctx, requesterID)
	if err != nil {
		return fmt.Errorf("check admin: %w", err)
	}
	if !isAdmin {
		return domain.ErrAdminOnly
	}

	return s.repo.DeleteTagByName(ctx, s.db, normalized)
}

func extractTagFilter(tagFilter *[]string) ([]string, bool) {
	if tagFilter == nil {
		return nil, false
	}

	result := make([]string, len(*tagFilter))
	copy(result, *tagFilter)
	return result, true
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
