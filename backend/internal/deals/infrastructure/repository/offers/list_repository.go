package offers

import (
	"barter-port/internal/deals/domain"
	"barter-port/pkg/repox"
	"context"
	"fmt"

	"github.com/google/uuid"
)

const rowsToSelect = `
	offers.id, offers.author_id, offers.name, offers.type, offers.action, offers.description, offers.created_at, offers.updated_at, offers.views,
	COALESCE((SELECT array_agg(ot.tag_name ORDER BY ot.tag_name) FROM offer_tags ot WHERE ot.offer_id = offers.id), '{}'::text[]) AS tags,
	COALESCE((SELECT array_agg(op.id ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = offers.id), '{}'::uuid[]) AS photo_ids,
	COALESCE((SELECT array_agg(op.url ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = offers.id), '{}'::text[]) AS photo_urls,
	offers.is_hidden, offers.modification_blocked`

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
	nextCursor := timeCursor(lastOffer)

	return offers, &nextCursor, nil
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
	nextCursor := popularityCursor(lastOffer)

	return offers, &nextCursor, nil
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
	nextCursor := favoriteOffersCursor(lastOffer)

	return offers, &nextCursor, nil
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
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (created_at, id) < ($2, $3)
			AND (is_hidden = FALSE OR $4)` + fmt.Sprintf(tagFilterClause, 6, 5, 5, 5, 5) + `
		ORDER BY created_at DESC, id DESC
		LIMIT $1`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, cursor.CreatedAt, cursor.Id, isAdmin, tags, tagsFilterPresent))
}

func (r *Repository) GetOffersOrderByTimeNoCursor(
	ctx context.Context,
	limit int,
	isAdmin bool,
	tags []string,
	tagsFilterPresent bool,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (is_hidden = FALSE OR $2)` + fmt.Sprintf(tagFilterClause, 4, 3, 3, 3, 3) + `
		ORDER BY created_at DESC, id DESC
		LIMIT $1`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, isAdmin, tags, tagsFilterPresent))
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
			AND (is_hidden = FALSE OR $5)
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
			AND (is_hidden = FALSE OR $3)
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
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (views, id) < ($2, $3)
			AND (is_hidden = FALSE OR $4)` + fmt.Sprintf(tagFilterClause, 6, 5, 5, 5, 5) + `
		ORDER BY views DESC, id DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, cursor.Views, cursor.Id, isAdmin, tags, tagsFilterPresent))
}

func (r *Repository) GetOffersOrderByPopularityNoCursor(
	ctx context.Context,
	limit int,
	isAdmin bool,
	tags []string,
	tagsFilterPresent bool,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (is_hidden = FALSE OR $2)` + fmt.Sprintf(tagFilterClause, 4, 3, 3, 3, 3) + `
		ORDER BY views DESC, id DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, isAdmin, tags, tagsFilterPresent))
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
			AND (is_hidden = FALSE OR $5)
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
			AND (is_hidden = FALSE OR $3)
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
		  AND (offers.is_hidden = FALSE OR offers.author_id = $1 OR $5)
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
		  AND (offers.is_hidden = FALSE OR offers.author_id = $1 OR $3)
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
			&offer.ModificationBlocked,
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
