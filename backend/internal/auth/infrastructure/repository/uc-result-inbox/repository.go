package uc_result_inbox

import (
	usersauth "barter-port/contracts/kafka/messages/users-auth"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrUCResultEventAlreadyExists = errors.New("event already exists")

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// WriteUCResultMessage writes a new user creation result event to the inbox table.
//
// Domain Errors:
//   - ErrUCResultEventAlreadyExists: Occurs if an event with the same ID already exists in the inbox.
func (r *Repository) WriteUCResultMessage(ctx context.Context, exec db.DB, message usersauth.UCResultMessage) error {
	query := `
		INSERT INTO user_creation_result_inbox (id, user_id, status, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err := exec.Exec(ctx, query, message.ID, message.UserID, message.Status, message.CreatedAt)
	if err != nil {
		if repox.IsUniqueViolation(err) {
			return ErrUCResultEventAlreadyExists
		}
		return err
	}

	return nil
}

// ReadUCResultMessagesForUpdate retrieves a batch of user creation result events from the inbox table for processing.
// It locks the selected rows to prevent concurrent processing by other workers.
func (r *Repository) ReadUCResultMessagesForUpdate(ctx context.Context, exec db.DB, limit int) ([]usersauth.UCResultMessage, error) {
	query := `
		SELECT id, user_id, status, created_at FROM user_creation_result_inbox
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

// DeleteUCResultMessage removes a user creation result event from the inbox table by its ID.
func (r *Repository) DeleteUCResultMessage(ctx context.Context, exec db.DB, id uuid.UUID) error {
	query := `
		DELETE FROM user_creation_result_inbox
		WHERE id = $1`

	_, err := exec.Exec(ctx, query, id)
	return err
}
