package offers

import (
	"barter-port/internal/deals/domain"
	"barter-port/pkg/repox"
	"context"

	"github.com/google/uuid"
)

const rowsToSelect = `
	id, author_id, name, type, action, description, created_at, updated_at, views,
	COALESCE((SELECT array_agg(op.id ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = offers.id), '{}'::uuid[]) AS photo_ids,
	COALESCE((SELECT array_agg(op.url ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = offers.id), '{}'::text[]) AS photo_urls,
	is_hidden, modification_blocked`

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

// ================================================================================
// ПО ВРЕМЕНИ
// ================================================================================

func (r *Repository) GetOffersOrderByTime(
	ctx context.Context,
	limit int,
	cursor domain.TimeCursor,
	isAdmin bool,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (created_at, id) < ($2, $3)
			AND (is_hidden = FALSE OR $4) 
		ORDER BY created_at DESC
		LIMIT $1`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, cursor.CreatedAt, cursor.Id, isAdmin))
}

func (r *Repository) GetOffersOrderByTimeNoCursor(
	ctx context.Context,
	limit int,
	isAdmin bool,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (is_hidden = FALSE OR $2)
		ORDER BY created_at DESC
		LIMIT $1`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, isAdmin))
}

// my

func (r *Repository) GetMyOffersOrderByTime(
	ctx context.Context,
	cursor domain.TimeCursor,
	userID uuid.UUID,
	limit int,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (created_at, id) < ($1, $2) AND author_id = $3
		ORDER BY created_at DESC
		LIMIT $4`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, cursor.CreatedAt, cursor.Id, userID, limit))
}

func (r *Repository) GetMyOffersOrderByTimeNoCursor(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]domain.Offer, *domain.TimeCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE author_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	return offersAndTimeCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, userID, limit))
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
		ORDER BY created_at DESC
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
		ORDER BY created_at DESC
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
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (created_at, id) < ($2, $3)
			AND (is_hidden = FALSE OR $4) 
		ORDER BY views DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, cursor.Views, cursor.Id, isAdmin))
}

func (r *Repository) GetOffersOrderByPopularityNoCursor(
	ctx context.Context,
	limit int,
	isAdmin bool,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (is_hidden = FALSE OR $2) 
		ORDER BY views DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, isAdmin))
}

// my

func (r *Repository) GetMyOffersOrderByPopularity(
	ctx context.Context,
	limit int,
	cursor domain.PopularityCursor,
	userID uuid.UUID,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE (created_at, id) < ($2, $3) AND author_id = $4
		ORDER BY views DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, cursor.Views, cursor.Id, userID))
}

func (r *Repository) GetMyOffersOrderByPopularityNoCursor(
	ctx context.Context,
	limit int,
	userID uuid.UUID,
) ([]domain.Offer, *domain.PopularityCursor, error) {
	query := `
		SELECT ` + rowsToSelect + `
		FROM offers
		WHERE author_id = $2
		ORDER BY views DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, userID))
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
		ORDER BY views DESC
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
		ORDER BY views DESC
		LIMIT $1`

	return offersAndPopularityCursor(repox.FetchStructs[domain.Offer](ctx, r.db, query, limit, authorIDs, isAdmin))
}
