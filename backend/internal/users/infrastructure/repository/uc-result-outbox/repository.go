package uc_result_outbox

import (
	usersauth "barter-port/contracts/kafka/messages/users-auth"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrUCResulMessageNotFound = errors.New("user creation result message not found")

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

// WriteUCResultMessage adds a new user creation result event to the outbox table.
// Produces only internal db errors.
func (r *Repository) WriteUCResultMessage(ctx context.Context, exec pgx.Tx, message usersauth.UCResultMessage) error {
	query := `
		INSERT INTO user_creation_result_outbox (id, event_id, user_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := exec.Exec(ctx, query, message.ID, message.EventID, message.UserID, message.Status, message.CreatedAt)
	return err
}

// ReadUCResultMessagesForUpdate retrieves a batch of user creation result events from the outbox table for processing.
// It locks the selected rows to prevent concurrent processing by other workers.
// Produces only internal db errors.
func (r *Repository) ReadUCResultMessagesForUpdate(ctx context.Context, exec pgx.Tx, limit int) ([]usersauth.UCResultMessage, error) {
	query := `
		SELECT id, event_id, user_id, status, created_at FROM user_creation_result_outbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`

	rows, err := exec.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (usersauth.UCResultMessage, error) {
		return pgx.RowToStructByName[usersauth.UCResultMessage](row)
	})
}

// DeleteUCResultMessage removes a user creation result event from the outbox table by its ID.
//
// Errors:
//   - ErrUCResulMessageNotFound: Occurs if no event is found with the given ID.
func (r *Repository) DeleteUCResultMessage(ctx context.Context, exec pgx.Tx, id uuid.UUID) error {
	query := `
		DELETE FROM user_creation_result_outbox
       	WHERE id = $1`

	tag, err := exec.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUCResulMessageNotFound
	}

	return nil
}
