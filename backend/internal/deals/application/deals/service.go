package deals

import (
	chatspb "barter-port/contracts/grpc/chats/v1"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/internal/deals/infrastructure/repository/deals"
	"barter-port/internal/deals/infrastructure/repository/drafts"
	failuresrepo "barter-port/internal/deals/infrastructure/repository/failures"
	"barter-port/internal/deals/infrastructure/repository/joins"
	"barter-port/internal/deals/infrastructure/repository/offers"
	"barter-port/pkg/authkit"
	"barter-port/pkg/db"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemPhotoStorage interface {
	CopyPhoto(ctx context.Context, sourceURL string, itemID uuid.UUID, index int) (string, error)
	DeletePhoto(ctx context.Context, itemID uuid.UUID, index int) error
}

type Service struct {
	db               *pgxpool.Pool
	draftsRepository *drafts.Repository
	dealsRepository  *deals.Repository
	failuresRepo     *failuresrepo.Repository
	joinsRepository  *joins.Repository
	offersRepository *offers.Repository
	itemPhotoStorage ItemPhotoStorage
	chatsClient      chatspb.ChatsServiceClient
	adminChecker     *authkit.AdminChecker
	logger           *slog.Logger
}

func NewService(
	db *pgxpool.Pool,
	draftsRepo *drafts.Repository,
	dealsRepo *deals.Repository,
	failuresRepo *failuresrepo.Repository,
	joinsRepo *joins.Repository,
	offersRepo *offers.Repository,
	itemPhotoStorage ItemPhotoStorage,
) *Service {
	return &Service{
		db:               db,
		draftsRepository: draftsRepo,
		dealsRepository:  dealsRepo,
		failuresRepo:     failuresRepo,
		joinsRepository:  joinsRepo,
		offersRepository: offersRepo,
		itemPhotoStorage: itemPhotoStorage,
		logger:           slog.Default(),
	}
}

func (s *Service) WithChatsClient(client chatspb.ChatsServiceClient) *Service {
	s.chatsClient = client
	return s
}

func (s *Service) WithAdminChecker(checker *authkit.AdminChecker) *Service {
	s.adminChecker = checker
	return s
}

func (s *Service) WithLogger(logger *slog.Logger) *Service {
	s.logger = logger
	return s
}

func (s *Service) DB() *pgxpool.Pool {
	return s.db
}

func (s *Service) DraftsRepository() *drafts.Repository {
	return s.draftsRepository
}

func (s *Service) DealsRepository() *deals.Repository {
	return s.dealsRepository
}

func (s *Service) FailuresRepository() *failuresrepo.Repository {
	return s.failuresRepo
}

func (s *Service) JoinsRepository() *joins.Repository {
	return s.joinsRepository
}

func (s *Service) OffersRepository() *offers.Repository {
	return s.offersRepository
}

func (s *Service) ChatsClient() chatspb.ChatsServiceClient {
	return s.chatsClient
}

func (s *Service) AdminChecker() *authkit.AdminChecker {
	return s.adminChecker
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

// ================================================================================
// CREATE DRAFT
// ================================================================================

// CreateDraft inserts a new draft deal into the database and returns its ID.
//
// Errors:
//   - domain.ErrNoOffers: if the items list is empty.
func (s *Service) CreateDraft(
	ctx context.Context,
	authorID uuid.UUID,
	name *string,
	description *string,
	offersList []domain.OfferIDAndInfo,
) (uuid.UUID, error) {
	if len(offersList) == 0 {
		return uuid.Nil, domain.ErrNoOffers
	}

	if name == nil {
		offerIDs := make([]uuid.UUID, len(offersList))
		for i, o := range offersList {
			offerIDs[i] = o.ID
		}

		names, err := s.offersRepository.GetOfferNamesByIDs(ctx, s.db, offerIDs)
		if err != nil {
			return uuid.Nil, fmt.Errorf("get offer names for draft name: %w", err)
		}

		name = new(strings.Join(names, ", "))
	}

	var id uuid.UUID
	var err error

	txErr := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		id, err = s.draftsRepository.CreateDraft(ctx, tx, authorID, name, description, offersList)
		return err
	})

	return id, txErr
}

// ================================================================================
// GET DRAFT IDS BY AUTHOR
// ================================================================================

// GetDraftsByAuthor returns a list of draft deal IDs created by the specified author.
//
// No domain errors
func (s *Service) GetDraftsByAuthor(
	ctx context.Context,
	authorID uuid.UUID,
	createdByMe bool,
) ([]htypes.DraftIDWithAuthorIDs, error) {
	return s.draftsRepository.GetDraftsByAuthor(ctx, s.db, authorID, createdByMe)
}

// ================================================================================
// GET DRAFT BY ID
// ================================================================================

// GetDraftByID returns a draft deal by its ID.
//
// Domain errors:
// - domain.ErrDraftNotFound: if no draft deal with the specified ID exists.
func (s *Service) GetDraftByID(ctx context.Context, id uuid.UUID) (domain.Draft, error) {
	return s.draftsRepository.GetDraftByID(ctx, s.db, id)
}

// DeleteDraftByID deletes a draft deal by its ID.
//
// Domain errors:
// - domain.ErrDraftNotFound: if no draft deal with the specified ID exists.
// - domain.ErrForbidden: if the user is not a participant of the draft.
func (s *Service) DeleteDraftByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		_, err := s.draftsRepository.GetDraftByID(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("get draft: %w", err)
		}

		isParticipant := false
		participants, err := s.draftsRepository.GetParticipants(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("get draft participants: %w", err)
		}

		for _, p := range participants {
			if p == userID {
				isParticipant = true
				break
			}
		}
		if !isParticipant {
			return domain.ErrForbidden
		}

		err = s.draftsRepository.DeleteDraft(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("delete draft: %w", err)
		}

		return nil
	})
}

// ================================================================================
// CONFIRM DRAFT
// ================================================================================

// ConfirmDraft allows a user to confirm their participation in a draft deal.
// If all users confirm, this creates a new deal based on draft
//
// Errors:
//   - domain.ErrDraftNotFound
//   - domain.ErrUserNotInDraft
func (s *Service) ConfirmDraft(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]htypes.UserConfirmed, error) {
	var users []htypes.UserConfirmed
	copiedPhotos := make([]copiedItemPhoto, 0)

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		err := s.draftsRepository.ConfirmDraftByID(ctx, tx, id, userID)
		if err != nil {
			return fmt.Errorf("could not confirm draft: %w", err)
		}

		users, err = s.draftsRepository.GetConfirms(ctx, tx, id)
		if err != nil {
			return err
		}

		ready := true
		for _, user := range users {
			if user.Confirmed == false {
				ready = false
			}
		}

		if ready {
			draft, err := s.draftsRepository.GetDraftByID(ctx, tx, id)
			if err != nil {
				return fmt.Errorf("could not find draft: %w", err)
			}

			id, copiedPhotos, err = s.createDeal(ctx, tx, draft)
			if err != nil {
				return fmt.Errorf("could not create deal: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		s.cleanupCopiedItemPhotos(ctx, copiedPhotos)
		return nil, err
	}

	return users, nil
}

// ================================================================================
// CANCEL DRAFT
// ================================================================================

// CancelDraft allows a user to cancel participation in a draft deal.
//
// Errors:
//   - domain.ErrDraftNotFound
//   - domain.ErrUserNotInDraft
func (s *Service) CancelDraft(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		err := s.draftsRepository.UnconfirmDraftByID(ctx, tx, id, userID)
		if err != nil {
			return fmt.Errorf("could not cancel draft: %w", err)
		}

		return nil
	})
}

// ================================================================================
// GET DEALS
// ================================================================================

// GetDeals returns deal IDs with participant UUIDs.
// If my is true, filters to only deals the user participates in.
// If open is true, filters to deals that are not in a final status.
//
// No domain errors.
func (s *Service) GetDeals(ctx context.Context, userID uuid.UUID, my bool, open bool) ([]htypes.DealIDWithParticipantIDs, error) {
	var filterUserID *uuid.UUID
	if my {
		filterUserID = &userID
	}
	return s.dealsRepository.GetDealIDs(ctx, s.db, filterUserID, open)
}

// ================================================================================
// GET DEAL BY ID
// ================================================================================

// GetDealByID returns a deal by its ID.
//
// Domain errors:
//   - domain.ErrDealNotFound: if no deal with the specified ID exists.
func (s *Service) GetDealByID(ctx context.Context, id uuid.UUID) (domain.Deal, error) {
	var deal domain.Deal
	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		deal, err = s.dealsRepository.GetDealByID(ctx, tx, id)
		return err
	})
	if err != nil {
		return domain.Deal{}, err
	}

	return deal, nil
}

// GetDealStatus returns the current status of a deal.
//
// Domain errors:
//   - domain.ErrDealNotFound: if no deal with the specified ID exists.
func (s *Service) GetDealStatus(ctx context.Context, id uuid.UUID) (enums.DealStatus, error) {
	deal, err := s.GetDealByID(ctx, id)
	if err != nil {
		return 0, err
	}

	return deal.Status, nil
}

func (s *Service) HasPendingFailureReview(ctx context.Context, id uuid.UUID) (bool, error) {
	if _, err := s.GetDealByID(ctx, id); err != nil {
		return false, err
	}

	return s.failuresRepo.HasPendingFailureReview(ctx, s.db, id)
}

// ================================================================================
// UPDATE DEAL NAME
// ================================================================================

// UpdateDealName changes the name of an existing deal.
//
// Domain errors:
//   - domain.ErrDealNotFound: if no deal with the specified ID exists.
//   - domain.ErrForbidden: if the user is not a participant of the deal.
func (s *Service) UpdateDealName(ctx context.Context, dealID uuid.UUID, userID uuid.UUID, name string) (domain.Deal, error) {
	var deal domain.Deal

	txErr := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		deal, err = s.dealsRepository.GetDealByID(ctx, tx, dealID)
		if err != nil {
			return err
		}

		isParticipant := false
		for _, p := range deal.Participants {
			if p == userID {
				isParticipant = true
				break
			}
		}
		if !isParticipant {
			return domain.ErrForbidden
		}

		if err = s.dealsRepository.UpdateDealName(ctx, tx, dealID, name); err != nil {
			return err
		}

		deal, err = s.dealsRepository.GetDealByID(ctx, tx, dealID)
		return err
	})

	if txErr != nil {
		return domain.Deal{}, txErr
	}

	return deal, nil
}

// ================================================================================
// GET DEAL STATUS VOTES
// ================================================================================

// GetDealStatusVotes returns all votes currently recorded for the deal status transition.
//
// Domain errors:
//   - domain.ErrDealNotFound
func (s *Service) GetDealStatusVotes(ctx context.Context, dealID uuid.UUID) (map[uuid.UUID]enums.DealStatus, error) {
	var votes map[uuid.UUID]enums.DealStatus

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.dealsRepository.GetDealByID(ctx, tx, dealID); err != nil {
			return err
		}

		var err error
		votes, err = s.dealsRepository.GetStatusVotes(ctx, tx, dealID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return votes, nil
}

// ================================================================================
// ADD DEAL ITEM
// ================================================================================

// AddDealItem adds a new item to an existing deal based on the user's offer.
//
// Domain errors:
//   - domain.ErrDealNotFound
//   - domain.ErrOfferNotFound
//   - domain.ErrForbidden          - user is not a participant or tries to use another user's offer
//   - domain.ErrInvalidDealStatus  — deal is not in LookingForParticipants
//   - domain.ErrInvalidQuantity
func (s *Service) AddDealItem(
	ctx context.Context,
	userID uuid.UUID,
	dealID uuid.UUID,
	offerID uuid.UUID,
	quantity int,
) (domain.Deal, error) {
	if quantity < 1 {
		return domain.Deal{}, domain.ErrInvalidQuantity
	}

	var updatedDeal domain.Deal
	copiedPhotos := make([]copiedItemPhoto, 0)

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		deal, err := s.dealsRepository.GetDealByID(ctx, tx, dealID)
		if err != nil {
			return err
		}

		if deal.Status != enums.DealStatusLookingForParticipants {
			return domain.ErrInvalidDealStatus
		}

		if err = s.EnsureNoPendingFailureReview(ctx, tx, dealID); err != nil {
			return err
		}

		if !ContainsUserID(deal.Participants, userID) {
			return domain.ErrForbidden
		}

		offer, err := s.offersRepository.GetOffer(ctx, tx, offerID)
		if err != nil {
			return err
		}

		if offer.AuthorId != userID {
			return domain.ErrForbidden
		}

		newItem := domain.Item{
			ID:          uuid.New(),
			OfferID:     &offerID,
			AuthorID:    offer.AuthorId,
			Name:        offer.Name,
			Description: offer.Description,
			Type:        offer.Type,
			Quantity:    quantity,
		}

		if offer.Action == enums.OfferActionGive {
			newItem.ProviderID = &userID
		} else {
			newItem.ReceiverID = &userID
		}

		if _, err = s.dealsRepository.AddItem(ctx, tx, dealID, newItem); err != nil {
			return err
		}

		itemCopies, err := s.copyOfferPhotosToItem(ctx, tx, newItem.ID, offer.PhotoUrls)
		if err != nil {
			return err
		}
		copiedPhotos = append(copiedPhotos, itemCopies...)

		updatedDeal, err = s.dealsRepository.GetDealByID(ctx, tx, dealID)
		return err
	})
	if err != nil {
		s.cleanupCopiedItemPhotos(ctx, copiedPhotos)
		return domain.Deal{}, err
	}

	return updatedDeal, nil
}

// ================================================================================
// UPDATE DEAL ITEM
// ================================================================================

// UpdateDealItem applies a partial update to a deal item.
//
// Content fields (Name, Description, Quantity) may only be changed by the item author.
// Provider/receiver roles follow claim/release rules (see domain errors below).
//
// Domain errors:
//   - domain.ErrDealNotFound
//   - domain.ErrItemNotFound
//   - domain.ErrForbidden        — not the author (content change)
//   - domain.ErrRoleAlreadyTaken — slot occupied by another user (claim)
//   - domain.ErrNotRoleHolder    — user does not hold the role (release)
func (s *Service) UpdateDealItem(
	ctx context.Context,
	userID uuid.UUID,
	dealID uuid.UUID,
	itemID uuid.UUID,
	patch htypes.ItemPatch,
) (domain.Item, error) {
	var item domain.Item

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		deal, err := s.dealsRepository.GetDealByID(ctx, tx, dealID)
		if err != nil {
			return err
		}

		if deal.Status != enums.DealStatusLookingForParticipants && deal.Status != enums.DealStatusDiscussion {
			return domain.ErrInvalidDealStatus
		}

		if err := s.EnsureNoPendingFailureReview(ctx, tx, dealID); err != nil {
			return err
		}

		hasContent := patch.Name != nil || patch.Description != nil || patch.Quantity != nil
		if hasContent {
			var err error
			item, err = s.dealsRepository.UpdateItem(ctx, tx, dealID, itemID, userID, patch)
			if err != nil {
				return err
			}
		}

		if patch.ClaimProvider {
			id, err := s.dealsRepository.GetItemReceiverID(ctx, tx, dealID, itemID)
			if err != nil {
				return err
			}
			if id != nil && *id == userID {
				return domain.ErrDuplicateRole
			}

			item, err = s.dealsRepository.ClaimItemProvider(ctx, tx, dealID, itemID, userID)
			return err
		}

		if patch.ReleaseProvider {
			var err error
			item, err = s.dealsRepository.ReleaseItemProvider(ctx, tx, dealID, itemID, userID)
			if err != nil {
				return err
			}
		}

		if patch.ClaimReceiver {
			id, err := s.dealsRepository.GetItemProviderID(ctx, tx, dealID, itemID)
			if err != nil {
				return err
			}
			if id != nil && *id == userID {
				return domain.ErrDuplicateRole
			}

			item, err = s.dealsRepository.ClaimItemReceiver(ctx, tx, dealID, itemID, userID)
			return err
		}

		if patch.ReleaseReceiver {
			var err error
			item, err = s.dealsRepository.ReleaseItemReceiver(ctx, tx, dealID, itemID, userID)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return domain.Item{}, err
	}

	return item, nil
}

// ================================================================================
// PROCESS DEAL STATUS UPDATE
// ================================================================================

func (s *Service) ProcessDealStatusUpdateRequest(
	ctx context.Context,
	dealID uuid.UUID,
	userID uuid.UUID,
	status enums.DealStatus,
) (domain.Deal, error) {
	switch status {
	case enums.DealStatusDiscussion, enums.DealStatusConfirmed, enums.DealStatusCompleted:
		return s.confirmDeal(ctx, dealID, userID, status)
	case enums.DealStatusCancelled:
		return s.cancelDeal(ctx, dealID, status)
	case enums.DealStatusFailed:
		return domain.Deal{}, domain.ErrForbidden
	default:
		return domain.Deal{}, fmt.Errorf("invalid status: %s", status)
	}
}

// ================================================================================
// HELPER METHODS
// ================================================================================

// ================================================================================
// CREATE DEAL
// ================================================================================

// createDeal creates a new deal based on the provided draft and its associated offers.
//
// Errors:
//   - domain.ErrDraftNotFound
func (s *Service) createDeal(ctx context.Context, tx pgx.Tx, draft domain.Draft) (uuid.UUID, []copiedItemPhoto, error) {
	items := make([]domain.Item, len(draft.Offers))
	for i, o := range draft.Offers {
		var receiver *uuid.UUID = nil
		var provider *uuid.UUID = nil
		offerID := o.Offer.ID

		if o.Offer.Action == enums.OfferActionGive {
			provider = &o.Offer.AuthorId
		} else {
			receiver = &o.Offer.AuthorId
		}

		items[i] = domain.Item{
			ID:          uuid.New(),
			OfferID:     &offerID,
			AuthorID:    o.Offer.AuthorId,
			ProviderID:  provider,
			ReceiverID:  receiver,
			Name:        o.Offer.Name,
			Description: o.Offer.Description,
			Type:        o.Offer.Type,
			Quantity:    o.Info.Quantity,
		}
	}

	id, err := s.dealsRepository.CreateDeal(ctx, tx, domain.Deal{
		Name:        draft.Name,
		Description: draft.Description,
		Items:       items,
	})
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("failed to create deal: %w", err)
	}

	copiedPhotos := make([]copiedItemPhoto, 0)
	for i, offerWithInfo := range draft.Offers {
		itemCopies, copyErr := s.copyOfferPhotosToItem(ctx, tx, items[i].ID, offerWithInfo.Offer.PhotoUrls)
		if copyErr != nil {
			s.cleanupCopiedItemPhotos(ctx, copiedPhotos)
			return uuid.Nil, nil, fmt.Errorf("failed to copy item photos: %w", copyErr)
		}
		copiedPhotos = append(copiedPhotos, itemCopies...)
	}

	err = s.draftsRepository.DeleteDraft(ctx, tx, draft.ID)
	if err != nil {
		s.cleanupCopiedItemPhotos(ctx, copiedPhotos)
		return uuid.Nil, nil, fmt.Errorf("failed to delete draft: %w", err)
	}

	return id, copiedPhotos, nil
}

func (s *Service) confirmDeal(ctx context.Context, id uuid.UUID, userID uuid.UUID, targetStatus enums.DealStatus) (domain.Deal, error) {
	var deal domain.Deal

	txErr := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		deal, err = s.dealsRepository.GetDealByID(ctx, tx, id)
		if err != nil {
			return err
		}

		if err = s.EnsureNoPendingFailureReview(ctx, tx, id); err != nil {
			return err
		}

		if deal.Status+1 != targetStatus {
			return domain.ErrInvalidDealStatus
		}

		if err = s.dealsRepository.SetStatusVote(ctx, tx, id, userID, targetStatus); err != nil {
			return err
		}

		votes, err := s.dealsRepository.GetStatusVotes(ctx, tx, id)
		if err != nil {
			return err
		}

		allVoted := len(votes) == len(deal.Participants)
		if allVoted {
			for _, v := range votes {
				if v != targetStatus {
					allVoted = false
					break
				}
			}
		}

		if allVoted {
			if targetStatus == enums.DealStatusDiscussion {
				ok, err := s.checkParticipants(ctx, tx, id)
				if err != nil {
					return err
				}
				if !ok {
					return domain.ErrDealParticipantsUnready
				}
				if err = s.joinsRepository.DeleteAllRequests(ctx, tx, id); err != nil {
					return err
				}
			}
			if err = s.dealsRepository.UpdateDealStatus(ctx, tx, id, targetStatus); err != nil {
				return err
			}
			if err = s.dealsRepository.DeleteStatusVotes(ctx, tx, id); err != nil {
				return err
			}
		}

		deal, err = s.dealsRepository.GetDealByID(ctx, tx, id)
		return err
	})

	if txErr != nil {
		return deal, txErr
	}

	if deal.Status == enums.DealStatusDiscussion && s.chatsClient != nil {
		participantStrs := make([]string, len(deal.Participants))
		for i, p := range deal.Participants {
			participantStrs[i] = p.String()
		}
		_, err := s.chatsClient.CreateChat(ctx, &chatspb.CreateChatRequest{
			DealId:         id.String(),
			ParticipantIds: participantStrs,
		})
		if err != nil {
			s.logger.Error("failed to create chat for deal",
				slog.String("deal_id", id.String()),
				slog.Any("error", err),
			)
		}
	}

	return deal, nil
}

type copiedItemPhoto struct {
	ItemID   uuid.UUID
	Position int
}

func (s *Service) copyOfferPhotosToItem(
	ctx context.Context,
	tx pgx.Tx,
	itemID uuid.UUID,
	photoURLs []string,
) ([]copiedItemPhoto, error) {
	if len(photoURLs) == 0 {
		return nil, nil
	}
	if s.itemPhotoStorage == nil {
		return nil, fmt.Errorf("item photo storage is not configured")
	}

	photos := make([]domain.ItemPhoto, 0, len(photoURLs))
	copies := make([]copiedItemPhoto, 0, len(photoURLs))
	for i, photoURL := range photoURLs {
		copiedURL, err := s.itemPhotoStorage.CopyPhoto(ctx, photoURL, itemID, i)
		if err != nil {
			return nil, err
		}
		photos = append(photos, domain.ItemPhoto{
			ID:       uuid.New(),
			ItemID:   itemID,
			URL:      copiedURL,
			Position: i,
		})
		copies = append(copies, copiedItemPhoto{
			ItemID:   itemID,
			Position: i,
		})
	}

	if err := s.dealsRepository.AddItemPhotos(ctx, tx, photos); err != nil {
		s.cleanupCopiedItemPhotos(ctx, copies)
		return nil, err
	}

	return copies, nil
}

func (s *Service) cleanupCopiedItemPhotos(ctx context.Context, copies []copiedItemPhoto) {
	if s.itemPhotoStorage == nil {
		return
	}

	for _, photo := range copies {
		if err := s.itemPhotoStorage.DeletePhoto(ctx, photo.ItemID, photo.Position); err != nil {
			s.logger.Warn(
				"failed to cleanup copied item photo",
				slog.String("item_id", photo.ItemID.String()),
				slog.Int("position", photo.Position),
				slog.Any("error", err),
			)
		}
	}
}

func (s *Service) checkParticipants(ctx context.Context, tx pgx.Tx, dealID uuid.UUID) (bool, error) {
	deal, err := s.dealsRepository.GetDealByID(ctx, tx, dealID)
	if err != nil {
		return false, err
	}

	type participantStats struct {
		hasProvider bool
		hasReceiver bool
	}

	statsByParticipant := make(map[uuid.UUID]participantStats, len(deal.Participants))
	for _, participantID := range deal.Participants {
		statsByParticipant[participantID] = participantStats{}
	}

	for _, item := range deal.Items {
		if item.ProviderID == nil || item.ReceiverID == nil {
			return false, nil
		}

		if stats, ok := statsByParticipant[*item.ProviderID]; ok {
			stats.hasProvider = true
			statsByParticipant[*item.ProviderID] = stats
		}

		if stats, ok := statsByParticipant[*item.ReceiverID]; ok {
			stats.hasReceiver = true
			statsByParticipant[*item.ReceiverID] = stats
		}
	}

	for _, stats := range statsByParticipant {
		if !stats.hasProvider || !stats.hasReceiver {
			return false, nil
		}
	}

	return true, nil
}

func (s *Service) cancelDeal(ctx context.Context, id uuid.UUID, targetStatus enums.DealStatus) (domain.Deal, error) {
	var deal domain.Deal

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		deal, err = s.dealsRepository.GetDealByID(ctx, tx, id)
		if err != nil {
			return err
		}

		if err = s.EnsureNoPendingFailureReview(ctx, tx, id); err != nil {
			return err
		}

		finalStatuses := []enums.DealStatus{
			enums.DealStatusCompleted,
			enums.DealStatusCancelled,
			enums.DealStatusFailed,
		}
		for _, fs := range finalStatuses {
			if deal.Status == fs {
				return domain.ErrInvalidDealStatus
			}
		}

		if err = s.dealsRepository.UpdateDealStatus(ctx, tx, id, targetStatus); err != nil {
			return err
		}
		if err = s.dealsRepository.DeleteStatusVotes(ctx, tx, id); err != nil {
			return err
		}

		deal, err = s.dealsRepository.GetDealByID(ctx, tx, id)
		return err
	})

	return deal, err
}

func (s *Service) EnsureNoPendingFailureReview(ctx context.Context, tx pgx.Tx, dealID uuid.UUID) error {
	hasPending, err := s.failuresRepo.HasPendingFailureReview(ctx, tx, dealID)
	if err != nil {
		return err
	}
	if hasPending {
		return domain.ErrFailureReviewRequired
	}

	return nil
}

func ContainsUserID(items []uuid.UUID, userID uuid.UUID) bool {
	for _, item := range items {
		if item == userID {
			return true
		}
	}
	return false
}
