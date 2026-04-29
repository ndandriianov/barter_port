package offers

import (
	"barter-port/internal/deals/domain"
	"barter-port/pkg/repox"
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
)

const rowsToSelect = `
	offers.id, offers.author_id, offers.name, offers.type, offers.action, offers.description, offers.created_at, offers.updated_at, offers.views,
	COALESCE((SELECT array_agg(ot.tag_name ORDER BY ot.tag_name) FROM offer_tags ot WHERE ot.offer_id = offers.id), '{}'::text[]) AS tags,
	COALESCE((SELECT array_agg(op.id ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = offers.id), '{}'::uuid[]) AS photo_ids,
	COALESCE((SELECT array_agg(op.url ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = offers.id), '{}'::text[]) AS photo_urls,
	offers.is_hidden, offers.hidden_by_author, offers.modification_blocked, offers.latitude, offers.longitude`

const tagFilterClause = `
	AND (
		NOT $%d
		OR (
			cardinality($%d::text[]) = 0
			AND NOT EXISTS (
				SELECT 1
				FROM offer_tags ot
				WHERE ot.offer_id = offers.id
			)
		)
		OR (
			cardinality($%d::text[]) > 0
			AND (
				SELECT COUNT(DISTINCT ot.tag_name)
				FROM offer_tags ot
				WHERE ot.offer_id = offers.id
				  AND ot.tag_name = ANY($%d::text[])
			) = cardinality($%d::text[])
		)
	)
`

// ================================================================================
// HELPERS
// ================================================================================

func timeCursor(offer domain.Offer) domain.TimeCursor {
	return domain.TimeCursor{
		CreatedAt: offer.CreatedAt,
		Id:        offer.ID,
	}
}

func offersAndTimeCursor(offers []domain.Offer, err error) ([]domain.Offer, *domain.TimeCursor, error) {
	if err != nil {
		return nil, nil, err
	}

	if len(offers) == 0 {
		return offers, nil, nil
	}
	lastOffer := offers[len(offers)-1]
	return offers, new(timeCursor(lastOffer)), nil
}

func popularityCursor(offer domain.Offer) domain.PopularityCursor {
	return domain.PopularityCursor{
		Views: offer.Views,
		Id:    offer.ID,
	}
}

func offersAndPopularityCursor(offers []domain.Offer, err error) ([]domain.Offer, *domain.PopularityCursor, error) {
	if err != nil {
		return nil, nil, err
	}

	if len(offers) == 0 {
		return offers, nil, nil
	}
	lastOffer := offers[len(offers)-1]
	return offers, new(popularityCursor(lastOffer)), nil
}

func distanceExpr(latParam, lonParam int) string {
	return fmt.Sprintf(`
		2 * 6371000 * asin(
			sqrt(
				power(sin(radians((offers.latitude - $%d) / 2)), 2) +
				cos(radians($%d)) * cos(radians(offers.latitude)) *
				power(sin(radians((offers.longitude - $%d) / 2)), 2)
			)
		)
	`, latParam, latParam, lonParam)
}

func (r *Repository) scanDistanceOffers(ctx context.Context, query string, args ...interface{}) ([]domain.Offer, *domain.DistanceCursor, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("sql get distance offers: %w", err)
	}
	defer rows.Close()

	offers := make([]domain.Offer, 0)
	var nextCursor *domain.DistanceCursor

	for rows.Next() {
		var (
			offer    domain.Offer
			distance float64
		)

		if err := rows.Scan(
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
			&offer.HiddenByAuthor,
			&offer.ModificationBlocked,
			&offer.Latitude,
			&offer.Longitude,
			&distance,
		); err != nil {
			return nil, nil, fmt.Errorf("scan distance offer: %w", err)
		}

		offer.DistanceMeters = new(int64(math.Round(distance)))
		offers = append(offers, offer)
		nextCursor = &domain.DistanceCursor{
			Distance: distance,
			Id:       offer.ID,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate distance offers: %w", err)
	}

	if len(offers) == 0 {
		return offers, nil, nil
	}

	return offers, nextCursor, nil
}

func favoriteOffersCursor(offer domain.FavoritedOffer) domain.FavoriteOffersCursor {
	return domain.FavoriteOffersCursor{
		FavoritedAt: offer.FavoritedAt,
		Id:          offer.ID,
	}
}

func offersAndFavoriteCursor(offers []domain.FavoritedOffer, err error) ([]domain.FavoritedOffer, *domain.FavoriteOffersCursor, error) {
	if err != nil {
		return nil, nil, err
	}

	if len(offers) == 0 {
		return offers, nil, nil
	}

	lastOffer := offers[len(offers)-1]
	return offers, new(favoriteOffersCursor(lastOffer)), nil
}

// ================================================================================
// ПО РАССТОЯНИЮ
// ================================================================================

func (r *Repository) GetOffersOrderByDistance(
	ctx context.Context,
	limit int,
	cursor domain.DistanceCursor,
	isAdmin bool,
	userLat float64,
	userLon float64,
	tags []string,
	tagsFilterPresent bool,
	hiddenAuthorIDs []uuid.UUID,
) ([]domain.Offer, *domain.DistanceCursor, error) {
	query := `
		WITH filtered_offers AS (
			SELECT ` + rowsToSelect + `,
			       ` + distanceExpr(2, 3) + ` AS distance_cursor
			FROM offers
			WHERE offers.latitude IS NOT NULL
			  AND offers.longitude IS NOT NULL
			  AND (NOT (offers.is_hidden OR offers.hidden_by_author) OR $6)
			  AND NOT (offers.author_id = ANY($7))` + fmt.Sprintf(tagFilterClause, 9, 8, 8, 8, 8) + `
		)
		SELECT *
		FROM filtered_offers
		WHERE distance_cursor > $4
		   OR (distance_cursor = $4 AND id > $5)
		ORDER BY distance_cursor ASC, id ASC
		LIMIT $1`

	return r.scanDistanceOffers(ctx, query, limit, userLat, userLon, cursor.Distance, cursor.Id, isAdmin, hiddenAuthorIDs, tags, tagsFilterPresent)
}

func (r *Repository) GetOffersOrderByDistanceNoCursor(
	ctx context.Context,
	limit int,
	isAdmin bool,
	userLat float64,
	userLon float64,
	tags []string,
	tagsFilterPresent bool,
	hiddenAuthorIDs []uuid.UUID,
) ([]domain.Offer, *domain.DistanceCursor, error) {
	query := `
		WITH filtered_offers AS (
			SELECT ` + rowsToSelect + `,
			       ` + distanceExpr(2, 3) + ` AS distance_cursor
			FROM offers
			WHERE offers.latitude IS NOT NULL
			  AND offers.longitude IS NOT NULL
			  AND (NOT (offers.is_hidden OR offers.hidden_by_author) OR $4)
			  AND NOT (offers.author_id = ANY($5))` + fmt.Sprintf(tagFilterClause, 7, 6, 6, 6, 6) + `
		)
		SELECT *
		FROM filtered_offers
		ORDER BY distance_cursor ASC, id ASC
		LIMIT $1`

	return r.scanDistanceOffers(ctx, query, limit, userLat, userLon, isAdmin, hiddenAuthorIDs, tags, tagsFilterPresent)
}

func (r *Repository) GetMyOffersOrderByDistance(
	ctx context.Context,
	limit int,
	cursor domain.DistanceCursor,
	userID uuid.UUID,
	userLat float64,
	userLon float64,
	tags []string,
	tagsFilterPresent bool,
) ([]domain.Offer, *domain.DistanceCursor, error) {
	query := `
		WITH filtered_offers AS (
			SELECT ` + rowsToSelect + `,
			       ` + distanceExpr(2, 3) + ` AS distance_cursor
			FROM offers
			WHERE offers.latitude IS NOT NULL
			  AND offers.longitude IS NOT NULL
			  AND offers.author_id = $6` + fmt.Sprintf(tagFilterClause, 8, 7, 7, 7, 7) + `
		)
		SELECT *
		FROM filtered_offers
		WHERE distance_cursor > $4
		   OR (distance_cursor = $4 AND id > $5)
		ORDER BY distance_cursor ASC, id ASC
		LIMIT $1`

	return r.scanDistanceOffers(ctx, query, limit, userLat, userLon, cursor.Distance, cursor.Id, userID, tags, tagsFilterPresent)
}

func (r *Repository) GetMyOffersOrderByDistanceNoCursor(
	ctx context.Context,
	limit int,
	userID uuid.UUID,
	userLat float64,
	userLon float64,
	tags []string,
	tagsFilterPresent bool,
) ([]domain.Offer, *domain.DistanceCursor, error) {
	query := `
		WITH filtered_offers AS (
			SELECT ` + rowsToSelect + `,
			       ` + distanceExpr(2, 3) + ` AS distance_cursor
			FROM offers
			WHERE offers.latitude IS NOT NULL
			  AND offers.longitude IS NOT NULL
			  AND offers.author_id = $4` + fmt.Sprintf(tagFilterClause, 6, 5, 5, 5, 5) + `
		)
		SELECT *
		FROM filtered_offers
		ORDER BY distance_cursor ASC, id ASC
		LIMIT $1`

	return r.scanDistanceOffers(ctx, query, limit, userLat, userLon, userID, tags, tagsFilterPresent)
}

func (r *Repository) GetSubscribedOffersOrderByDistance(
	ctx context.Context,
	limit int,
	cursor domain.DistanceCursor,
	authorIDs []uuid.UUID,
	isAdmin bool,
	userLat float64,
	userLon float64,
) ([]domain.Offer, *domain.DistanceCursor, error) {
	query := `
		WITH filtered_offers AS (
			SELECT ` + rowsToSelect + `,
			       ` + distanceExpr(2, 3) + ` AS distance_cursor
			FROM offers
			WHERE offers.latitude IS NOT NULL
			  AND offers.longitude IS NOT NULL
			  AND offers.author_id = ANY($6)
			  AND (NOT (offers.is_hidden OR offers.hidden_by_author) OR $7)
		)
		SELECT *
		FROM filtered_offers
		WHERE distance_cursor > $4
		   OR (distance_cursor = $4 AND id > $5)
		ORDER BY distance_cursor ASC, id ASC
		LIMIT $1`

	return r.scanDistanceOffers(ctx, query, limit, userLat, userLon, cursor.Distance, cursor.Id, authorIDs, isAdmin)
}

func (r *Repository) GetSubscribedOffersOrderByDistanceNoCursor(
	ctx context.Context,
	limit int,
	authorIDs []uuid.UUID,
	isAdmin bool,
	userLat float64,
	userLon float64,
) ([]domain.Offer, *domain.DistanceCursor, error) {
	query := `
		WITH filtered_offers AS (
			SELECT ` + rowsToSelect + `,
			       ` + distanceExpr(2, 3) + ` AS distance_cursor
			FROM offers
			WHERE offers.latitude IS NOT NULL
			  AND offers.longitude IS NOT NULL
			  AND offers.author_id = ANY($4)
			  AND (NOT (offers.is_hidden OR offers.hidden_by_author) OR $5)
		)
		SELECT *
		FROM filtered_offers
		ORDER BY distance_cursor ASC, id ASC
		LIMIT $1`

	return r.scanDistanceOffers(ctx, query, limit, userLat, userLon, authorIDs, isAdmin)
}

// ================================================================================
// ПО ВРЕМЕНИ
// ================================================================================

func (r *Repository) GetOffersOrderByTime(
	ctx context.Context,
	limit int,
	cursor domain.TimeCursor,
	isAdmin bool,
	tags []string,
	tagsFilterPresent bool,
	hiddenAuthorIDs []uuid.UUID,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (created_at, id) < ($2, $3)
			AND (NOT (is_hidden OR hidden_by_author) OR $4)
			AND NOT (author_id = ANY($5))` + fmt.Sprintf(tagFilterClause, 7, 6, 6, 6, 6) + `
		ORDER BY created_at DESC, id DESC
		LIMIT $1`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, cursor.CreatedAt, cursor.Id, isAdmin, hiddenAuthorIDs, tags, tagsFilterPresent))
}

func (r *Repository) GetOffersOrderByTimeNoCursor(
	ctx context.Context,
	limit int,
	isAdmin bool,
	tags []string,
	tagsFilterPresent bool,
	hiddenAuthorIDs []uuid.UUID,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (NOT (is_hidden OR hidden_by_author) OR $2)
		  AND NOT (author_id = ANY($3))` + fmt.Sprintf(tagFilterClause, 5, 4, 4, 4, 4) + `
		ORDER BY created_at DESC, id DESC
		LIMIT $1`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, isAdmin, hiddenAuthorIDs, tags, tagsFilterPresent))
}

// my

func (r *Repository) GetMyOffersOrderByTime(
	ctx context.Context,
	cursor domain.TimeCursor,
	userID uuid.UUID,
	limit int,
	tags []string,
	tagsFilterPresent bool,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (created_at, id) < ($1, $2) AND author_id = $3` + fmt.Sprintf(tagFilterClause, 6, 5, 5, 5, 5) + `
		ORDER BY created_at DESC, id DESC
		LIMIT $4`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, cursor.CreatedAt, cursor.Id, userID, limit, tags, tagsFilterPresent))
}

func (r *Repository) GetMyOffersOrderByTimeNoCursor(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	tags []string,
	tagsFilterPresent bool,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE author_id = $1` + fmt.Sprintf(tagFilterClause, 4, 3, 3, 3, 3) + `
		ORDER BY created_at DESC, id DESC
		LIMIT $2`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, userID, limit, tags, tagsFilterPresent))
}

// subscribed

func (r *Repository) GetSubscribedOffersOrderByTime(
	ctx context.Context,
	limit int,
	cursor domain.TimeCursor,
	authorIDs []uuid.UUID,
	isAdmin bool,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (created_at, id) < ($2, $3)
			AND author_id = ANY($4)
			AND (NOT (is_hidden OR hidden_by_author) OR $5)
		ORDER BY created_at DESC, id DESC
		LIMIT $1`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, cursor.CreatedAt, cursor.Id, authorIDs, isAdmin))
}

func (r *Repository) GetSubscribedOffersOrderByTimeNoCursor(
	ctx context.Context,
	limit int,
	authorIDs []uuid.UUID,
	isAdmin bool,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE author_id = ANY($2)
			AND (NOT (is_hidden OR hidden_by_author) OR $3)
		ORDER BY created_at DESC, id DESC
		LIMIT $1`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, authorIDs, isAdmin))
}

// ================================================================================
// ПО ПОПУЛЯРНОСТИ
// ================================================================================

func (r *Repository) GetOffersOrderByPopularity(
	ctx context.Context,
	limit int,
	cursor domain.PopularityCursor,
	isAdmin bool,
	tags []string,
	tagsFilterPresent bool,
	hiddenAuthorIDs []uuid.UUID,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (views, id) < ($2, $3)
			AND (NOT (is_hidden OR hidden_by_author) OR $4)
			AND NOT (author_id = ANY($5))` + fmt.Sprintf(tagFilterClause, 7, 6, 6, 6, 6) + `
		ORDER BY views DESC, id DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, cursor.Views, cursor.Id, isAdmin, hiddenAuthorIDs, tags, tagsFilterPresent))
}

func (r *Repository) GetOffersOrderByPopularityNoCursor(
	ctx context.Context,
	limit int,
	isAdmin bool,
	tags []string,
	tagsFilterPresent bool,
	hiddenAuthorIDs []uuid.UUID,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (NOT (is_hidden OR hidden_by_author) OR $2)
		  AND NOT (author_id = ANY($3))` + fmt.Sprintf(tagFilterClause, 5, 4, 4, 4, 4) + `
		ORDER BY views DESC, id DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, isAdmin, hiddenAuthorIDs, tags, tagsFilterPresent))
}

// my

func (r *Repository) GetMyOffersOrderByPopularity(
	ctx context.Context,
	limit int,
	cursor domain.PopularityCursor,
	userID uuid.UUID,
	tags []string,
	tagsFilterPresent bool,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (views, id) < ($2, $3) AND author_id = $4` + fmt.Sprintf(tagFilterClause, 6, 5, 5, 5, 5) + `
		ORDER BY views DESC, id DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, cursor.Views, cursor.Id, userID, tags, tagsFilterPresent))
}

func (r *Repository) GetMyOffersOrderByPopularityNoCursor(
	ctx context.Context,
	limit int,
	userID uuid.UUID,
	tags []string,
	tagsFilterPresent bool,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE author_id = $2` + fmt.Sprintf(tagFilterClause, 4, 3, 3, 3, 3) + `
		ORDER BY views DESC, id DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, userID, tags, tagsFilterPresent))
}

// subscribed

func (r *Repository) GetSubscribedOffersOrderByPopularity(
	ctx context.Context,
	limit int,
	cursor domain.PopularityCursor,
	authorIDs []uuid.UUID,
	isAdmin bool,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (views, id) < ($2, $3)
			AND author_id = ANY($4)
			AND (NOT (is_hidden OR hidden_by_author) OR $5)
		ORDER BY views DESC, id DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, cursor.Views, cursor.Id, authorIDs, isAdmin))
}

func (r *Repository) GetSubscribedOffersOrderByPopularityNoCursor(
	ctx context.Context,
	limit int,
	authorIDs []uuid.UUID,
	isAdmin bool,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE author_id = ANY($2)
			AND (NOT (is_hidden OR hidden_by_author) OR $3)
		ORDER BY views DESC, id DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, authorIDs, isAdmin))
}

func (r *Repository) GetFavoriteOffers(
	ctx context.Context,
	userID uuid.UUID,
	cursor domain.FavoriteOffersCursor,
	limit int,
	isAdmin bool,
) ([]domain.FavoritedOffer, *domain.FavoriteOffersCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `,
		       fo.created_at AS favorited_at
		FROM favorite_offers fo
		JOIN offers ON offers.id = fo.offer_id
		WHERE fo.user_id = $1
		  AND (fo.created_at, fo.offer_id) < ($2, $3)
		  AND (NOT (offers.is_hidden OR offers.hidden_by_author) OR offers.author_id = $1 OR $5)
		ORDER BY fo.created_at DESC, fo.offer_id DESC
		LIMIT $4`

	return offersAndFavoriteCursor(r.scanFavoriteOffers(ctx, query, userID, cursor.FavoritedAt, cursor.Id, limit, isAdmin))
}

func (r *Repository) GetFavoriteOffersNoCursor(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	isAdmin bool,
) ([]domain.FavoritedOffer, *domain.FavoriteOffersCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `,
		       fo.created_at AS favorited_at
		FROM favorite_offers fo
		JOIN offers ON offers.id = fo.offer_id
		WHERE fo.user_id = $1
		  AND (NOT (offers.is_hidden OR offers.hidden_by_author) OR offers.author_id = $1 OR $3)
		ORDER BY fo.created_at DESC, fo.offer_id DESC
		LIMIT $2`

	return offersAndFavoriteCursor(r.scanFavoriteOffers(ctx, query, userID, limit, isAdmin))
}

func (r *Repository) scanFavoriteOffers(ctx context.Context, query string, args ...interface{}) ([]domain.FavoritedOffer, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("sql get favorite offers: %w", err)
	}
	defer rows.Close()

	offers := make([]domain.FavoritedOffer, 0)
	for rows.Next() {
		var offer domain.FavoritedOffer
		if err := rows.Scan(
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
			&offer.HiddenByAuthor,
			&offer.ModificationBlocked,
			&offer.Latitude,
			&offer.Longitude,
			&offer.FavoritedAt,
		); err != nil {
			return nil, fmt.Errorf("scan favorite offer: %w", err)
		}
		offers = append(offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate favorite offers: %w", err)
	}

	return offers, nil
}
