package reputation_events_outbox

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
func (r *Repository) WriteOutboxMessage(ctx context.Context, exec db.DB, msg dealsusers.ReputationMessage) error {
	const query = `
		INSERT INTO reputation_events_outbox (id, source_type, source_id, user_id, delta, created_at, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := exec.Exec(ctx, query,
		msg.ID, msg.SourceType, msg.SourceID, msg.UserID, msg.Delta, msg.CreatedAt, msg.Comment,
	)
	return err
}

// ReadOutboxMessagesForUpdate retrieves a batch of penalty events for publishing.
// Rows are locked with FOR UPDATE SKIP LOCKED.
func (r *Repository) ReadOutboxMessagesForUpdate(ctx context.Context, exec db.DB, limit int) ([]dealsusers.ReputationMessage, error) {
	const query = `
		SELECT id, source_type, source_id, user_id, delta, created_at, comment
		FROM reputation_events_outbox
		ORDER BY created_at, id
		LIMIT $1
		FOR UPDATE SKIP LOCKED`

	rows, err := exec.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("sql read outbox messages: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (dealsusers.ReputationMessage, error) {
		return pgx.RowToStructByName[dealsusers.ReputationMessage](row)
	})
}

// DeleteOutboxMessage removes a penalty event from the outbox.
func (r *Repository) DeleteOutboxMessage(ctx context.Context, exec db.DB, id uuid.UUID) error {
	_, err := exec.Exec(ctx, `DELETE FROM reputation_events_outbox WHERE id = $1`, id)
	return err
}
