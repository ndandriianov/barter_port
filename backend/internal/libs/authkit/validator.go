package authkit

import "golang.org/x/net/context"

type Validator interface {
	ValidateAccess(ctx context.Context, token string) (Principal, error)
}
