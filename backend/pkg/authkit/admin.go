package authkit

import (
	authpb "barter-port/contracts/grpc/auth/v1"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type AdminChecker struct {
	authClient authpb.AuthServiceClient
}

func NewAdminChecker(authClient authpb.AuthServiceClient) *AdminChecker {
	return &AdminChecker{authClient: authClient}
}

// IsAdmin asks the auth service whether the user is the configured admin.
func (c *AdminChecker) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	if c == nil || c.authClient == nil {
		return false, ErrAuthClientNotConfigured
	}

	resp, err := c.authClient.GetMe(ctx, &authpb.GetMeRequest{Id: userID.String()})
	if err != nil {
		return false, fmt.Errorf("auth grpc get me: %w", err)
	}

	return resp.GetIsAdmin(), nil
}
