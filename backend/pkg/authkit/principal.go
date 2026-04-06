package authkit

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Principal struct {
	UserID uuid.UUID `json:"user_id"`
	Type   TokenType `json:"type"`

	Subject   string     `json:"subject"`
	IssuedAt  *time.Time `json:"issued_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	JTI       string     `json:"jti"`
}

type contextKey string

const principalKey contextKey = "authkit_principal"
const userIDKey contextKey = "authkit_user_id"

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
