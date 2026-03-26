package inbox

import (
	authusers "barter-port/internal/contracts/kafka/messages/auth-users"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
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

// WriteUserCreationMessage writes a new user creation event to the inbox table.
//
// Domain Errors:
//   - ErrUCEventAlreadyExists: Occurs if an event with the same ID already exists in the inbox.
func (r *Repository) WriteUserCreationMessage(ctx context.Context, exec db.DB, message authusers.UserCreationMessage) error {
	query := `
		INSERT INTO user_creation_inbox (id, event_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err := exec.Exec(ctx, query, message.ID, message.EventID, message.UserID, message.CreatedAt)
	if err != nil {
		if repox.IsUniqueViolation(err) {
			return ErrUCEventAlreadyExists
		}
		return err
	}

	return nil
}

// ReadUserCreationMessagesForUpdate retrieves a batch of user creation events from the inbox table for processing.
// It locks the selected rows to prevent concurrent processing by other workers.
//
// No Domain Errors
func (r *Repository) ReadUserCreationMessagesForUpdate(ctx context.Context, exec db.DB, limit int) ([]authusers.UserCreationMessage, error) {
	query := `
		SELECT id, event_id, user_id, created_at FROM user_creation_inbox
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

// DeleteUserCreationMessage removes a user creation event from the inbox table by its ID.
//
// No Domain Errors
func (r *Repository) DeleteUserCreationMessage(ctx context.Context, exec db.DB, id uuid.UUID) error {
	query := `
		DELETE FROM user_creation_inbox
		WHERE id = $1`

	_, err := exec.Exec(ctx, query, id)
	return err
}
