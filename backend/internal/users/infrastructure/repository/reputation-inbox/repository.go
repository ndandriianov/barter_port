package reputation_inbox

import (
	dealsusers "barter-port/contracts/kafka/messages/deals-users"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

var ErrReputationEventAlreadyExists = errors.New("reputation event already exists")

// InboxMessage mirrors the user_reputation_inbox table.
type InboxMessage struct {
	ID         uuid.UUID `db:"id"`
	SourceType string    `db:"source_type"`
	SourceID   uuid.UUID `db:"source_id"`
	UserID     uuid.UUID `db:"user_id"`
	Delta      int       `db:"delta"`
	CreatedAt  time.Time `db:"created_at"`
	Comment    *string   `db:"comment"`
}

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// WriteReputationInboxMessage writes a penalty event to the inbox table (idempotent).
//
// Domain errors:
//   - ErrReputationEventAlreadyExists
func (r *Repository) WriteReputationInboxMessage(ctx context.Context, exec db.DB, msg dealsusers.ReputationMessage) error {
	const query = `
		INSERT INTO user_reputation_inbox (id, source_type, source_id, user_id, delta, created_at, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (source_type, source_id) DO NOTHING`

	tag, err := exec.Exec(ctx, query, uuid.New(), msg.SourceType, msg.SourceID, msg.UserID, msg.Delta, msg.CreatedAt, msg.Comment)
	if err != nil {
		if repox.IsUniqueViolation(err) {
			return ErrReputationEventAlreadyExists
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrReputationEventAlreadyExists
	}
	return nil
}

// ReadMessagesForUpdate retrieves a batch of inbox messages for processing.
func (r *Repository) ReadMessagesForUpdate(ctx context.Context, exec db.DB, limit int) ([]InboxMessage, error) {
	const query = `
		SELECT i.id, i.source_type, i.source_id, i.user_id, i.delta, i.created_at, i.comment
		FROM user_reputation_inbox i
		WHERE EXISTS (
			SELECT 1
			FROM users u
			WHERE u.id = i.user_id
		)
		ORDER BY i.created_at, i.id
		LIMIT $1
		FOR UPDATE SKIP LOCKED`

	rows, err := exec.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("sql read reputation inbox: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (InboxMessage, error) {
		return pgx.RowToStructByName[InboxMessage](row)
	})
}

// DeleteMessage removes a message from the inbox.
func (r *Repository) DeleteMessage(ctx context.Context, exec db.DB, id uuid.UUID) error {
	_, err := exec.Exec(ctx, `DELETE FROM user_reputation_inbox WHERE id = $1`, id)
	return err
}
