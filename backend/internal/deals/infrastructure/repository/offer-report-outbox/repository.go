package offer_report_outbox

import (
	dealsusers "barter-port/contracts/kafka/messages/deals-users"
	"barter-port/pkg/db"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// WriteOutboxMessage writes a penalty event to the outbox table.
func (r *Repository) WriteOutboxMessage(ctx context.Context, exec db.DB, msg dealsusers.OfferReportPenaltyMessage) error {
	const query = `
		INSERT INTO offer_report_penalty_outbox (id, report_id, offer_id, user_id, delta, reviewed_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := exec.Exec(ctx, query,
		msg.ID, msg.ReportID, msg.OfferID, msg.UserID, msg.Delta, msg.ReviewedBy, msg.CreatedAt,
	)
	return err
}

// ReadOutboxMessagesForUpdate retrieves a batch of penalty events for publishing.
// Rows are locked with FOR UPDATE SKIP LOCKED.
func (r *Repository) ReadOutboxMessagesForUpdate(ctx context.Context, exec db.DB, limit int) ([]dealsusers.OfferReportPenaltyMessage, error) {
	const query = `
		SELECT id, report_id, offer_id, user_id, delta, reviewed_by, created_at
		FROM offer_report_penalty_outbox
		ORDER BY created_at, id
		LIMIT $1
		FOR UPDATE SKIP LOCKED`

	rows, err := exec.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("sql read outbox messages: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (dealsusers.OfferReportPenaltyMessage, error) {
		return pgx.RowToStructByName[dealsusers.OfferReportPenaltyMessage](row)
	})
}

// DeleteOutboxMessage removes a penalty event from the outbox.
func (r *Repository) DeleteOutboxMessage(ctx context.Context, exec db.DB, id uuid.UUID) error {
	_, err := exec.Exec(ctx, `DELETE FROM offer_report_penalty_outbox WHERE id = $1`, id)
	return err
}
