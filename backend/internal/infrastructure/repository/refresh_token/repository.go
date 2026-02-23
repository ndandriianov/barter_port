package refresh_token

import (
	"errors"
	"sync"

	"github.com/ndandriianov/barter_port/backend/internal/model"
)

var ErrRefreshNotFound = errors.New("refresh token not found")

type InMemoryRefreshRepo struct {
	mu    sync.RWMutex
	byJTI map[string]model.RefreshToken
}

func NewInMemoryRefreshRepo() *InMemoryRefreshRepo {
	return &InMemoryRefreshRepo{
		byJTI: make(map[string]model.RefreshToken),
	}
}

func (r *InMemoryRefreshRepo) Save(token model.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.byJTI[token.JTI] = token
	return nil
}

func (r *InMemoryRefreshRepo) GetByJTI(jti string) (model.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.byJTI[jti]
	if !ok {
		return model.RefreshToken{}, ErrRefreshNotFound
	}
	return t, nil
}

func (r *InMemoryRefreshRepo) Revoke(jti string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	token, ok := r.byJTI[jti]
	if !ok {
		return ErrRefreshNotFound
	}

	token.Revoked = true
	r.byJTI[jti] = token

	return nil
}

func (r *InMemoryRefreshRepo) DeleteAllForUser(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, v := range r.byJTI {
		if v.UserID == userID {
			delete(r.byJTI, k)
		}
	}
	return nil
}
