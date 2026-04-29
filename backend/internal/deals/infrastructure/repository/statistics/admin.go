package statistics

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type AdminPlatformStatisticsResult struct {
	Offers  AdminPlatformOfferStatisticsResult
	Deals   AdminPlatformDealStatisticsResult
	Reports AdminPlatformReportStatisticsResult
	Reviews AdminPlatformReviewStatisticsResult
}

type AdminPlatformOfferStatisticsResult struct {
	Total          int
	Hidden         AdminHiddenOfferStatisticsResult
	ByType         AdminOfferTypeDistributionResult
	ByAction       AdminOfferActionDistributionResult
	TopTags        []AdminTopTagStatResult
	AveragePerUser float64
	Drafts         int
	TotalViews     int64
	TopByFavorites []AdminTopFavoriteOfferStatResult
	AverageRating  *float64
}

type AdminHiddenOfferStatisticsResult struct {
	Moderated      int
	HiddenByAuthor int
}

type AdminOfferTypeDistributionResult struct {
	Good    int
	Service int
}

type AdminOfferActionDistributionResult struct {
	Give int
	Take int
}

type AdminTopTagStatResult struct {
	Tag         string
	OffersCount int
}

type AdminTopFavoriteOfferStatResult struct {
	OfferID        uuid.UUID
	AuthorID       uuid.UUID
	Name           string
	FavoritesCount int
}

type AdminPlatformDealStatisticsResult struct {
	Total                    int
	ByStatus                 AdminDealStatusDistributionResult
	SuccessfulConversionRate float64
	AverageParticipants      float64
	MultiPartyShare          float64
}

type AdminDealStatusDistributionResult struct {
	LookingForParticipants int
	Discussion             int
	Confirmed              int
	Completed              int
	Failed                 int
	Cancelled              int
}

type AdminPlatformReportStatisticsResult struct {
	Total                     int
	Pending                   int
	BlockedOffers             int
	AdminFailureResolutions   int
	TopUsersByReceivedReports []AdminTopReportedUserStatResult
}

type AdminTopReportedUserStatResult struct {
	UserID       uuid.UUID
	ReportsCount int
}

type AdminPlatformReviewStatisticsResult struct {
	Total              int
	AverageRating      *float64
	RatingDistribution AdminRatingDistributionResult
}

type AdminRatingDistributionResult struct {
	OneStar    int
	TwoStars   int
	ThreeStars int
	FourStars  int
	FiveStars  int
}

type AdminUserStatisticsResult struct {
	Deals   AdminUserDealStatisticsResult
	Offers  AdminUserOfferStatisticsResult
	Reviews AdminUserReviewStatisticsResult
	Reports AdminUserReportStatisticsResult
}

type AdminUserDealStatisticsResult struct {
	Completed int
	Active    int
	Failed    AdminUserFailedDealStatisticsResult
	Cancelled int
}

type AdminUserFailedDealStatisticsResult struct {
	Total       int
	Responsible int
	Affected    int
}

type AdminUserOfferStatisticsResult struct {
	Published  int
	TotalViews int64
}

type AdminUserReviewStatisticsResult struct {
	Received              int
	AverageReceivedRating *float64
	Written               int
}

type AdminUserReportStatisticsResult struct {
	Received AdminUserReceivedReportStatisticsResult
	Filed    int
}

type AdminUserReceivedReportStatisticsResult struct {
	Accepted int
	Rejected int
}

func (r *Repository) GetAdminPlatformStatistics(ctx context.Context) (*AdminPlatformStatisticsResult, error) {
	result := &AdminPlatformStatisticsResult{
		Offers: AdminPlatformOfferStatisticsResult{
			TopTags:        make([]AdminTopTagStatResult, 0),
			TopByFavorites: make([]AdminTopFavoriteOfferStatResult, 0),
		},
		Reports: AdminPlatformReportStatisticsResult{
			TopUsersByReceivedReports: make([]AdminTopReportedUserStatResult, 0),
		},
	}

	if err := r.queryAdminPlatformOffers(ctx, result); err != nil {
		return nil, fmt.Errorf("query admin platform offers: %w", err)
	}
	if err := r.queryAdminPlatformDeals(ctx, result); err != nil {
		return nil, fmt.Errorf("query admin platform deals: %w", err)
	}
	if err := r.queryAdminPlatformReports(ctx, result); err != nil {
		return nil, fmt.Errorf("query admin platform reports: %w", err)
	}
	if err := r.queryAdminPlatformReviews(ctx, result); err != nil {
		return nil, fmt.Errorf("query admin platform reviews: %w", err)
	}

	return result, nil
}

func (r *Repository) GetAdminUserStatistics(ctx context.Context, userID uuid.UUID) (*AdminUserStatisticsResult, error) {
	result := &AdminUserStatisticsResult{}

	if err := r.queryAdminUserDeals(ctx, userID, result); err != nil {
		return nil, fmt.Errorf("query admin user deals: %w", err)
	}
	if err := r.queryAdminUserOffers(ctx, userID, result); err != nil {
		return nil, fmt.Errorf("query admin user offers: %w", err)
	}
	if err := r.queryAdminUserReviews(ctx, userID, result); err != nil {
		return nil, fmt.Errorf("query admin user reviews: %w", err)
	}
	if err := r.queryAdminUserReports(ctx, userID, result); err != nil {
		return nil, fmt.Errorf("query admin user reports: %w", err)
	}

	return result, nil
}

func (r *Repository) queryAdminPlatformOffers(ctx context.Context, result *AdminPlatformStatisticsResult) error {
	const summaryQuery = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE is_hidden AND hidden_reason IS NOT NULL) AS moderated_hidden,
			COUNT(*) FILTER (WHERE is_hidden AND hidden_reason IS NULL) AS hidden_by_author,
			COUNT(*) FILTER (WHERE type = 'good') AS good_count,
			COUNT(*) FILTER (WHERE type = 'service') AS service_count,
			COUNT(*) FILTER (WHERE action = 'give') AS give_count,
			COUNT(*) FILTER (WHERE action = 'take') AS take_count,
			COALESCE(COUNT(*)::float8 / NULLIF(COUNT(DISTINCT author_id), 0), 0) AS average_per_user,
			COALESCE((SELECT COUNT(*) FROM draft_deals), 0) AS drafts,
			COALESCE(SUM(views), 0) AS total_views,
			(SELECT AVG(rating)::float8 FROM deal_reviews WHERE offer_id IS NOT NULL) AS average_rating
		FROM offers
	`

	if err := r.db.QueryRow(ctx, summaryQuery).Scan(
		&result.Offers.Total,
		&result.Offers.Hidden.Moderated,
		&result.Offers.Hidden.HiddenByAuthor,
		&result.Offers.ByType.Good,
		&result.Offers.ByType.Service,
		&result.Offers.ByAction.Give,
		&result.Offers.ByAction.Take,
		&result.Offers.AveragePerUser,
		&result.Offers.Drafts,
		&result.Offers.TotalViews,
		&result.Offers.AverageRating,
	); err != nil {
		return err
	}

	const topTagsQuery = `
		SELECT tag_name, COUNT(*) AS offers_count
		FROM offer_tags
		GROUP BY tag_name
		ORDER BY offers_count DESC, tag_name ASC
		LIMIT 10
	`

	rows, err := r.db.Query(ctx, topTagsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var item AdminTopTagStatResult
		if err = rows.Scan(&item.Tag, &item.OffersCount); err != nil {
			return err
		}
		result.Offers.TopTags = append(result.Offers.TopTags, item)
	}

	const topFavoritesQuery = `
		SELECT o.id, o.author_id, o.name, COUNT(f.user_id) AS favorites_count
		FROM offers o
		JOIN favorite_offers f ON f.offer_id = o.id
		GROUP BY o.id, o.author_id, o.name
		ORDER BY favorites_count DESC, o.id ASC
		LIMIT 10
	`

	rows, err = r.db.Query(ctx, topFavoritesQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var item AdminTopFavoriteOfferStatResult
		if err = rows.Scan(&item.OfferID, &item.AuthorID, &item.Name, &item.FavoritesCount); err != nil {
			return err
		}
		result.Offers.TopByFavorites = append(result.Offers.TopByFavorites, item)
	}

	return nil
}

func (r *Repository) queryAdminPlatformDeals(ctx context.Context, result *AdminPlatformStatisticsResult) error {
	const query = `
		WITH participant_counts AS (
			SELECT d.id, d.status, COUNT(p.user_id) AS participants_count
			FROM deals d
			LEFT JOIN participants p ON p.deal_id = d.id
			GROUP BY d.id, d.status
		)
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'LookingForParticipants') AS looking_for_participants,
			COUNT(*) FILTER (WHERE status = 'Discussion') AS discussion,
			COUNT(*) FILTER (WHERE status = 'Confirmed') AS confirmed,
			COUNT(*) FILTER (WHERE status = 'Completed') AS completed,
			COUNT(*) FILTER (WHERE status = 'Failed') AS failed,
			COUNT(*) FILTER (WHERE status = 'Cancelled') AS cancelled,
			COALESCE(
				COUNT(*) FILTER (WHERE status = 'Completed')::float8 /
				NULLIF(COUNT(*) FILTER (WHERE status IN ('Completed', 'Failed', 'Cancelled')), 0),
				0
			) AS successful_conversion_rate,
			COALESCE(AVG(participants_count)::float8, 0) AS average_participants,
			COALESCE(COUNT(*) FILTER (WHERE participants_count > 2)::float8 / NULLIF(COUNT(*), 0), 0) AS multi_party_share
		FROM participant_counts
	`

	return r.db.QueryRow(ctx, query).Scan(
		&result.Deals.Total,
		&result.Deals.ByStatus.LookingForParticipants,
		&result.Deals.ByStatus.Discussion,
		&result.Deals.ByStatus.Confirmed,
		&result.Deals.ByStatus.Completed,
		&result.Deals.ByStatus.Failed,
		&result.Deals.ByStatus.Cancelled,
		&result.Deals.SuccessfulConversionRate,
		&result.Deals.AverageParticipants,
		&result.Deals.MultiPartyShare,
	)
}

func (r *Repository) queryAdminPlatformReports(ctx context.Context, result *AdminPlatformStatisticsResult) error {
	const summaryQuery = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'Pending') AS pending,
			COALESCE((SELECT COUNT(*) FROM offers WHERE is_hidden AND hidden_reason IS NOT NULL), 0) AS blocked_offers,
			COALESCE((SELECT COUNT(*) FROM deal_failures WHERE confirmed_by_admin IS NOT NULL), 0) AS admin_failure_resolutions
		FROM offer_reports
	`

	if err := r.db.QueryRow(ctx, summaryQuery).Scan(
		&result.Reports.Total,
		&result.Reports.Pending,
		&result.Reports.BlockedOffers,
		&result.Reports.AdminFailureResolutions,
	); err != nil {
		return err
	}

	const topUsersQuery = `
		SELECT offer_author_id, COUNT(*) AS reports_count
		FROM offer_reports
		GROUP BY offer_author_id
		ORDER BY reports_count DESC, offer_author_id ASC
		LIMIT 10
	`

	rows, err := r.db.Query(ctx, topUsersQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var item AdminTopReportedUserStatResult
		if err = rows.Scan(&item.UserID, &item.ReportsCount); err != nil {
			return err
		}
		result.Reports.TopUsersByReceivedReports = append(result.Reports.TopUsersByReceivedReports, item)
	}

	return nil
}

func (r *Repository) queryAdminPlatformReviews(ctx context.Context, result *AdminPlatformStatisticsResult) error {
	const query = `
		SELECT
			COUNT(*) AS total,
			AVG(rating)::float8 AS average_rating,
			COUNT(*) FILTER (WHERE rating = 1) AS one_star,
			COUNT(*) FILTER (WHERE rating = 2) AS two_stars,
			COUNT(*) FILTER (WHERE rating = 3) AS three_stars,
			COUNT(*) FILTER (WHERE rating = 4) AS four_stars,
			COUNT(*) FILTER (WHERE rating = 5) AS five_stars
		FROM deal_reviews
	`

	return r.db.QueryRow(ctx, query).Scan(
		&result.Reviews.Total,
		&result.Reviews.AverageRating,
		&result.Reviews.RatingDistribution.OneStar,
		&result.Reviews.RatingDistribution.TwoStars,
		&result.Reviews.RatingDistribution.ThreeStars,
		&result.Reviews.RatingDistribution.FourStars,
		&result.Reviews.RatingDistribution.FiveStars,
	)
}

func (r *Repository) queryAdminUserDeals(ctx context.Context, userID uuid.UUID, result *AdminUserStatisticsResult) error {
	const query = `
		SELECT
			COUNT(*) FILTER (WHERE d.status = 'Completed') AS completed,
			COUNT(*) FILTER (WHERE d.status IN ('LookingForParticipants', 'Discussion', 'Confirmed')) AS active,
			COUNT(*) FILTER (WHERE d.status = 'Failed') AS failed_total,
			COUNT(*) FILTER (WHERE d.status = 'Cancelled') AS cancelled,
			COUNT(*) FILTER (WHERE d.status = 'Failed' AND df.user_id = $1) AS failed_responsible,
			COUNT(*) FILTER (WHERE d.status = 'Failed' AND (df.user_id IS NULL OR df.user_id <> $1)) AS failed_affected
		FROM participants p
		JOIN deals d ON d.id = p.deal_id
		LEFT JOIN deal_failures df ON df.deal_id = d.id
		WHERE p.user_id = $1
	`

	return r.db.QueryRow(ctx, query, userID).Scan(
		&result.Deals.Completed,
		&result.Deals.Active,
		&result.Deals.Failed.Total,
		&result.Deals.Cancelled,
		&result.Deals.Failed.Responsible,
		&result.Deals.Failed.Affected,
	)
}

func (r *Repository) queryAdminUserOffers(ctx context.Context, userID uuid.UUID, result *AdminUserStatisticsResult) error {
	const query = `
		SELECT
			COUNT(*) AS published,
			COALESCE(SUM(views), 0) AS total_views
		FROM offers
		WHERE author_id = $1
	`

	return r.db.QueryRow(ctx, query, userID).Scan(
		&result.Offers.Published,
		&result.Offers.TotalViews,
	)
}

func (r *Repository) queryAdminUserReviews(ctx context.Context, userID uuid.UUID, result *AdminUserStatisticsResult) error {
	const query = `
		SELECT
			COUNT(*) FILTER (WHERE provider_id = $1) AS received,
			AVG(rating) FILTER (WHERE provider_id = $1)::float8 AS average_received_rating,
			COUNT(*) FILTER (WHERE author_id = $1) AS written
		FROM deal_reviews
		WHERE provider_id = $1 OR author_id = $1
	`

	return r.db.QueryRow(ctx, query, userID).Scan(
		&result.Reviews.Received,
		&result.Reviews.AverageReceivedRating,
		&result.Reviews.Written,
	)
}

func (r *Repository) queryAdminUserReports(ctx context.Context, userID uuid.UUID, result *AdminUserStatisticsResult) error {
	const receivedQuery = `
		SELECT
			COUNT(*) FILTER (WHERE status = 'Accepted') AS accepted,
			COUNT(*) FILTER (WHERE status = 'Rejected') AS rejected
		FROM offer_reports
		WHERE offer_author_id = $1
	`

	if err := r.db.QueryRow(ctx, receivedQuery, userID).Scan(
		&result.Reports.Received.Accepted,
		&result.Reports.Received.Rejected,
	); err != nil {
		return err
	}

	const filedQuery = `
		SELECT COUNT(DISTINCT offer_report_id)
		FROM offer_reports_messages
		WHERE author_id = $1
	`

	return r.db.QueryRow(ctx, filedQuery, userID).Scan(&result.Reports.Filed)
}
