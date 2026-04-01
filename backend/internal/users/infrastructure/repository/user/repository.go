package user

import (
	"barter-port/internal/users/domain"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// AddUser adds a new user to the repository.
//
// Errors:
//   - domain.ErrUserAlreadyExists: Occurs if a user with the same ID already exists in the repository.
func (r *Repository) AddUser(ctx context.Context, db db.DB, userID uuid.UUID) error {
	query := `
		INSERT INTO users (id)
		VALUES ($1)
	`

	_, err := db.Exec(ctx, query, userID)
	if err != nil && repox.IsUniqueViolation(err) {
		return domain.ErrUserAlreadyExists
	}
	return err
}

// GetUserById returns user by id.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) GetUserById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, name, bio
		FROM users
		WHERE id = $1
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// DeleteUser should be used only with transaction, with deleting auth-service user and
// transferring userId to deleted users table
func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

// UpdateName updates the name of a user.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	query := `
		UPDATE users
		SET name = $2
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// UpdateBio updates the bio of a user.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error {
	query := `
		UPDATE users
		SET bio = $2
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, bio)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// GetNamesForUserIDs returns a map of user IDs to their corresponding names.
//
// No domain Errors
func (r *Repository) GetNamesForUserIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	query := `
		SELECT id, name
		FROM users
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("sql: %w", err)
	}
	defer rows.Close()

	names := make(map[uuid.UUID]string)
	for rows.Next() {
		var id uuid.UUID
		var name string

		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		names[id] = name
	}

	return names, nil
}
