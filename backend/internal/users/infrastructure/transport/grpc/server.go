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
		curInfo := u.GetInfo()
		info[i] = &userspb.UserInfo{
			Id:   curInfo.Id.String(),
			Name: curInfo.Name,
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

func (s *Server) GetUserLocation(ctx context.Context, request *userspb.GetUserLocationRequest) (*userspb.GetUserLocationResponse, error) {
	userID, err := uuid.Parse(request.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id %s: %w", request.GetUserId(), err)
	}

	lat, lon, err := s.usersService.GetLocation(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user location: %w", err)
	}

	resp := &userspb.GetUserLocationResponse{}
	if lat != nil {
		resp.Latitude = lat
	}
	if lon != nil {
		resp.Longitude = lon
	}
	return resp, nil
}

func (s *Server) ListSubscriptions(ctx context.Context, request *userspb.ListSubscriptionsRequest) (*userspb.ListSubscriptionsResponse, error) {
	if request == nil {
		return &userspb.ListSubscriptionsResponse{}, fmt.Errorf("request is nil")
	}

	userID, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id %s: %w", request.UserId, err)
	}

	userInfos, err := s.usersService.GetSubscriptionsUserInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user infos: %w", err)
	}

	info := make([]*userspb.UserInfo, len(userInfos))
	for i, u := range userInfos {
		info[i] = &userspb.UserInfo{
			Id:   u.Id.String(),
			Name: u.Name,
		}
	}

	return &userspb.ListSubscriptionsResponse{Subscriptions: info}, nil
}

func (s *Server) ListHiddenUsers(ctx context.Context, request *userspb.ListHiddenUsersRequest) (*userspb.ListHiddenUsersResponse, error) {
	if request == nil {
		return &userspb.ListHiddenUsersResponse{}, fmt.Errorf("request is nil")
	}

	userID, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id %s: %w", request.UserId, err)
	}

	userInfos, err := s.usersService.GetHiddenUsersUserInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hidden user infos: %w", err)
	}

	info := make([]*userspb.UserInfo, len(userInfos))
	for i, u := range userInfos {
		info[i] = &userspb.UserInfo{
			Id:   u.Id.String(),
			Name: u.Name,
		}
	}

	return &userspb.ListHiddenUsersResponse{Users: info}, nil
}

func (s *Server) IsUserHiddenByAnyOwners(
	ctx context.Context,
	request *userspb.IsUserHiddenByAnyOwnersRequest,
) (*userspb.IsUserHiddenByAnyOwnersResponse, error) {
	if request == nil {
		return &userspb.IsUserHiddenByAnyOwnersResponse{}, fmt.Errorf("request is nil")
	}

	hiddenUserID, err := uuid.Parse(request.HiddenUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hidden user id %s: %w", request.HiddenUserId, err)
	}

	ownerUserIDs := make([]uuid.UUID, len(request.OwnerUserIds))
	for i, ownerUserID := range request.OwnerUserIds {
		parsedID, parseErr := uuid.Parse(ownerUserID)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse owner user id %s: %w", ownerUserID, parseErr)
		}
		ownerUserIDs[i] = parsedID
	}

	isHidden, err := s.usersService.IsUserHiddenByAnyOwners(ctx, hiddenUserID, ownerUserIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to check hidden user by owners: %w", err)
	}

	return &userspb.IsUserHiddenByAnyOwnersResponse{IsHidden: isHidden}, nil
}
