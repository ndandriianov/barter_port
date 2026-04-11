package deals

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const reviewSelectCols = `id, deal_id, item_id, offer_id, author_id, provider_id, rating, comment, created_at, updated_at`

func scanReview(row interface{ Scan(...any) error }) (domain.Review, error) {
	var r domain.Review
	err := row.Scan(
		&r.ID,
		&r.DealID,
		&r.ItemID,
		&r.OfferID,
		&r.AuthorID,
		&r.ProviderID,
		&r.Rating,
		&r.Comment,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Review{}, domain.ErrReviewNotFound
	}
	if err != nil {
		return domain.Review{}, fmt.Errorf("scan review: %w", err)
	}
	return r, nil
}

func scanReviews(rows pgx.Rows) ([]domain.Review, error) {
	var result []domain.Review
	for rows.Next() {
		r, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	if result == nil {
		result = []domain.Review{}
	}
	return result, nil
}

// ================================================================================
// CREATE REVIEW
// ================================================================================

func (r *Repository) CreateReview(
	ctx context.Context,
	tx pgx.Tx,
	dealID, authorID, providerID uuid.UUID,
	itemID, offerID *uuid.UUID,
	rating int,
	comment *string,
) (domain.Review, error) {
	query := `
		INSERT INTO deal_reviews (deal_id, item_id, offer_id, author_id, provider_id, rating, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + reviewSelectCols

	review, err := scanReview(tx.QueryRow(ctx, query, dealID, itemID, offerID, authorID, providerID, rating, comment))
	if err != nil {
		if repox.IsUniqueViolation(err) {
			return domain.Review{}, domain.ErrReviewAlreadyExists
		}
		// scanReview wraps ErrNoRows as ErrReviewNotFound, but INSERT won't return no-rows on success
		// Re-check: if we got ErrReviewNotFound here it means an INSERT returned no rows (shouldn't happen)
		return domain.Review{}, fmt.Errorf("sql create review: %w", err)
	}
	return review, nil
}

// ================================================================================
// GET REVIEW BY ID
// ================================================================================

func (r *Repository) GetReviewByID(
	ctx context.Context,
	exec db.DB,
	reviewID uuid.UUID,
) (domain.Review, error) {
	query := `SELECT ` + reviewSelectCols + ` FROM deal_reviews WHERE id = $1`
	review, err := scanReview(exec.QueryRow(ctx, query, reviewID))
	if err != nil {
		return domain.Review{}, fmt.Errorf("sql get review by id: %w", err)
	}
	return review, nil
}

// ================================================================================
// UPDATE REVIEW
// ================================================================================

func (r *Repository) UpdateReview(
	ctx context.Context,
	tx pgx.Tx,
	reviewID uuid.UUID,
	rating *int,
	comment *string,
) (domain.Review, error) {
	query := `
		UPDATE deal_reviews
		SET rating     = COALESCE($2, rating),
		    comment    = CASE WHEN $3::text IS NOT NULL THEN $3 ELSE comment END,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING ` + reviewSelectCols

	review, err := scanReview(tx.QueryRow(ctx, query, reviewID, rating, comment))
	if err != nil {
		return domain.Review{}, fmt.Errorf("sql update review: %w", err)
	}
	return review, nil
}

// ================================================================================
// DELETE REVIEW
// ================================================================================

func (r *Repository) DeleteReview(
	ctx context.Context,
	tx pgx.Tx,
	reviewID uuid.UUID,
) error {
	tag, err := tx.Exec(ctx, `DELETE FROM deal_reviews WHERE id = $1`, reviewID)
	if err != nil {
		return fmt.Errorf("sql delete review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrReviewNotFound
	}
	return nil
}

// ================================================================================
// GET REVIEWS BY OFFER ID
// ================================================================================

func (r *Repository) GetReviewsByOfferID(
	ctx context.Context,
	exec db.DB,
	offerID uuid.UUID,
) ([]domain.Review, error) {
	query := `SELECT ` + reviewSelectCols + `
		FROM deal_reviews
		WHERE offer_id = $1
		ORDER BY created_at DESC`

	rows, err := exec.Query(ctx, query, offerID)
	if err != nil {
		return nil, fmt.Errorf("sql get reviews by offer id: %w", err)
	}
	defer rows.Close()
	return scanReviews(rows)
}

// ================================================================================
// GET REVIEWS SUMMARY BY OFFER ID
// ================================================================================

func (r *Repository) GetReviewsSummaryByOfferID(
	ctx context.Context,
	exec db.DB,
	offerID uuid.UUID,
) (htypes.ReviewSummary, error) {
	return r.getReviewsSummary(ctx, exec, "offer_id", offerID)
}

// ================================================================================
// GET REVIEWS BY PROVIDER ID
// ================================================================================

func (r *Repository) GetReviewsByProviderID(
	ctx context.Context,
	exec db.DB,
	providerID uuid.UUID,
) ([]domain.Review, error) {
	query := `SELECT ` + reviewSelectCols + `
		FROM deal_reviews
		WHERE provider_id = $1
		ORDER BY created_at DESC`

	rows, err := exec.Query(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("sql get reviews by provider id: %w", err)
	}
	defer rows.Close()
	return scanReviews(rows)
}

// ================================================================================
// GET REVIEWS SUMMARY BY PROVIDER ID
// ================================================================================

func (r *Repository) GetReviewsSummaryByProviderID(
	ctx context.Context,
	exec db.DB,
	providerID uuid.UUID,
) (htypes.ReviewSummary, error) {
	return r.getReviewsSummary(ctx, exec, "provider_id", providerID)
}

// getReviewsSummary is a shared helper for offer and provider summary queries.
func (r *Repository) getReviewsSummary(
	ctx context.Context,
	exec db.DB,
	column string,
	id uuid.UUID,
) (htypes.ReviewSummary, error) {
	query := fmt.Sprintf(`
		SELECT
		    COUNT(*)::int,
		    COALESCE(AVG(rating), 0)::float8,
		    COUNT(*) FILTER (WHERE rating = 1)::int,
		    COUNT(*) FILTER (WHERE rating = 2)::int,
		    COUNT(*) FILTER (WHERE rating = 3)::int,
		    COUNT(*) FILTER (WHERE rating = 4)::int,
		    COUNT(*) FILTER (WHERE rating = 5)::int
		FROM deal_reviews
		WHERE %s = $1`, column)

	var s htypes.ReviewSummary
	err := exec.QueryRow(ctx, query, id).Scan(
		&s.Count, &s.AvgRating,
		&s.Rating1, &s.Rating2, &s.Rating3, &s.Rating4, &s.Rating5,
	)
	if err != nil {
		return htypes.ReviewSummary{}, fmt.Errorf("sql get reviews summary by %s: %w", column, err)
	}
	return s, nil
}

// ================================================================================
// GET REVIEWS BY AUTHOR ID
// ================================================================================

func (r *Repository) GetReviewsByAuthorID(
	ctx context.Context,
	exec db.DB,
	authorID uuid.UUID,
) ([]domain.Review, error) {
	query := `SELECT ` + reviewSelectCols + `
		FROM deal_reviews
		WHERE author_id = $1
		ORDER BY created_at DESC`

	rows, err := exec.Query(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("sql get reviews by author id: %w", err)
	}
	defer rows.Close()
	return scanReviews(rows)
}

// ================================================================================
// GET REVIEWS BY DEAL ID
// ================================================================================

func (r *Repository) GetReviewsByDealID(
	ctx context.Context,
	exec db.DB,
	dealID uuid.UUID,
) ([]domain.Review, error) {
	query := `SELECT ` + reviewSelectCols + `
		FROM deal_reviews
		WHERE deal_id = $1
		ORDER BY created_at DESC`

	rows, err := exec.Query(ctx, query, dealID)
	if err != nil {
		return nil, fmt.Errorf("sql get reviews by deal id: %w", err)
	}
	defer rows.Close()
	return scanReviews(rows)
}

// ================================================================================
// GET REVIEWS BY DEAL ITEM ID
// ================================================================================

func (r *Repository) GetReviewsByDealItemID(
	ctx context.Context,
	exec db.DB,
	dealID, itemID uuid.UUID,
) ([]domain.Review, error) {
	query := `SELECT ` + reviewSelectCols + `
		FROM deal_reviews
		WHERE deal_id = $1 AND item_id = $2
		ORDER BY created_at DESC`

	rows, err := exec.Query(ctx, query, dealID, itemID)
	if err != nil {
		return nil, fmt.Errorf("sql get reviews by deal item id: %w", err)
	}
	defer rows.Close()
	return scanReviews(rows)
}

// ================================================================================
// GET ITEM FOR REVIEW
// ================================================================================

// GetItemForReview returns an item by deal and item ID, used for review business logic.
// Returns ErrItemNotFound if the item does not exist in the deal.
func (r *Repository) GetItemForReview(
	ctx context.Context,
	exec db.DB,
	dealID, itemID uuid.UUID,
) (domain.Item, error) {
	query := `
		SELECT id, offer_id, author_id, provider_id, receiver_id,
		       name, description, type, updated_at, quantity
		FROM items
		WHERE deal_id = $1 AND id = $2`

	item, err := scanItem(exec.QueryRow(ctx, query, dealID, itemID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, domain.ErrItemNotFound
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("sql get item for review: %w", err)
	}
	return item, nil
}

// ================================================================================
// REVIEW EXISTS FOR USER
// ================================================================================

// ReviewExistsForUser checks whether a review already exists for a specific context and user.
// Uses IS NOT DISTINCT FROM to correctly handle NULL comparisons for itemID and offerID.
func (r *Repository) ReviewExistsForUser(
	ctx context.Context,
	exec db.DB,
	dealID, authorID uuid.UUID,
	itemID, offerID *uuid.UUID,
	providerID uuid.UUID,
) (bool, error) {
	query := `
		SELECT EXISTS (
		    SELECT 1 FROM deal_reviews
		    WHERE deal_id    = $1
		      AND author_id  = $2
		      AND (item_id   IS NOT DISTINCT FROM $3)
		      AND (offer_id  IS NOT DISTINCT FROM $4)
		      AND provider_id = $5
		)`

	var exists bool
	err := exec.QueryRow(ctx, query, dealID, authorID, itemID, offerID, providerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("sql review exists for user: %w", err)
	}
	return exists, nil
}

// ================================================================================
// GET PENDING REVIEWS FOR PARTICIPANT
// ================================================================================

// GetPendingReviewsForParticipant returns all review eligibility entries for items
// where the given user is the receiver in a deal. Uses a single query to avoid N+1.
func (r *Repository) GetPendingReviewsForParticipant(
	ctx context.Context,
	exec db.DB,
	dealID, userID uuid.UUID,
) ([]htypes.ReviewEligibility, error) {
	query := `
		SELECT
		    i.id          AS item_id,
		    i.offer_id,
		    i.provider_id,
		    CASE
		        WHEN i.offer_id IS NULL                               THEN 'item-only'
		        WHEN i.offer_id IS NOT NULL AND i.updated_at IS NULL  THEN 'offer-only'
		        ELSE 'offer+item'
		    END AS context_type,
		    EXISTS (
		        SELECT 1 FROM deal_reviews dr
		        WHERE dr.deal_id   = $1
		          AND dr.author_id = $2
		          AND (
		              (i.offer_id IS NULL
		               AND dr.item_id = i.id AND dr.offer_id IS NULL)
		              OR (i.offer_id IS NOT NULL AND i.updated_at IS NULL
		                  AND dr.offer_id = i.offer_id AND dr.item_id IS NULL
		                  AND dr.provider_id = i.provider_id)
		              OR (i.offer_id IS NOT NULL AND i.updated_at IS NOT NULL
		                  AND dr.offer_id = i.offer_id AND dr.item_id = i.id
		                  AND dr.provider_id = i.provider_id)
		          )
		    ) AS already_reviewed
		FROM items i
		WHERE i.deal_id    = $1
		  AND i.receiver_id = $2
		ORDER BY i.id`

	rows, err := exec.Query(ctx, query, dealID, userID)
	if err != nil {
		return nil, fmt.Errorf("sql get pending reviews for participant: %w", err)
	}
	defer rows.Close()

	var result []htypes.ReviewEligibility
	for rows.Next() {
		var (
			itemID          uuid.UUID
			offerID         *uuid.UUID
			providerID      *uuid.UUID
			contextType     string
			alreadyReviewed bool
		)
		if err = rows.Scan(&itemID, &offerID, &providerID, &contextType, &alreadyReviewed); err != nil {
			return nil, fmt.Errorf("scan pending review row: %w", err)
		}

		e := htypes.ReviewEligibility{
			ContextType: types.ReviewContextType(contextType),
			ProviderID:  providerID,
			OfferID:     offerID,
			DealID:      dealID,
		}

		// itemRef is set for item-only and offer+item contexts
		if contextType == string(types.ItemOnly) || contextType == string(types.OfferItem) {
			e.ItemID = &itemID
		}

		switch {
		case providerID == nil:
			r := types.ProviderMissing
			e.Reason = &r
		case *providerID == userID:
			r := types.SameProviderAndReceiver
			e.Reason = &r
		case alreadyReviewed:
			r := types.AlreadyReviewed
			e.Reason = &r
		default:
			e.CanCreate = true
		}

		result = append(result, e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows pending reviews: %w", err)
	}
	if result == nil {
		result = []htypes.ReviewEligibility{}
	}
	return result, nil
}
