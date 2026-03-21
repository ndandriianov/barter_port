package outbox

import (
	authusers "barter-port/internal/contracts/kafka/messages/auth-users"
	"errors"

	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrUserCreationMessageNotFound = errors.New("user creation message not found")

type Repository struct{}

// WriteUserCreationMessage adds a new user creation event to the outbox table.
// Produces only internal db errors.
func (r *Repository) WriteUserCreationMessage(ctx context.Context, exec pgx.Tx, message authusers.UserCreationMessage) error {
	query := `
		INSERT INTO user_creation_outbox (id, event_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err := exec.Exec(ctx, query, message.ID, message.EventID, message.UserID, message.CreatedAt)
	return err
}

// ReadUserCreationMessagesForUpdate retrieves a batch of user creation events from the outbox table for processing.
// It locks the selected rows to prevent concurrent processing by other workers.
// Returns a slice of UserCreationEvent and any error encountered during the operation.
func (r *Repository) ReadUserCreationMessagesForUpdate(
	ctx context.Context,
	exec pgx.Tx,
	limit int,
) ([]authusers.UserCreationMessage, error) {
	query := `
		SELECT id, event_id, user_id, created_at FROM user_creation_outbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`

	rows, err := exec.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (authusers.UserCreationMessage, error) {
		return pgx.RowToStructByName[authusers.UserCreationMessage](row)
	})
}

// DeleteUserCreationMessage removes a user creation event from the outbox table by its ID.
//
// Errors:
//   - ErrUserCreationMessageNotFound: Occurs if no event is found with the given ID.
func (r *Repository) DeleteUserCreationMessage(ctx context.Context, exec pgx.Tx, id uuid.UUID) error {
	query := `
		DELETE FROM user_creation_outbox
		WHERE id = $1`

	tag, err := exec.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserCreationMessageNotFound
	}

	return nil
}
