package offer_reports

import (
	"barter-port/internal/deals/domain"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"

	"github.com/google/uuid"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateReport inserts a new offer report.
func (r *Repository) CreateReport(ctx context.Context, exec db.DB, report domain.OfferReport) error {
	const query = `
		INSERT INTO offer_reports (id, offer_id, offer_author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := exec.Exec(ctx, query, report.ID, report.OfferID, report.OfferAuthorID, string(report.Status), report.CreatedAt)
	return err
}

// AddReporterMessage adds a message to an existing report.
//
// Domain errors:
//   - domain.ErrReporterAlreadyAttached: if the reporter already has a message in this report.
func (r *Repository) AddReporterMessage(ctx context.Context, exec db.DB, msg domain.OfferReportMessage) error {
	const query = `
		INSERT INTO offer_reports_messages (offer_report_id, author_id, message)
		VALUES ($1, $2, $3)`

	_, err := exec.Exec(ctx, query, msg.OfferReportID, msg.AuthorID, msg.Message)
	if err != nil {
		if repox.IsUniqueViolation(err) {
			return domain.ErrReporterAlreadyAttached
		}
		return err
	}
	return nil
}

// GetReportByID retrieves a report by ID.
//
// Domain errors:
//   - domain.ErrReportNotFound
func (r *Repository) GetReportByID(ctx context.Context, exec db.DB, reportID uuid.UUID) (*domain.OfferReport, error) {
	const query = `
		SELECT id, offer_id, offer_author_id, status, created_at,
		       reviewed_at, reviewed_by, resolution_comment, applied_penalty_delta
		FROM offer_reports
		WHERE id = $1`

	var report domain.OfferReport
	err := exec.QueryRow(ctx, query, reportID).Scan(
		&report.ID,
		&report.OfferID,
		&report.OfferAuthorID,
		&report.Status,
		&report.CreatedAt,
		&report.ReviewedAt,
		&report.ReviewedBy,
		&report.ResolutionComment,
		&report.AppliedPenaltyDelta,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReportNotFound
		}
		return nil, err
	}
	return &report, nil
}

// GetPendingReportForOffer returns the current pending report for an offer, or nil if none.
func (r *Repository) GetPendingReportForOffer(ctx context.Context, exec db.DB, offerID uuid.UUID) (*domain.OfferReport, error) {
	const query = `
		SELECT id, offer_id, offer_author_id, status, created_at,
		       reviewed_at, reviewed_by, resolution_comment, applied_penalty_delta
		FROM offer_reports
		WHERE offer_id = $1 AND status = 'Pending'`

	var report domain.OfferReport
	err := exec.QueryRow(ctx, query, offerID).Scan(
		&report.ID,
		&report.OfferID,
		&report.OfferAuthorID,
		&report.Status,
		&report.CreatedAt,
		&report.ReviewedAt,
		&report.ReviewedBy,
		&report.ResolutionComment,
		&report.AppliedPenaltyDelta,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &report, nil
}

// ListReports returns all reports, optionally filtered by status.
func (r *Repository) ListReports(ctx context.Context, exec db.DB, status *domain.OfferReportStatus) ([]domain.OfferReport, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `
			SELECT id, offer_id, offer_author_id, status, created_at,
			       reviewed_at, reviewed_by, resolution_comment, applied_penalty_delta
			FROM offer_reports
			WHERE status = $1
			ORDER BY created_at DESC`
		args = append(args, string(*status))
	} else {
		query = `
			SELECT id, offer_id, offer_author_id, status, created_at,
			       reviewed_at, reviewed_by, resolution_comment, applied_penalty_delta
			FROM offer_reports
			ORDER BY created_at DESC`
	}

	rows, err := exec.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("sql list reports: %w", err)
	}
	defer rows.Close()

	var reports []domain.OfferReport
	for rows.Next() {
		var report domain.OfferReport
		if err = rows.Scan(
			&report.ID,
			&report.OfferID,
			&report.OfferAuthorID,
			&report.Status,
			&report.CreatedAt,
			&report.ReviewedAt,
			&report.ReviewedBy,
			&report.ResolutionComment,
			&report.AppliedPenaltyDelta,
		); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

// GetReportMessages returns all messages for a given report.
func (r *Repository) GetReportMessages(ctx context.Context, exec db.DB, reportID uuid.UUID) ([]domain.OfferReportMessage, error) {
	const query = `
		SELECT offer_report_id, author_id, message
		FROM offer_reports_messages
		WHERE offer_report_id = $1`

	rows, err := exec.Query(ctx, query, reportID)
	if err != nil {
		return nil, fmt.Errorf("sql get report messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.OfferReportMessage
	for rows.Next() {
		var msg domain.OfferReportMessage
		if err = rows.Scan(&msg.OfferReportID, &msg.AuthorID, &msg.Message); err != nil {
			return nil, fmt.Errorf("scan report message: %w", err)
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// GetReportMessagesForOfferReports returns messages for a slice of report IDs, grouped by report ID.
func (r *Repository) GetReportMessagesForOfferReports(ctx context.Context, exec db.DB, reportIDs []uuid.UUID) (map[uuid.UUID][]domain.OfferReportMessage, error) {
	if len(reportIDs) == 0 {
		return map[uuid.UUID][]domain.OfferReportMessage{}, nil
	}

	const query = `
		SELECT offer_report_id, author_id, message
		FROM offer_reports_messages
		WHERE offer_report_id = ANY($1)`

	rows, err := exec.Query(ctx, query, reportIDs)
	if err != nil {
		return nil, fmt.Errorf("sql get report messages for reports: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]domain.OfferReportMessage)
	for rows.Next() {
		var msg domain.OfferReportMessage
		if err = rows.Scan(&msg.OfferReportID, &msg.AuthorID, &msg.Message); err != nil {
			return nil, fmt.Errorf("scan report message: %w", err)
		}
		result[msg.OfferReportID] = append(result[msg.OfferReportID], msg)
	}
	return result, rows.Err()
}

// GetReportsForOffer returns all reports for an offer (all statuses).
func (r *Repository) GetReportsForOffer(ctx context.Context, exec db.DB, offerID uuid.UUID) ([]domain.OfferReport, error) {
	const query = `
		SELECT id, offer_id, offer_author_id, status, created_at,
		       reviewed_at, reviewed_by, resolution_comment, applied_penalty_delta
		FROM offer_reports
		WHERE offer_id = $1
		ORDER BY created_at DESC`

	rows, err := exec.Query(ctx, query, offerID)
	if err != nil {
		return nil, fmt.Errorf("sql get reports for offer: %w", err)
	}
	defer rows.Close()

	var reports []domain.OfferReport
	for rows.Next() {
		var report domain.OfferReport
		if err = rows.Scan(
			&report.ID,
			&report.OfferID,
			&report.OfferAuthorID,
			&report.Status,
			&report.CreatedAt,
			&report.ReviewedAt,
			&report.ReviewedBy,
			&report.ResolutionComment,
			&report.AppliedPenaltyDelta,
		); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

// UpdateReportResolution sets the review fields on a report.
func (r *Repository) UpdateReportResolution(
	ctx context.Context,
	exec db.DB,
	reportID uuid.UUID,
	reviewedBy uuid.UUID,
	newStatus domain.OfferReportStatus,
	comment *string,
	penaltyDelta *int,
	now time.Time,
) error {
	const query = `
		UPDATE offer_reports
		SET status             = $2,
		    reviewed_at        = $3,
		    reviewed_by        = $4,
		    resolution_comment = $5,
		    applied_penalty_delta = $6
		WHERE id = $1`

	tag, err := exec.Exec(ctx, query, reportID, string(newStatus), now, reviewedBy, comment, penaltyDelta)
	if err != nil {
		return fmt.Errorf("sql update report resolution: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrReportNotFound
	}
	return nil
}

// BlockOfferModification sets modification_blocked=true on an offer.
func (r *Repository) BlockOfferModification(ctx context.Context, exec db.DB, offerID uuid.UUID, now time.Time) error {
	const query = `
		UPDATE offers
		SET modification_blocked = TRUE, modification_blocked_at = $2
		WHERE id = $1`

	_, err := exec.Exec(ctx, query, offerID, now)
	return err
}

// UnblockOfferModification clears modification_blocked on an offer.
func (r *Repository) UnblockOfferModification(ctx context.Context, exec db.DB, offerID uuid.UUID) error {
	const query = `
		UPDATE offers
		SET modification_blocked = FALSE, modification_blocked_at = NULL
		WHERE id = $1`

	_, err := exec.Exec(ctx, query, offerID)
	return err
}

// HideOffer sets is_hidden=true on an offer.
func (r *Repository) HideOffer(ctx context.Context, exec db.DB, offerID uuid.UUID, reason string, now time.Time) error {
	const query = `
		UPDATE offers
		SET is_hidden = TRUE, hidden_at = $2, hidden_reason = $3
		WHERE id = $1`

	_, err := exec.Exec(ctx, query, offerID, now, reason)
	return err
}
