package repository

import (
	"barter-port/internal/items/model"
	"barter-port/pkg/repox"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

type ItemRepository struct {
	db *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{db: db}
}

// AddItem inserts a new item into the database.
// Returns an error if the insertion fails.
func (r *ItemRepository) AddItem(ctx context.Context, item model.Item) error {
	query := `
		INSERT INTO items (id, author_id, name, type, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query, item.ID, item.AuthorId, item.Name, item.Type.String(), item.Action.String(), item.Description, item.CreatedAt)
	return err
}

// GetItemsOrderByTime retrieves items from the database ordered by creation time.
// It supports cursor-based pagination using a TimeCursor.
// If the cursor is nil, it retrieves the most recent items.
// Returns a slice of items, a new TimeCursor for the next page, and an error if the query fails.
func (r *ItemRepository) GetItemsOrderByTime(
	ctx context.Context,
	cursor *model.TimeCursor,
	limit int,
) ([]model.Item, *model.TimeCursor, error) {

	var query string
	var args []interface{}

	if cursor == nil {
		query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM items
		ORDER BY created_at DESC
		LIMIT $1
		`
		args = append(args, limit)
	} else {
		query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM items
		WHERE (created_at, id) < ($1, $2)
		ORDER BY created_at DESC 
		LIMIT $3
		`
		args = append(args, cursor.CreatedAt, cursor.Id, limit)
	}

	items, err := repox.FetchStructs[model.Item](ctx, r.db, query, args...)
	if err != nil {
		return nil, nil, err
	}

	if len(items) == 0 {
		return items, nil, nil
	}

	lastItem := items[len(items)-1]
	newCursor := model.TimeCursor{
		CreatedAt: lastItem.CreatedAt,
		Id:        lastItem.ID,
	}

	return items, &newCursor, nil
}

// GetItemsOrderByPopularity retrieves items from the database ordered by popularity (views).
// It supports cursor-based pagination using a PopularityCursor.
// If the cursor is nil, it retrieves the most popular items.
// Returns a slice of items, a new PopularityCursor for the next page, and an error if the query fails.
func (r *ItemRepository) GetItemsOrderByPopularity(
	ctx context.Context,
	cursor *model.PopularityCursor,
	limit int,
) ([]model.Item, *model.PopularityCursor, error) {

	var query string
	var args []interface{}

	if cursor == nil {
		query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM items
		ORDER BY created_at DESC
		LIMIT $1
		`
		args = append(args, limit)
	} else {
		query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM items
		WHERE (views, id) < ($1, $2) 
		ORDER BY views DESC 
		LIMIT $3
		`
		args = append(args, cursor.Views, cursor.Id, limit)
	}

	items, err := repox.FetchStructs[model.Item](ctx, r.db, query, args...)
	if err != nil {
		return nil, nil, err
	}

	if len(items) == 0 {
		return items, nil, nil
	}

	lastItem := items[len(items)-1]
	newCursor := model.PopularityCursor{
		Views: lastItem.Views,
		Id:    lastItem.ID,
	}

	return items, &newCursor, nil
}
