package token

import (
	"errors"
	"sync"

	"github.com/ndandriianov/barter_port/backend/internal/model"
)

var (
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenAlreadyExists = errors.New("token already exists")
)

type InMemoryTokenRepo struct {
	mu     sync.RWMutex
	byHash map[string]model.EmailVerificationToken
}

func NewInMemoryTokenRepo() *InMemoryTokenRepo {
	return &InMemoryTokenRepo{
		byHash: make(map[string]model.EmailVerificationToken),
	}
}

// Save stores a new email verification token in the repository.
// Errors:
//   - errors.ErrTokenAlreadyExists: Occurs if a token with the same hash already exists in the repository.
func (r *InMemoryTokenRepo) Save(t model.EmailVerificationToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byHash[t.TokenHash]; exists {
		return ErrTokenAlreadyExists
	}

	r.byHash[t.TokenHash] = t
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
func (r *InMemoryTokenRepo) DeleteAllForUser(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, v := range r.byHash {
		if v.UserID == userID {
			delete(r.byHash, k)
		}
	}
}
