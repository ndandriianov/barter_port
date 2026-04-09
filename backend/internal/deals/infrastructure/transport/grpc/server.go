package grpc

import (
	dealspb "barter-port/contracts/grpc/deals/v1"
	"barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/domain"
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	dealspb.UnimplementedDealsServiceServer
	dealsService *deals.Service
}

func NewServer(dealsService *deals.Service) *Server {
	return &Server{dealsService: dealsService}
}

func (s *Server) GetDealStatus(ctx context.Context, req *dealspb.GetDealStatusRequest) (*dealspb.GetDealStatusResponse, error) {
	if req.GetDealId() == "" {
		return nil, status.Error(codes.InvalidArgument, "deal_id is required")
	}

	dealID, err := uuid.Parse(req.GetDealId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid deal_id: %v", err)
	}

	dealStatus, err := s.dealsService.GetDealStatus(ctx, dealID)
	if err != nil {
		if errors.Is(err, domain.ErrDealNotFound) {
			return nil, status.Error(codes.NotFound, "deal not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to get deal status: %v", err)
	}

	return &dealspb.GetDealStatusResponse{Status: dealStatus.String()}, nil
}
