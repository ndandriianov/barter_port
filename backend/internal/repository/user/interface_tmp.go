package user

import "github.com/ndandriianov/barter_port/backend/internal/model"

type UserRepository interface {
	Create(u model.User) error
	GetByEmail(email string) (model.User, error)
	GetByID(id string) (model.User, error)
	VerifyEmail(userID string) error
}
