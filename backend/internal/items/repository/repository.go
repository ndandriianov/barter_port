package repository

import (
	"barter-port/internal/items/model"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

type ItemRepository struct {
	db *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r ItemRepository) AddItem(ctx context.Context, item model.Item) error {
	query := `
		INSERT INTO items (id, name, type, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query, item.ID, item.Name, item.Type.String(), item.Action.String(), item.Description, item.CreatedAt)
	return err
}

func (r ItemRepository) GetItemsOrderByTime(ctx context.Context, nextCursor uuid.UUID, limit int) ([]model.Item, error) {
	query := `
		SELECT id, name, type, action, description, created_at
		FROM items
		WHERE id > $1
		ORDER BY created_at DESC 
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, nextCursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Item])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []model.Item{}, nil
		}
		return nil, err
	}

	return items, nil
}

func (r ItemRepository) GetItemsOrderByPopularity(ctx context.Context, nextCursor uuid.UUID, limit int) ([]model.Item, error) {
	query := `
		SELECT id, name, type, action, description, created_at
		FROM items
		WHERE id > $1
		ORDER BY views DESC 
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, nextCursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Item])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []model.Item{}, nil
		}
		return nil, err
	}

	return items, nil
}
