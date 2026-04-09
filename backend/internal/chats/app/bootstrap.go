package app

import (
	authpb "barter-port/contracts/grpc/auth/v1"
	userspb "barter-port/contracts/grpc/users/v1"
	"barter-port/pkg/bootstrap"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitAuthGRPCClient(cfg bootstrap.Config) (authpb.AuthServiceClient, *grpc.ClientConn, error) {
	if cfg.AuthGRPCAddr == "" {
		return nil, nil, fmt.Errorf("auth grpc address is not configured")
	}

	conn, err := grpc.NewClient(cfg.AuthGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create auth grpc connection: %w", err)
	}

	return authpb.NewAuthServiceClient(conn), conn, nil
}

func InitUsersGRPCClient(cfg bootstrap.Config) (userspb.UsersServiceClient, *grpc.ClientConn, error) {
	if cfg.UsersGRPCAddr == "" {
		return nil, nil, fmt.Errorf("users grpc address is not configured")
	}

	conn, err := grpc.NewClient(cfg.UsersGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create users grpc connection: %w", err)
	}

	return userspb.NewUsersServiceClient(conn), conn, nil
}
