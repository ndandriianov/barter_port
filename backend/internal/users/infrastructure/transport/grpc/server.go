package grpc

import (
	userspb "barter-port/contracts/grpc/users/v1"
	"barter-port/internal/users/application/user"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Server struct {
	userspb.UnimplementedUsersServiceServer
	usersService *user.Service
}

func NewServer(usersService *user.Service) *Server {
	return &Server{usersService: usersService}
}

func (s Server) GetUsersWithInfo(ctx context.Context, request *userspb.GetUsersWithInfoRequest) (*userspb.GetUsersWithInfoResponse, error) {
	ids := make([]uuid.UUID, len(request.Ids))
	for i, id := range request.Ids {
		parsedId, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user id %s: %w", id, err)
		}
		ids[i] = parsedId
	}

	names, err := s.usersService.GetNamesForUserIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get names for user ids: %w", err)
	}

	info := make([]*userspb.UserInfo, len(request.Ids))
	for i, id := range request.Ids {
		if names[ids[i]] == nil {
			continue
		}
		info[i] = &userspb.UserInfo{
			Id:   id,
			Name: *names[ids[i]],
		}
	}

	return &userspb.GetUsersWithInfoResponse{Users: info}, nil
}
