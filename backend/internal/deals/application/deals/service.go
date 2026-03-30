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
	dealsRepository deals.Repository
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateDraft(
	ctx context.Context,
	authorID uuid.UUID,
	name *string,
	description *string,
	items []struct {
		ID       uuid.UUID
		Quantity int
	},
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

func (s *Service) GetDraftIDsByAuthor(ctx context.Context, authorID uuid.UUID) ([]uuid.UUID, error) {
	return s.dealsRepository.GetDraftIDsByAuthor(ctx, s.db, authorID)
}

func (s *Service) GetDraftByID(ctx context.Context, id uuid.UUID) (domain.Draft, error) {
	return s.dealsRepository.GetDraftByID(ctx, s.db, id)
}
