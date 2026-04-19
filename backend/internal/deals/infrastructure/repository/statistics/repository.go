package statistics

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

type Result struct {
	DealsCompleted int
	DealsFailed    int
	DealsActive    int

	OffersTotal      int
	OffersTotalViews int64

	ReviewsWritten               int
	ReviewsReceived              int
	ReviewsAverageRatingReceived *float64

	ReportsOnMyOffersTotal    int
	ReportsOnMyOffersPending  int
	ReportsOnMyOffersAccepted int
	ReportsOnMyOffersRejected int
	ReportsFiledByMe          int
}

func (r *Repository) GetMyStatistics(ctx context.Context, userID uuid.UUID) (*Result, error) {
	res := &Result{}

	if err := r.queryDeals(ctx, userID, res); err != nil {
		return nil, fmt.Errorf("query deals: %w", err)
	}
	if err := r.queryOffers(ctx, userID, res); err != nil {
		return nil, fmt.Errorf("query offers: %w", err)
	}
	if err := r.queryReviews(ctx, userID, res); err != nil {
		return nil, fmt.Errorf("query reviews: %w", err)
	}
	if err := r.queryReports(ctx, userID, res); err != nil {
		return nil, fmt.Errorf("query reports: %w", err)
	}

	return res, nil
}

func (r *Repository) queryDeals(ctx context.Context, userID uuid.UUID, res *Result) error {
	const q = `
		SELECT
			COUNT(*) FILTER (WHERE d.status = 'Completed') AS completed,
			COUNT(*) FILTER (WHERE d.status = 'Failed')    AS failed,
			COUNT(*) FILTER (WHERE d.status IN ('LookingForParticipants', 'Discussion', 'Confirmed')) AS active
		FROM participants p
		JOIN deals d ON d.id = p.deal_id
		WHERE p.user_id = $1`

	return r.db.QueryRow(ctx, q, userID).Scan(
		&res.DealsCompleted,
		&res.DealsFailed,
		&res.DealsActive,
	)
}

func (r *Repository) queryOffers(ctx context.Context, userID uuid.UUID, res *Result) error {
	const q = `
		SELECT
			COUNT(*)            AS total,
			COALESCE(SUM(views), 0) AS total_views
		FROM offers
		WHERE author_id = $1`

	return r.db.QueryRow(ctx, q, userID).Scan(
		&res.OffersTotal,
		&res.OffersTotalViews,
	)
}

func (r *Repository) queryReviews(ctx context.Context, userID uuid.UUID, res *Result) error {
	const q = `
		SELECT
			COUNT(*) FILTER (WHERE author_id   = $1) AS written,
			COUNT(*) FILTER (WHERE provider_id = $1) AS received,
			AVG(rating)  FILTER (WHERE provider_id = $1) AS avg_rating
		FROM deal_reviews
		WHERE author_id = $1 OR provider_id = $1`

	return r.db.QueryRow(ctx, q, userID).Scan(
		&res.ReviewsWritten,
		&res.ReviewsReceived,
		&res.ReviewsAverageRatingReceived,
	)
}

func (r *Repository) queryReports(ctx context.Context, userID uuid.UUID, res *Result) error {
	const qOnMyOffers = `
		SELECT
			COUNT(*)                                          AS total,
			COUNT(*) FILTER (WHERE status = 'Pending')   AS pending,
			COUNT(*) FILTER (WHERE status = 'Accepted')  AS accepted,
			COUNT(*) FILTER (WHERE status = 'Rejected')  AS rejected
		FROM offer_reports
		WHERE offer_author_id = $1`

	if err := r.db.QueryRow(ctx, qOnMyOffers, userID).Scan(
		&res.ReportsOnMyOffersTotal,
		&res.ReportsOnMyOffersPending,
		&res.ReportsOnMyOffersAccepted,
		&res.ReportsOnMyOffersRejected,
	); err != nil {
		return err
	}

	const qFiled = `
		SELECT COUNT(DISTINCT offer_report_id)
		FROM offer_reports_messages
		WHERE author_id = $1`

	return r.db.QueryRow(ctx, qFiled, userID).Scan(&res.ReportsFiledByMe)
}
