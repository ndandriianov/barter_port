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

func (s *Server) ListUsers(ctx context.Context, _ *userspb.ListUsersRequest) (*userspb.ListUsersResponse, error) {
	users, err := s.usersService.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	info := make([]*userspb.UserInfo, len(users))
	for i, u := range users {
		name := ""
		if u.Name != nil {
			name = *u.Name
		}
		info[i] = &userspb.UserInfo{
			Id:   u.Id.String(),
			Name: name,
		}
	}

	return &userspb.ListUsersResponse{Users: info}, nil
}

func (s *Server) ListUsersForChatCreation(
	ctx context.Context,
	request *userspb.ListUsersForChatCreationRequest,
) (*userspb.ListUsersForChatCreationResponse, error) {

	id, err := uuid.Parse(request.RequesterUserId)
	if err != nil {
		return nil, fmt.Errorf("parse user id %s: %w", request.RequesterUserId, err)
	}

	users, err := s.usersService.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	info := make([]*userspb.UserInfo, 0, len(users))
	for _, u := range users {
		ok, err := s.usersService.CanCreateChat(ctx, id, u.Id)
		if err != nil {
			return nil, fmt.Errorf("can create chat: %w", err)
		}

		if !ok {
			continue
		}

		name := ""
		if u.Name != nil {
			name = *u.Name
		}
		info = append(info, &userspb.UserInfo{
			Id:   u.Id.String(),
			Name: name,
		})
	}

	return &userspb.ListUsersForChatCreationResponse{Users: info}, nil
}

func (s *Server) GetUsersWithInfo(ctx context.Context, request *userspb.GetUsersWithInfoRequest) (*userspb.GetUsersWithInfoResponse, error) {
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

func (s *Server) CheckSubscription(ctx context.Context, request *userspb.CheckSubscriptionRequest) (*userspb.CheckSubscriptionResponse, error) {
	requesterUserId, err := uuid.Parse(request.RequesterUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id %s: %w", request.RequesterUserId, err)
	}

	targetUserId, err := uuid.Parse(request.TargetUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id %s: %w", request.TargetUserId, err)
	}

	isTargetSubscribed, hasCreatedSubscription, err := s.usersService.CheckSubscription(ctx, requesterUserId, targetUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to check subscription: %w", err)
	}

	return &userspb.CheckSubscriptionResponse{
		IsSubscribed:           isTargetSubscribed,
		HasCreatedSubscription: hasCreatedSubscription,
	}, nil
}
