package inbox

import (
	authusers "barter-port/internal/contracts/kafka/auth-users"
	"barter-port/internal/libs/db"
	"barter-port/internal/libs/repox"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

var ErrUCEventAlreadyExists = errors.New("event already exists")

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// WriteUserCreationEvent writes a new user creation event to the inbox table.
//
// Domain Errors:
//   - ErrUCEventAlreadyExists: Occurs if an event with the same ID already exists in the inbox.
func (r *Repository) WriteUserCreationEvent(ctx context.Context, exec db.DB, event authusers.UserCreationEvent) error {
	query := `
		INSERT INTO user_creation_inbox (id, user_id, created_at)
		VALUES ($1, $2, $3)`

	_, err := exec.Exec(ctx, query, event.ID, event.UserID, event.CreatedAt)
	if err != nil {
		if repox.IsUniqueViolation(err) {
			return ErrUCEventAlreadyExists
		}
		return err
	}

	return nil
}

// ReadUserCreationEventsForUpdate retrieves a batch of user creation events from the inbox table for processing.
// It locks the selected rows to prevent concurrent processing by other workers.
//
// No Domain Errors
func (r *Repository) ReadUserCreationEventsForUpdate(ctx context.Context, exec db.DB, limit int) ([]authusers.UserCreationEvent, error) {
	query := `
		SELECT id, user_id, created_at FROM user_creation_inbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`

	rows, err := exec.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (authusers.UserCreationEvent, error) {
		return pgx.RowToStructByName[authusers.UserCreationEvent](row)
	})
}

// DeleteUserCreationEvent removes a user creation event from the inbox table by its ID.
//
// No Domain Errors
func (r *Repository) DeleteUserCreationEvent(ctx context.Context, exec db.DB, id uuid.UUID) error {
	query := `
		DELETE FROM user_creation_inbox
		WHERE id = $1`

	_, err := exec.Exec(ctx, query, id)
	return err
}
