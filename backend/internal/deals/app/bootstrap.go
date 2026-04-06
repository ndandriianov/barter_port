package app

import (
	chatspb "barter-port/contracts/grpc/chats/v1"
	userspb "barter-port/contracts/grpc/users/v1"
	"barter-port/pkg/bootstrap"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitUsersGRPCClient(cfg bootstrap.Config) (userspb.UsersServiceClient, *grpc.ClientConn, error) {
	if cfg.UsersGRPCAddr == "" {
		return nil, nil, fmt.Errorf("failed to initialize grpc client: users grpc address is not configured")
	}

	conn, err := grpc.NewClient(cfg.UsersGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create auth grpc connection: %w", err)
	}

	return userspb.NewUsersServiceClient(conn), conn, nil
}

func InitChatsGRPCClient(cfg bootstrap.Config) (chatspb.ChatsServiceClient, *grpc.ClientConn, error) {
	if cfg.ChatsGRPCAddr == "" {
		return nil, nil, fmt.Errorf("chats grpc address is not configured")
	}

	conn, err := grpc.NewClient(cfg.ChatsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create chats grpc connection: %w", err)
	}

	return chatspb.NewChatsServiceClient(conn), conn, nil
}
