package user

import (
	"barter-port/internal/auth/domain"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
	"errors"

	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyInUse = errors.New("email already in use")
)

type Repository struct {
}

func NewRepository() *Repository {
	return &Repository{}
}

// Create adds a new user to the repository.
// Errors:
//   - errors.ErrEmailAlreadyInUse - email already exists
func (r *Repository) Create(ctx context.Context, exec db.DB, u domain.User) error {
	query := `
		INSERT INTO users
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := exec.Exec(ctx, query, u.ID, u.Email, u.PasswordHash, u.EmailVerified, u.CreatedAt)
	if err != nil {
		if repox.IsUniqueViolation(err) {
			return ErrEmailAlreadyInUse
		}
		return err
	}

	return nil
}

// GetByEmail retrieves a user by their email address.
// Errors:
//   - errors.ErrUserNotFound: Occurs if no user is found with the given email address.
func (r *Repository) GetByEmail(ctx context.Context, exec db.DB, email string) (domain.User, error) {
	query := `
		SELECT id, email, password_hash, email_verified, created_at
		FROM users
		WHERE email = $1
	`

	rows, err := exec.Query(ctx, query, email)
	if err != nil {
		return domain.User{}, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, err
	}

	return user, nil
}

// GetByID retrieves a user by their unique ID.
// Errors:
//   - errors.ErrUserNotFound: Occurs if no user is found with the given ID.
func (r *Repository) GetByID(ctx context.Context, exec db.DB, id uuid.UUID) (domain.User, error) {
	query := `
		SELECT id, email, password_hash, email_verified, created_at
		FROM users
		WHERE id = $1
	`

	rows, err := exec.Query(ctx, query, id)
	if err != nil {
		return domain.User{}, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, err
	}

	return user, nil
}

// VerifyEmailIfNotVerified marks a user's email as verified. And returns true if the email was updated,
// false if it was already verified.
//
// Errors:
//   - errors.ErrUserNotFound: Occurs if no user is found with the given userID.
func (r *Repository) VerifyEmailIfNotVerified(ctx context.Context, exec db.DB, userID uuid.UUID) (changed bool, err error) {
	query := `
		WITH updated AS (
			UPDATE users
			SET email_verified = true
			WHERE id = $1 AND email_verified = false
			RETURNING id
		),
		existing AS (
			SELECT id
			FROM users
			WHERE id = $1
		)
		SELECT exists(SELECT 1 FROM updated)  AS changed,
			   exists(SELECT 1 FROM existing) AS exists
	`

	var exists bool

	if err = exec.QueryRow(ctx, query, userID).Scan(&changed, &exists); err != nil {
		return false, err
	}

	if !exists {
		return false, ErrUserNotFound
	}

	return changed, nil
}
