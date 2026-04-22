package offers

import (
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/pkg/db"
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
func (r *Repository) AddOffer(ctx context.Context, exec db.DB, offer domain.Offer) error {
	query := `
		INSERT INTO offers (id, author_id, name, type, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := exec.Exec(ctx, query, offer.ID, offer.AuthorId, offer.Name, offer.Type.String(), offer.Action.String(), offer.Description, offer.CreatedAt)
	return err
}

func (r *Repository) ReplaceOfferTags(ctx context.Context, exec db.DB, offerID uuid.UUID, tags []string) error {
	if _, err := exec.Exec(ctx, `DELETE FROM offer_tags WHERE offer_id = $1`, offerID); err != nil {
		return fmt.Errorf("delete offer tags: %w", err)
	}

	if len(tags) > 0 {
		if _, err := exec.Exec(ctx, `
			INSERT INTO tags (name)
			SELECT DISTINCT tag_name
			FROM unnest($1::text[]) AS tag_name
			ON CONFLICT (name) DO NOTHING
		`, tags); err != nil {
			return fmt.Errorf("upsert tags: %w", err)
		}

		if _, err := exec.Exec(ctx, `
			INSERT INTO offer_tags (offer_id, tag_name)
			SELECT $1, tag_name
			FROM unnest($2::text[]) AS tag_name
		`, offerID, tags); err != nil {
			return fmt.Errorf("insert offer tags: %w", err)
		}
	}

	if err := r.DeleteUnusedTags(ctx, exec); err != nil {
		return err
	}

	return nil
}

func (r *Repository) AddOfferPhotos(ctx context.Context, exec db.DB, photos []domain.OfferPhoto) error {
	if len(photos) == 0 {
		return nil
	}

	query := `
		INSERT INTO offer_photos (id, offer_id, url, position)
		VALUES ($1, $2, $3, $4)
	`

	for _, photo := range photos {
		if _, err := exec.Exec(ctx, query, photo.ID, photo.OfferID, photo.URL, photo.Position); err != nil {
			return fmt.Errorf("insert offer photo: %w", err)
		}
	}

	return nil
}

func (r *Repository) GetOfferPhotos(ctx context.Context, exec db.DB, offerID uuid.UUID) ([]domain.OfferPhoto, error) {
	const query = `
		SELECT id, offer_id, url, position
		FROM offer_photos
		WHERE offer_id = $1
		ORDER BY position
	`

	rows, err := exec.Query(ctx, query, offerID)
	if err != nil {
		return nil, fmt.Errorf("sql get offer photos: %w", err)
	}
	defer rows.Close()

	result := make([]domain.OfferPhoto, 0)
	for rows.Next() {
		var photo domain.OfferPhoto
		if err = rows.Scan(&photo.ID, &photo.OfferID, &photo.URL, &photo.Position); err != nil {
			return nil, fmt.Errorf("scan offer photo: %w", err)
		}
		result = append(result, photo)
	}

	return result, rows.Err()
}

func (r *Repository) DeleteOfferPhotos(ctx context.Context, exec db.DB, offerID uuid.UUID, photoIDs []uuid.UUID) error {
	if len(photoIDs) == 0 {
		return nil
	}

	const query = `
		DELETE FROM offer_photos
		WHERE offer_id = $1
		  AND id = ANY($2)
	`

	if _, err := exec.Exec(ctx, query, offerID, photoIDs); err != nil {
		return fmt.Errorf("delete offer photos: %w", err)
	}

	return nil
}

func (r *Repository) ListTags(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT name FROM tags ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	result := make([]string, 0)
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		result = append(result, tag)
	}

	return result, rows.Err()
}

func (r *Repository) DeleteTagByName(ctx context.Context, exec db.DB, name string) error {
	tag, err := exec.Exec(ctx, `DELETE FROM tags WHERE name = $1`, name)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTagNotFound
	}

	return nil
}

func (r *Repository) DeleteUnusedTags(ctx context.Context, exec db.DB) error {
	if _, err := exec.Exec(ctx, `
		DELETE FROM tags t
		WHERE NOT EXISTS (
			SELECT 1
			FROM offer_tags ot
			WHERE ot.tag_name = t.name
		)
	`); err != nil {
		return fmt.Errorf("delete unused tags: %w", err)
	}

	return nil
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
			o.updated_at,
			o.views,
			COALESCE((SELECT array_agg(ot.tag_name ORDER BY ot.tag_name) FROM offer_tags ot WHERE ot.offer_id = o.id), '{}'::text[]) AS tags,
			COALESCE((SELECT array_agg(op.id ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = o.id), '{}'::uuid[]) AS photo_ids,
			COALESCE((SELECT array_agg(op.url ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = o.id), '{}'::text[]) AS photo_urls,
			o.is_hidden,
			o.modification_blocked
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
			&item.UpdatedAt,
			&item.Views,
			&item.Tags,
			&item.PhotoIds,
			&item.PhotoUrls,
			&item.IsHidden,
			&item.ModificationBlocked,
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
		SELECT id, author_id, name, type, action, description, created_at, updated_at, views,
		       COALESCE((SELECT array_agg(ot.tag_name ORDER BY ot.tag_name) FROM offer_tags ot WHERE ot.offer_id = offers.id), '{}'::text[]) AS tags,
		       COALESCE((SELECT array_agg(op.id ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = offers.id), '{}'::uuid[]) AS photo_ids,
		       COALESCE((SELECT array_agg(op.url ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = offers.id), '{}'::text[]) AS photo_urls,
		       is_hidden, modification_blocked
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
		&offer.UpdatedAt,
		&offer.Views,
		&offer.Tags,
		&offer.PhotoIds,
		&offer.PhotoUrls,
		&offer.IsHidden,
		&offer.ModificationBlocked,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOfferNotFound
		}
		return nil, err
	}

	return &offer, nil
}

// ViewOffer increments the offer views counter.
//
// Errors:
//   - domain.ErrOfferNotFound: if no item with the given ID exists.
func (r *Repository) ViewOffer(ctx context.Context, exec db.DB, id uuid.UUID) error {
	tag, err := exec.Exec(ctx, `UPDATE offers SET views = views + 1 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("sql view offer: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOfferNotFound
	}

	return nil
}

// UpdateOffer applies a partial update to an offer.
//
// Domain errors:
//   - domain.ErrOfferNotFound: if the offer does not exist.
//   - domain.ErrForbidden: if the offer exists but belongs to another author.
func (r *Repository) UpdateOffer(
	ctx context.Context,
	exec db.DB,
	offerID uuid.UUID,
	userID uuid.UUID,
	patch htypes.OfferPatch,
) (domain.Offer, error) {
	var itemType *string
	if patch.Type != nil {
		itemType = new(patch.Type.String())
	}

	var action *string
	if patch.Action != nil {
		action = new(patch.Action.String())
	}

	const query = `
		UPDATE offers
		SET name        = COALESCE($3, name),
		    description = COALESCE($4, description),
		    type        = COALESCE($5, type),
		    action      = COALESCE($6, action),
		    updated_at  = NOW()
		WHERE id = $1
		  AND author_id = $2
		RETURNING id, author_id, name, type, action, description, created_at, updated_at, views,
		          COALESCE((SELECT array_agg(ot.tag_name ORDER BY ot.tag_name) FROM offer_tags ot WHERE ot.offer_id = offers.id), '{}'::text[]) AS tags,
		          COALESCE((SELECT array_agg(op.id ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = offers.id), '{}'::uuid[]) AS photo_ids,
		          COALESCE((SELECT array_agg(op.url ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = offers.id), '{}'::text[]) AS photo_urls,
		          is_hidden, modification_blocked`

	var offer domain.Offer
	err := exec.QueryRow(ctx, query, offerID, userID, patch.Name, patch.Description, itemType, action).Scan(
		&offer.ID,
		&offer.AuthorId,
		&offer.Name,
		&offer.Type,
		&offer.Action,
		&offer.Description,
		&offer.CreatedAt,
		&offer.UpdatedAt,
		&offer.Views,
		&offer.Tags,
		&offer.PhotoIds,
		&offer.PhotoUrls,
		&offer.IsHidden,
		&offer.ModificationBlocked,
	)
	if err == nil {
		return offer, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Offer{}, fmt.Errorf("sql update offer: %w", err)
	}

	var authorID uuid.UUID
	checkErr := exec.QueryRow(ctx, `SELECT author_id FROM offers WHERE id = $1`, offerID).Scan(&authorID)
	if errors.Is(checkErr, pgx.ErrNoRows) {
		return domain.Offer{}, domain.ErrOfferNotFound
	}
	if checkErr != nil {
		return domain.Offer{}, fmt.Errorf("sql check offer: %w", checkErr)
	}

	return domain.Offer{}, domain.ErrForbidden
}

// DeleteOffer deletes an offer owned by the specified author.
//
// Domain errors:
//   - domain.ErrOfferNotFound: if the offer does not exist.
//   - domain.ErrForbidden: if the offer exists but belongs to another author.
func (r *Repository) DeleteOffer(ctx context.Context, exec db.DB, offerID uuid.UUID, userID uuid.UUID) error {
	tag, err := exec.Exec(ctx, `DELETE FROM offers WHERE id = $1 AND author_id = $2`, offerID, userID)
	if err != nil {
		return fmt.Errorf("sql delete offer: %w", err)
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	var authorID uuid.UUID
	checkErr := exec.QueryRow(ctx, `SELECT author_id FROM offers WHERE id = $1`, offerID).Scan(&authorID)
	if errors.Is(checkErr, pgx.ErrNoRows) {
		return domain.ErrOfferNotFound
	}
	if checkErr != nil {
		return fmt.Errorf("sql check offer before delete: %w", checkErr)
	}

	return domain.ErrForbidden
}
