package user

import (
	"barter-port/internal/auth/model"
	"barter-port/internal/libs/repox"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyInUse = errors.New("email already in use")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create adds a new user to the repository.
// Errors:
//   - errors.ErrEmailAlreadyInUse - email already exists
func (r *Repository) Create(ctx context.Context, u model.User) error {
	query := `
		INSERT INTO users
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query, u.ID, u.Email, u.PasswordHash, u.EmailVerified, u.CreatedAt)
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
func (r *Repository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	query := `
		SELECT id, email, password_hash, email_verified, created_at
		FROM users
		WHERE email = $1
	`

	rows, err := r.db.Query(ctx, query, email)
	if err != nil {
		return model.User{}, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, err
	}

	return user, nil
}

// GetByID retrieves a user by their unique ID.
// Errors:
//   - errors.ErrUserNotFound: Occurs if no user is found with the given ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	query := `
		SELECT id, email, password_hash, email_verified, created_at
		FROM users
		WHERE id = $1
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return model.User{}, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, err
	}

	return user, nil
}

// VerifyEmail marks a user's email as verified.
// Errors:
//   - errors.ErrUserNotFound: Occurs if no user is found with the given userID.
func (r *Repository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET email_verified = true
		WHERE id = $1
	`

	cmdTag, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
