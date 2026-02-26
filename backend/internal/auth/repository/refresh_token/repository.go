package refresh_token

import (
	"barter-port/internal/auth/model"
	"barter-port/internal/auth/repository"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

var (
	ErrRefreshNotFound      = errors.New("refresh token not found")
	ErrRefreshAlreadyExists = errors.New("refresh token with this JTI already exists")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Save(ctx context.Context, token model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (jti, user_id, expires_at, revoked)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query, token.JTI, token.UserID, token.ExpiresAt, token.Revoked)
	if err != nil {
		if repository.IsUniqueViolation(err) {
			return ErrRefreshAlreadyExists
		}
		return err
	}

	return nil
}

func (r *Repository) GetByJTI(ctx context.Context, jti string) (model.RefreshToken, error) {
	query := `
		SELECT jti, user_id, expires_at, revoked
		FROM refresh_tokens
		WHERE jti = $1
	`

	rows, err := r.db.Query(ctx, query, jti)
	if err != nil {
		return model.RefreshToken{}, err
	}
	defer rows.Close()

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.RefreshToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshToken{}, ErrRefreshNotFound
		}
		return model.RefreshToken{}, err
	}

	return token, nil
}

func (r *Repository) Revoke(ctx context.Context, jti string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE jti = $1
	`

	cmdTag, err := r.db.Exec(ctx, query, jti)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrRefreshNotFound
	}

	return nil
}

func (r *Repository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}
