package outbox

import (
	"barter-port/internal/auth/domain"
	"errors"

	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrUserCreationEventNotFound = errors.New("user creation event not found")

type Repository struct{}

// WriteUserCreationEvent adds a new user creation event to the outbox table.
// Produces only internal db errors.
func (r *Repository) WriteUserCreationEvent(ctx context.Context, exec pgx.Tx, event domain.UserCreationEvent) error {
	query := `
		INSERT INTO user_creation_outbox (id, user_id, created_at)
		VALUES ($1, $2, $3)`

	_, err := exec.Exec(ctx, query, event.ID, event.UserID, event.CreatedAt)
	return err

}

// ReadUserCreationEventsForUpdate retrieves a batch of user creation events from the outbox table for processing.
// It locks the selected rows to prevent concurrent processing by other workers.
// Returns a slice of UserCreationEvent and any error encountered during the operation.
func (r *Repository) ReadUserCreationEventsForUpdate(ctx context.Context, exec pgx.Tx, limit int) ([]domain.UserCreationEvent, error) {
	query := `
		SELECT id, user_id, created_at FROM user_creation_outbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`

	rows, err := exec.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.UserCreationEvent, error) {
		return pgx.RowToStructByName[domain.UserCreationEvent](row)
	})
}

// DeleteUserCreationEvent removes a user creation event from the outbox table by its ID.
//
// Errors:
//   - ErrUserCreationEventNotFound: Occurs if no event is found with the given ID.
func (r *Repository) DeleteUserCreationEvent(ctx context.Context, exec pgx.Tx, id uuid.UUID) error {
	query := `
		DELETE FROM user_creation_outbox
		WHERE id = $1`

	tag, err := exec.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserCreationEventNotFound
	}

	return nil
}
