package bootstrap

import (
	userspb "barter-port/contracts/grpc/users/v1"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initUsersGRPCClient(cfg Config) (userspb.UsersServiceClient, *grpc.ClientConn, error) {
	if cfg.UsersGRPCAddr == "" {
		return nil, nil, fmt.Errorf("failed to initialize grpc server: users grpc address is not configured")
	}

	conn, err := grpc.NewClient(cfg.AuthGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create auth grpc connection: %w", err)
	}

	return userspb.NewUsersServiceClient(conn), conn, nil
}
