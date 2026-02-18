package user

import (
	"sync"

	"github.com/ndandriianov/barter_port/backend/internal/errors"
	"github.com/ndandriianov/barter_port/backend/internal/model"
)

type InMemoryUserRepo struct {
	mu      sync.RWMutex
	byID    map[string]model.User
	byEmail map[string]string // email -> userID
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		byID:    make(map[string]model.User),
		byEmail: make(map[string]string),
	}
}

func (r *InMemoryUserRepo) Create(u model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byEmail[u.Email]; ok {
		return errors.ErrEmailAlreadyInUse
	}

	r.byID[u.ID] = u
	r.byEmail[u.Email] = u.ID
	return nil
}

func (r *InMemoryUserRepo) GetByEmail(email string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byEmail[email]
	if !ok {
		return model.User{}, errors.ErrUserNotFound
	}

	u, ok := r.byID[id]
	if !ok {
		return model.User{}, errors.ErrUserNotFound
	}

	return u, nil
}

func (r *InMemoryUserRepo) GetByID(id string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.byID[id]
	if !ok {
		return model.User{}, errors.ErrUserNotFound
	}
	return u, nil
}

func (r *InMemoryUserRepo) VerifyEmail(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.byID[userID]
	if !ok {
		return errors.ErrUserNotFound
	}

	u.EmailVerified = true
	r.byID[userID] = u
	return nil
}
