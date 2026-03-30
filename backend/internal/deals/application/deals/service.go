package deals

import (
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/infrastructure/repository/deals"
	"barter-port/pkg/db"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db              *pgxpool.Pool
	dealsRepository *deals.Repository
}

func NewService(db *pgxpool.Pool, repo *deals.Repository) *Service {
	return &Service{db: db, dealsRepository: repo}
}

// ================================================================================
// CREATE DRAFT
// ================================================================================

// CreateDraft inserts a new draft deal into the database and returns its ID.
//
// No domain errors
func (s *Service) CreateDraft(
	ctx context.Context,
	authorID uuid.UUID,
	name *string,
	description *string,
	items []domain.ItemIDsAndInfo,
) (uuid.UUID, error) {
	var id uuid.UUID
	var err error

	txErr := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		// TODO: проверить items

		id, err = s.dealsRepository.CreateDraft(ctx, tx, authorID, name, description, items)
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
	return s.dealsRepository.GetDraftIDsByAuthor(ctx, s.db, authorID)
}

// ================================================================================
// GET DRAFT BY ID
// ================================================================================

// GetDraftByID returns a draft deal by its ID.
//
// Domain errors:
// - domain.ErrDraftNotFound: if no draft deal with the specified ID exists.
func (s *Service) GetDraftByID(ctx context.Context, id uuid.UUID) (domain.Draft, error) {
	return s.dealsRepository.GetDraftByID(ctx, s.db, id)
}
