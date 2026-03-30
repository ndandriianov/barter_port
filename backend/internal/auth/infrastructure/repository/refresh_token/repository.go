package refresh_token

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

// Save adds a new refresh token to the repository.
//
// Errors:
//   - domain.ErrRefreshAlreadyExists: Occurs if a refresh token with the same JTI already exists in the repository.
func (r *Repository) Save(ctx context.Context, exec db.DB, token domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (jti, user_id, expires_at, revoked)
		VALUES ($1, $2, $3, $4)
	`

	_, err := exec.Exec(ctx, query, token.JTI, token.UserID, token.ExpiresAt, token.Revoked)
	if err != nil {
		if repox.IsUniqueViolation(err) {
			return domain.ErrRefreshAlreadyExists
		}
		return err
	}

	return nil
}

// GetByJTI retrieves a refresh token by its JTI.
//
// Errors:
//   - domain.ErrRefreshNotFound: Occurs if no refresh token is found with the given JTI.
func (r *Repository) GetByJTI(ctx context.Context, exec db.DB, jti string) (domain.RefreshToken, error) {
	query := `
		SELECT jti, user_id, expires_at, revoked
		FROM refresh_tokens
		WHERE jti = $1
	`

	rows, err := exec.Query(ctx, query, jti)
	if err != nil {
		return domain.RefreshToken{}, err
	}
	defer rows.Close()

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.RefreshToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RefreshToken{}, domain.ErrRefreshNotFound
		}
		return domain.RefreshToken{}, err
	}

	return token, nil
}

// Revoke marks a refresh token as revoked by its JTI.
//
// Errors:
//   - domain.ErrRefreshNotFound: Occurs if no refresh token is found with the given JTI.
func (r *Repository) Revoke(ctx context.Context, exec db.DB, jti string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE jti = $1
	`

	cmdTag, err := exec.Exec(ctx, query, jti)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrRefreshNotFound
	}

	return nil
}

// DeleteAllForUser removes all refresh tokens associated with a specific user.
//
// Errors: returns only internal errors, as the operation is idempotent and does not fail if no tokens are found for the user.
func (r *Repository) DeleteAllForUser(ctx context.Context, exec db.DB, userID uuid.UUID) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`

	_, err := exec.Exec(ctx, query, userID)
	return err
}
