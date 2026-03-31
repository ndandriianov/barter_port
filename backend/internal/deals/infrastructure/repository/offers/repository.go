package offers

import (
	"barter-port/internal/deals/domain"
	"barter-port/pkg/repox"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// AddOffer inserts a new item into the database.
// Returns an error if the insertion fails.
func (r *Repository) AddOffer(ctx context.Context, offer domain.Offer) error {
	query := `
		INSERT INTO offers (id, author_id, name, type, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query, offer.ID, offer.AuthorId, offer.Name, offer.Type.String(), offer.Action.String(), offer.Description, offer.CreatedAt)
	return err
}

// GetOffersOrderByTime retrieves items from the database ordered by creation time.
// It supports cursor-based pagination using a TimeCursor.
// If the cursor is nil, it retrieves the most recent items.
// Returns a slice of items, a new TimeCursor for the next page, and an error if the query fails.
func (r *Repository) GetOffersOrderByTime(
	ctx context.Context,
	cursor *domain.TimeCursor,
	limit int,
) ([]domain.Offer, *domain.TimeCursor, error) {

	var query string
	var args []interface{}

	if cursor == nil {
		query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM offers
		ORDER BY created_at DESC
		LIMIT $1
		`
		args = append(args, limit)
	} else {
		query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM offers
		WHERE (created_at, id) < ($1, $2)
		ORDER BY created_at DESC 
		LIMIT $3
		`
		args = append(args, cursor.CreatedAt, cursor.Id, limit)
	}

	offers, err := repox.FetchStructs[domain.Offer](ctx, r.db, query, args...)
	if err != nil {
		return nil, nil, err
	}

	if len(offers) == 0 {
		return offers, nil, nil
	}

	lastOffer := offers[len(offers)-1]
	newCursor := domain.TimeCursor{
		CreatedAt: lastOffer.CreatedAt,
		Id:        lastOffer.ID,
	}

	return offers, &newCursor, nil
}

// GetItemsOrderByPopularity retrieves items from the database ordered by popularity (views).
// It supports cursor-based pagination using a PopularityCursor.
// If the cursor is nil, it retrieves the most popular items.
// Returns a slice of items, a new PopularityCursor for the next page, and an error if the query fails.
func (r *Repository) GetItemsOrderByPopularity(
	ctx context.Context,
	cursor *domain.PopularityCursor,
	limit int,
) ([]domain.Offer, *domain.PopularityCursor, error) {

	var query string
	var args []interface{}

	if cursor == nil {
		query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM offers
		ORDER BY created_at DESC
		LIMIT $1
		`
		args = append(args, limit)
	} else {
		query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM offers
		WHERE (views, id) < ($1, $2) 
		ORDER BY views DESC 
		LIMIT $3
		`
		args = append(args, cursor.Views, cursor.Id, limit)
	}

	offers, err := repox.FetchStructs[domain.Offer](ctx, r.db, query, args...)
	if err != nil {
		return nil, nil, err
	}

	if len(offers) == 0 {
		return offers, nil, nil
	}

	lastOffer := offers[len(offers)-1]
	newCursor := domain.PopularityCursor{
		Views: lastOffer.Views,
		Id:    lastOffer.ID,
	}

	return offers, &newCursor, nil
}
