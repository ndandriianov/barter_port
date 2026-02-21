package refresh_token

import (
	"errors"
	"sync"

	"github.com/ndandriianov/barter_port/backend/internal/model"
)

var ErrRefreshNotFound = errors.New("refresh token not found")

type RefreshTokenRepository interface {
	Save(t model.RefreshToken) error
	GetByHash(hash string) (model.RefreshToken, error)
	Revoke(hash string) error
	DeleteAllForUser(userID string) error
}

type InMemoryRefreshRepo struct {
	mu     sync.RWMutex
	byHash map[string]model.RefreshToken
}

func NewInMemoryRefreshRepo() *InMemoryRefreshRepo {
	return &InMemoryRefreshRepo{
		byHash: make(map[string]model.RefreshToken),
	}
}

func (r *InMemoryRefreshRepo) Save(t model.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.byHash[t.TokenHash] = t
	return nil
}

func (r *InMemoryRefreshRepo) GetByHash(hash string) (model.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.byHash[hash]
	if !ok {
		return model.RefreshToken{}, ErrRefreshNotFound
	}
	return t, nil
}

func (r *InMemoryRefreshRepo) Revoke(hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.byHash[hash]
	if !ok {
		return ErrRefreshNotFound
	}
	t.Revoked = true
	r.byHash[hash] = t
	return nil
}

func (r *InMemoryRefreshRepo) DeleteAllForUser(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, v := range r.byHash {
		if v.UserID == userID {
			delete(r.byHash, k)
		}
	}
	return nil
}
