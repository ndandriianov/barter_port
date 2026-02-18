package token

import "github.com/ndandriianov/barter_port/backend/internal/model"

type TokenRepository interface {
	Save(t model.EmailVerificationToken) error
	GetByHash(tokenHash string) (model.EmailVerificationToken, error)
	MarkUsed(tokenHash string) error
	DeleteAllForUser(userID string) error
}
