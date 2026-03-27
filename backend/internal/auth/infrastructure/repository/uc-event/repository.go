package uc_event

import (
	"barter-port/internal/auth/application"
	"barter-port/internal/auth/domain"
	"barter-port/pkg/db"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// Add inserts a new user creation event into the user_creation_events table.
// Produces only internal db errors.
func (r *Repository) Add(ctx context.Context, exec db.DB, event domain.UserCreationEvent) error {
	query := `INSERT INTO user_creation_events (user_id, created_at, status) VALUES ($1, $2, $3)`
	_, err := exec.Exec(ctx, query, event.UserID, event.CreatedAt, event.Status)
	return err
}

// GetByUserID retrieves a user creation event by user ID.
//
// Errors:
//   - application.ErrUserNotFound
func (r *Repository) GetByUserID(ctx context.Context, exec db.DB, userID uuid.UUID) (*domain.UserCreationEvent, error) {
	query := `SELECT user_id, created_at, status FROM user_creation_events WHERE user_id = $1`

	rows, err := exec.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	event, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.UserCreationEvent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrUserNotFound
		}
		return nil, err
	}

	return &event, nil
}

// SetStatus updates the status of a user creation event for a given user ID.
//
// Errors:
//   - application.ErrUserNotFound
func (r *Repository) SetStatus(ctx context.Context, exec db.DB, userID uuid.UUID, status string) error {
	query := `UPDATE user_creation_events SET status = $1 WHERE user_id = $2`

	tags, err := exec.Exec(ctx, query, status, userID)
	if err != nil {
		return err
	}
	if tags.RowsAffected() == 0 {
		return application.ErrUserNotFound
	}

	return nil
}
