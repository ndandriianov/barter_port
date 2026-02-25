package email_token

import (
	"barter-port/internal/infrastructure/repository"
	"errors"
	"sync"

	"barter-port/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

var (
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenAlreadyExists = errors.New("token already exists")
)

type InMemoryTokenRepo struct {
	mu     sync.RWMutex
	byHash map[string]model.EmailVerificationToken

	db *pgxpool.Pool
}

func NewInMemoryTokenRepo(db *pgxpool.Pool) *InMemoryTokenRepo {
	return &InMemoryTokenRepo{
		byHash: make(map[string]model.EmailVerificationToken),
		db:     db,
	}
}

// Save stores a new email verification token in the repository.
// Errors:
//   - errors.ErrTokenAlreadyExists: Occurs if a token with the same hash already exists in the repository.
func (r *InMemoryTokenRepo) Save(ctx context.Context, t model.EmailVerificationToken) error {
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
func (r *InMemoryTokenRepo) GetByHash(tokenHash string) (model.EmailVerificationToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.byHash[tokenHash]
	if !ok {
		return model.EmailVerificationToken{}, ErrTokenNotFound
	}
	return t, nil
}

// MarkUsed marks an email verification token as used.
// Errors:
//   - errors.ErrTokenNotFound: Occurs if no token is found with the given hash.
func (r *InMemoryTokenRepo) MarkUsed(tokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.byHash[tokenHash]
	if !ok {
		return ErrTokenNotFound
	}

	t.Used = true
	r.byHash[tokenHash] = t
	return nil
}

// DeleteAllForUser removes all tokens associated with a specific user.
// Errors: None.
func (r *InMemoryTokenRepo) DeleteAllForUser(userID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, v := range r.byHash {
		if v.UserID == userID {
			delete(r.byHash, k)
		}
	}
}
