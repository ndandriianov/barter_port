package user

import (
	"errors"
	"sync"

	"barter-port/internal/model"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyInUse = errors.New("email already in use")
)

type InMemoryUserRepo struct {
	mu      sync.RWMutex
	byID    map[uuid.UUID]model.User
	byEmail map[string]uuid.UUID // email -> userID
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		byID:    make(map[uuid.UUID]model.User),
		byEmail: make(map[string]uuid.UUID),
	}
}

// Create adds a new user to the repository.
// Errors:
//   - errors.ErrEmailAlreadyInUse - email already exists
func (r *InMemoryUserRepo) Create(u model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byEmail[u.Email]; ok {
		return ErrEmailAlreadyInUse
	}

	r.byID[u.ID] = u
	r.byEmail[u.Email] = u.ID
	return nil
}

// GetByEmail retrieves a user by their email address.
// Errors:
//   - errors.ErrUserNotFound: Occurs if no user is found with the given email address.
func (r *InMemoryUserRepo) GetByEmail(email string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byEmail[email]
	if !ok {
		return model.User{}, ErrUserNotFound
	}

	u, ok := r.byID[id]
	if !ok {
		return model.User{}, ErrUserNotFound
	}

	return u, nil
}

// GetByID retrieves a user by their unique ID.
// Errors:
//   - errors.ErrUserNotFound: Occurs if no user is found with the given ID.
func (r *InMemoryUserRepo) GetByID(id uuid.UUID) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.byID[id]
	if !ok {
		return model.User{}, ErrUserNotFound
	}
	return u, nil
}

// VerifyEmail marks a user's email as verified.
// Errors:
//   - errors.ErrUserNotFound: Occurs if no user is found with the given userID.
func (r *InMemoryUserRepo) VerifyEmail(userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.byID[userID]
	if !ok {
		return ErrUserNotFound
	}

	u.EmailVerified = true
	r.byID[userID] = u
	return nil
}
