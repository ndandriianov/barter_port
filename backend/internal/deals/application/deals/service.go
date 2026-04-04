package deals

import (
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/internal/deals/infrastructure/repository/deals"
	"barter-port/internal/deals/infrastructure/repository/drafts"
	"barter-port/internal/deals/infrastructure/repository/offers"
	"barter-port/pkg/db"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db               *pgxpool.Pool
	draftsRepository *drafts.Repository
	dealsRepository  *deals.Repository
	offersRepository *offers.Repository
}

func NewService(db *pgxpool.Pool, draftsRepo *drafts.Repository, dealsRepo *deals.Repository) *Service {
	return &Service{db: db, draftsRepository: draftsRepo, dealsRepository: dealsRepo}
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
	offers []domain.OfferIDAndInfo,
) (uuid.UUID, error) {
	var id uuid.UUID
	var err error

	txErr := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		if len(offers) == 0 {
			return domain.ErrNoOffers
		}
		// TODO: проверить offers

		id, err = s.draftsRepository.CreateDraft(ctx, tx, authorID, name, description, offers)
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

			id, err = s.createDeal(ctx, tx, draft)
			if err != nil {
				return fmt.Errorf("could not create deal: %w", err)
			}
		}

		return nil
	})
	if err != nil {
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

// GetDeals returns deal IDs with participant UUIDs. If my is true, filters to only deals the user participates in.
//
// No domain errors.
func (s *Service) GetDeals(ctx context.Context, userID uuid.UUID, my bool) ([]htypes.DealIDWithParticipantIDs, error) {
	var filterUserID *uuid.UUID
	if my {
		filterUserID = &userID
	}
	return s.dealsRepository.GetDealIDs(ctx, s.db, filterUserID)
}

// ================================================================================
// GET DEAL BY ID
// ================================================================================

// GetDealByID returns a deal by its ID.
//
// Domain errors:
//   - domain.ErrDealNotFound: if no deal with the specified ID exists.
func (s *Service) GetDealByID(ctx context.Context, id uuid.UUID) (domain.Deal, error) {
	return s.dealsRepository.GetDealByID(ctx, s.db, id)
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
func (s *Service) createDeal(ctx context.Context, tx pgx.Tx, draft domain.Draft) (uuid.UUID, error) {
	items := make([]domain.Item, len(draft.Offers))
	for i, o := range draft.Offers {
		var receiver *uuid.UUID = nil
		var provider *uuid.UUID = nil

		if o.Offer.Action == enums.OfferActionGive {
			provider = &o.Offer.AuthorId
		} else {
			receiver = &o.Offer.AuthorId
		}

		items[i] = domain.Item{
			ID:          o.Offer.ID,
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
		return uuid.Nil, fmt.Errorf("failed to create deal: %w", err)
	}

	err = s.draftsRepository.DeleteDraft(ctx, tx, draft.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to delete draft: %w", err)
	}

	return id, nil
}
