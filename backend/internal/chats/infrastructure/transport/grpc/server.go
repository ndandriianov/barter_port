package grpc

import (
	chatspb "barter-port/contracts/grpc/chats/v1"
	"barter-port/internal/chats/application"
	"barter-port/internal/chats/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	chatspb.UnimplementedChatsServiceServer
	chatsService *application.Service
}

func NewServer(chatsService *application.Service) *Server {
	return &Server{chatsService: chatsService}
}

func (s *Server) CreateChat(ctx context.Context, req *chatspb.CreateChatRequest) (*chatspb.CreateChatResponse, error) {
	if len(req.ParticipantIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "participant_ids is required")
	}

	participantIDs := make([]uuid.UUID, len(req.ParticipantIds))
	for i, id := range req.ParticipantIds {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid participant id %s: %v", id, err)
		}
		participantIDs[i] = parsed
	}

	var dealID *uuid.UUID
	if req.DealId != "" {
		parsed, err := uuid.Parse(req.DealId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid deal_id: %v", err)
		}
		dealID = &parsed
	}

	chat, err := s.chatsService.CreateChat(ctx, dealID, participantIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("failed to create chat: %v", err))
	}

	return &chatspb.CreateChatResponse{ChatId: chat.ID.String()}, nil
}

func (s *Server) GetDealChatId(ctx context.Context, req *chatspb.GetDealChatIdRequest) (*chatspb.GetDealChatIdResponse, error) {
	if req.DealId == "" {
		return nil, status.Error(codes.InvalidArgument, "deal_id is required")
	}

	dealID, err := uuid.Parse(req.DealId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid deal_id: %v", err)
	}

	chatID, err := s.chatsService.GetDealChatID(ctx, dealID)
	if err != nil {
		if errors.Is(err, domain.ErrChatNotFound) {
			return nil, status.Error(codes.NotFound, "chat for deal not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get chat id for deal: %v", err)
	}

	return &chatspb.GetDealChatIdResponse{ChatId: chatID.String()}, nil
}
