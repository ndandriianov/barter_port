package deals

import (
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/internal/deals/infrastructure/repository/drafts"
	"barter-port/pkg/db"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db               *pgxpool.Pool
	draftsRepository *drafts.Repository
}

func NewService(db *pgxpool.Pool, repo *drafts.Repository) *Service {
	return &Service{db: db, draftsRepository: repo}
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

// GetDraftIDsByAuthor returns a list of draft deal IDs created by the specified author.
//
// No domain errors
func (s *Service) GetDraftIDsByAuthor(ctx context.Context, authorID uuid.UUID) ([]uuid.UUID, error) {
	return s.draftsRepository.GetDraftIDsByAuthor(ctx, s.db, authorID)
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
// Errors: no domain errors
func (s *Service) ConfirmDraft(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]htypes.UserConfirmed, error) {
	var users []htypes.UserConfirmed

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		err := s.draftsRepository.ConfirmDraftByID(ctx, tx, id, userID)
		if err != nil {
			return err
		}

		users, err = s.draftsRepository.GetConfirms(ctx, tx, id)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}
