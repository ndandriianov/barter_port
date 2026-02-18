package token

import (
	"sync"

	"github.com/ndandriianov/barter_port/backend/internal/errors"
	"github.com/ndandriianov/barter_port/backend/internal/model"
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

func (r *InMemoryTokenRepo) Save(t model.EmailVerificationToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byHash[t.TokenHash]; exists {
		return errors.ErrTokenAlreadyExists
	}

	r.byHash[t.TokenHash] = t
	return nil
}

func (r *InMemoryTokenRepo) GetByHash(tokenHash string) (model.EmailVerificationToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.byHash[tokenHash]
	if !ok {
		return model.EmailVerificationToken{}, errors.ErrTokenNotFound
	}
	return t, nil
}

func (r *InMemoryTokenRepo) MarkUsed(tokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.byHash[tokenHash]
	if !ok {
		return errors.ErrTokenNotFound
	}

	t.Used = true
	r.byHash[tokenHash] = t
	return nil
}

func (r *InMemoryTokenRepo) DeleteAllForUser(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, v := range r.byHash {
		if v.UserID == userID {
			delete(r.byHash, k)
		}
	}
}
