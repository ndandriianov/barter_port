package deals

import (
	"barter-port/internal/deals/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// ================================================================================
// CREATE DEAL
// ================================================================================

// CreateDeal creates a new deal and its associated items in the database within a transaction.
//
// No domain Errors.
func (r *Repository) CreateDeal(ctx context.Context, tx pgx.Tx, deal domain.Deal) (uuid.UUID, error) {
	dealQuery := `
		INSERT INTO deals (name, description)
		VALUES ($1, $2)
		RETURNING id;`

	var id uuid.UUID
	err := tx.QueryRow(ctx, dealQuery, deal.Name, deal.Description).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("sql deal: %w", err)
	}

	offersQuery := `
		INSERT INTO items (deal_id, author_id, receiver_id, name, description, type) 
		VALUES ($1, $2, $3, $4, $5, $6)`

	for _, item := range deal.Items {
		_, err = tx.Exec(ctx, offersQuery, id, item.AuthorID, item.ReceiverID, item.Name, item.Description, item.Type)
		if err != nil {
			return uuid.Nil, fmt.Errorf("sql items: %w", err)
		}
	}

	return id, nil
}
