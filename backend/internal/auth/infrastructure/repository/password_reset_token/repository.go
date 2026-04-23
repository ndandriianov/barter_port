package password_reset_token

import (
	"barter-port/internal/auth/domain"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

type Repository struct {
}

func NewRepository() *Repository {
	return &Repository{}
}

// Save stores a new password reset token in the repository.
// Errors:
//   - domain.ErrTokenAlreadyExists: Occurs if a token with the same hash already exists in the repository.
func (r *Repository) Save(ctx context.Context, exec db.DB, t domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := exec.Exec(ctx, query, t.TokenHash, t.UserID, t.ExpiresAt, t.Used, t.CreatedAt)
	if err != nil {
		if repox.IsUniqueViolation(err) {
			return domain.ErrTokenAlreadyExists
		}
		return err
	}

	return nil
}

// GetByHashForUpdate retrieves a password reset token by its hash.
// Errors:
//   - domain.ErrTokenNotFound: Occurs if no token is found with the given hash.
func (r *Repository) GetByHashForUpdate(ctx context.Context, exec db.DB, tokenHash string) (domain.PasswordResetToken, error) {
	query := `
		SELECT token_hash, user_id, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
		FOR UPDATE
	`

	rows, err := exec.Query(ctx, query, tokenHash)
	if err != nil {
		return domain.PasswordResetToken{}, err
	}
	defer rows.Close()

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.PasswordResetToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PasswordResetToken{}, domain.ErrTokenNotFound
		}
		return domain.PasswordResetToken{}, err
	}

	return token, nil
}

// MarkUsed marks a password reset token as used.
// Errors:
//   - domain.ErrTokenNotFound: Occurs if no token is found with the given hash.
func (r *Repository) MarkUsed(ctx context.Context, exec db.DB, tokenHash string) error {
	query := `
		UPDATE password_reset_tokens
		SET used = true
		WHERE token_hash = $1
	`

	cmdTag, err := exec.Exec(ctx, query, tokenHash)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrTokenNotFound
	}

	return nil
}

// DeleteAllForUser removes all password reset tokens associated with a specific user.
// Errors: only internal db errors.
func (r *Repository) DeleteAllForUser(ctx context.Context, exec db.DB, userID uuid.UUID) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE user_id = $1
	`

	_, err := exec.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}
