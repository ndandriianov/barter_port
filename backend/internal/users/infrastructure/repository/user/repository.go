package user

import (
	"barter-port/internal/libs/db"
	"barter-port/internal/libs/repox"
	"barter-port/internal/users/model"
	"errors"

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
//   - Returns only database errors, no domain-specific errors are expected.
func (r *Repository) AddUser(ctx context.Context, db db.DB, userID uuid.UUID) error {
	query := `
		INSERT INTO users_db.public.users (id)
		VALUES ($1)
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil && repox.IsUniqueViolation(err) {
		return model.ErrUserAlreadyExists
	}
	return err
}

// GetUserById returns user by id.
//
// Errors:
//   - model.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) GetUserById(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, name, bio
		FROM users_db.public.users
		WHERE id = $1
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
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
//   - model.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	query := `
		UPDATE users_db.public.users
		SET name = $2
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrUserNotFound
	}

	return nil
}

// UpdateBio updates the bio of a user.
//
// Errors:
//   - model.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error {
	query := `
		UPDATE users_db.public.users
		SET bio = $2
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, *bio)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrUserNotFound
	}

	return nil
}
