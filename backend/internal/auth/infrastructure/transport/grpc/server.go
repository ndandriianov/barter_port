package grpc

import (
	authpb "barter-port/contracts/grpc/auth/v1"
	"barter-port/internal/auth/application"
	"barter-port/internal/auth/domain"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	authpb.UnimplementedAuthServiceServer
	authService *application.Service
}

func NewServer(authService *application.Service) *Server {
	return &Server{authService: authService}
}

func (s *Server) GetMe(ctx context.Context, req *authpb.GetMeRequest) (*authpb.GetMeResponse, error) {
	if req == nil || strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	me, err := s.authService.GetMe(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return toGetMeResponse(me), nil
}

func toGetMeResponse(user domain.User) *authpb.GetMeResponse {
	return &authpb.GetMeResponse{
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
}
