package offers

import (
	"barter-port/internal/deals/domain"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ================================================================================
// ADD OFFER
// ================================================================================

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

// ================================================================================
// GET OFFERS ORDER BY TIME
// ================================================================================

// GetOffersOrderByTime retrieves items from the database ordered by creation time.
// It supports cursor-based pagination using a TimeCursor.
// If the cursor is nil, it retrieves the most recent items.
// Returns a slice of items, a new TimeCursor for the next page, and an error if the query fails.
func (r *Repository) GetOffersOrderByTime(
	ctx context.Context,
	cursor *domain.TimeCursor,
	limit int,
	authorID *uuid.UUID,
) ([]domain.Offer, *domain.TimeCursor, error) {

	var query string
	var args []interface{}

	if authorID == nil {
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
	} else {
		if cursor == nil {
			query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM offers
		WHERE author_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		`
			args = append(args, *authorID, limit)
		} else {
			query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM offers
		WHERE author_id = $1 AND (created_at, id) < ($2, $3)
		ORDER BY created_at DESC
		LIMIT $4
		`
			args = append(args, *authorID, cursor.CreatedAt, cursor.Id, limit)
		}
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

// ================================================================================
// GET OFFERS ORDER BY POPULARITY
// ================================================================================

// GetOffersOrderByPopularity retrieves items from the database ordered by popularity (views).
// It supports cursor-based pagination using a PopularityCursor.
// If the cursor is nil, it retrieves the most popular items.
// Returns a slice of items, a new PopularityCursor for the next page, and an error if the query fails.
func (r *Repository) GetOffersOrderByPopularity(
	ctx context.Context,
	cursor *domain.PopularityCursor,
	limit int,
	authorID *uuid.UUID,
) ([]domain.Offer, *domain.PopularityCursor, error) {

	var query string
	var args []interface{}

	if authorID == nil {
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
	} else {
		if cursor == nil {
			query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM offers
		WHERE author_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		`
			args = append(args, *authorID, limit)
		} else {
			query = `
		SELECT id, author_id, name, type, action, description, created_at, views
		FROM offers
		WHERE author_id = $1 AND (views, id) < ($2, $3)
		ORDER BY views DESC
		LIMIT $4
		`
			args = append(args, *authorID, cursor.Views, cursor.Id, limit)
		}
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

// ================================================================================
// GET OFFER NAMES BY IDS
// ================================================================================

// GetOfferNamesByIDs returns offer names for the given IDs, preserving input order.
// IDs not found in the database are silently skipped.
func (r *Repository) GetOfferNamesByIDs(ctx context.Context, exec db.DB, ids []uuid.UUID) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT o.name
		FROM unnest($1::uuid[]) WITH ORDINALITY u(id, ord)
		JOIN offers o ON o.id = u.id
		ORDER BY u.ord`

	rows, err := exec.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("sql get offer names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan offer name: %w", err)
		}
		names = append(names, name)
	}

	return names, rows.Err()
}

// GetOffersByIDs returns offers for the provided IDs, preserving input order.
// IDs not found in the database are silently skipped.
func (r *Repository) GetOffersByIDs(ctx context.Context, exec db.DB, ids []uuid.UUID) ([]domain.Offer, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT
			o.id,
			o.author_id,
			o.name,
			o.type,
			o.action,
			o.description,
			o.created_at,
			o.views
		FROM unnest($1::uuid[]) WITH ORDINALITY u(id, ord)
		JOIN offers o ON o.id = u.id
		ORDER BY u.ord`

	rows, err := exec.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("sql get offers by ids: %w", err)
	}
	defer rows.Close()

	result := make([]domain.Offer, 0, len(ids))
	for rows.Next() {
		var item domain.Offer
		if err = rows.Scan(
			&item.ID,
			&item.AuthorId,
			&item.Name,
			&item.Type,
			&item.Action,
			&item.Description,
			&item.CreatedAt,
			&item.Views,
		); err != nil {
			return nil, fmt.Errorf("scan offer: %w", err)
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

// ================================================================================
// GET OFFER
// ================================================================================

// GetOfferByID retrieves a single offer by its ID using the repository's own pool.
//
// Errors:
//   - domain.ErrOfferNotFound: if no item with the given ID exists.
func (r *Repository) GetOfferByID(ctx context.Context, id uuid.UUID) (*domain.Offer, error) {
	return r.GetOffer(ctx, r.db, id)
}

// GetOffer retrieves a single item from the database by its ID.
//
// Errors:
//   - domain.ErrOfferNotFound: if no item with the given ID exists.
func (r *Repository) GetOffer(ctx context.Context, exec db.DB, id uuid.UUID) (*domain.Offer, error) {
	query := `
		SELECT id, author_id, name, type, action, description, created_at
		FROM offers
		WHERE id = $1`

	var offer domain.Offer
	err := exec.QueryRow(ctx, query, id).Scan(
		&offer.ID,
		&offer.AuthorId,
		&offer.Name,
		&offer.Type,
		&offer.Action,
		&offer.Description,
		&offer.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOfferNotFound
		}
		return nil, err
	}

	return &offer, nil
}
