package email_token

import (
	"barter-port/internal/auth/model"
	"barter-port/internal/auth/repository"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

var (
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenAlreadyExists = errors.New("token already exists")
)

type Repository struct {
	mu     sync.RWMutex
	byHash map[string]model.EmailVerificationToken

	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		byHash: make(map[string]model.EmailVerificationToken),
		db:     db,
	}
}

// Save stores a new email verification token in the repository.
// Errors:
//   - errors.ErrTokenAlreadyExists: Occurs if a token with the same hash already exists in the repository.
func (r *Repository) Save(ctx context.Context, t model.EmailVerificationToken) error {
	query := `
		INSERT INTO email_tokens
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query, t.TokenHash, t.UserID, t.ExpiresAt, t.Used, t.CreatedAt)
	if err != nil {
		if repository.IsUniqueViolation(err) {
			return ErrTokenAlreadyExists
		}
		return err
	}

	return nil
}

// GetByHash retrieves an email verification token by its hash.
// Errors:
//   - errors.ErrTokenNotFound: Occurs if no token is found with the given hash.
func (r *Repository) GetByHash(ctx context.Context, tokenHash string) (model.EmailVerificationToken, error) {
	query := `
		SELECT token_hash, user_id, expires_at, used, created_at
		FROM email_tokens
		WHERE token_hash = $1
	`

	rows, err := r.db.Query(ctx, query, tokenHash)
	if err != nil {
		return model.EmailVerificationToken{}, err
	}
	defer rows.Close()

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.EmailVerificationToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EmailVerificationToken{}, ErrTokenNotFound
		}
		return model.EmailVerificationToken{}, err
	}

	return token, nil
}

// MarkUsed marks an email verification token as used.
// Errors:
//   - errors.ErrTokenNotFound: Occurs if no token is found with the given hash.
func (r *Repository) MarkUsed(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE email_tokens
		SET used = true
		WHERE token_hash = $1
	`

	cmdTag, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrTokenNotFound
	}

	return nil
}

// DeleteAllForUser removes all tokens associated with a specific user.
// Errors: only internal db errors.
func (r *Repository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM email_tokens
		WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}
