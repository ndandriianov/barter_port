package statistics

import (
	authpb "barter-port/contracts/grpc/auth/v1"
	"barter-port/internal/deals/domain"
	"context"
	"fmt"

	statsrepo "barter-port/internal/deals/infrastructure/repository/statistics"
	"barter-port/pkg/authkit"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	repo         *statsrepo.Repository
	authClient   authpb.AuthServiceClient
	adminChecker *authkit.AdminChecker
}

func NewService(repo *statsrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) WithAuthClient(authClient authpb.AuthServiceClient) *Service {
	s.authClient = authClient
	s.adminChecker = authkit.NewAdminChecker(authClient)
	return s
}

func (s *Service) GetMyStatistics(ctx context.Context, userID uuid.UUID) (*statsrepo.Result, error) {
	return s.repo.GetMyStatistics(ctx, userID)
}

func (s *Service) GetAdminPlatformStatistics(
	ctx context.Context,
	requesterID uuid.UUID,
) (*statsrepo.AdminPlatformStatisticsResult, error) {
	if err := s.ensureAdmin(ctx, requesterID); err != nil {
		return nil, err
	}

	return s.repo.GetAdminPlatformStatistics(ctx)
}

func (s *Service) GetAdminUserStatistics(
	ctx context.Context,
	requesterID uuid.UUID,
	targetUserID uuid.UUID,
) (*statsrepo.AdminUserStatisticsResult, error) {
	if err := s.ensureAdmin(ctx, requesterID); err != nil {
		return nil, err
	}
	if err := s.ensureUserExists(ctx, targetUserID); err != nil {
		return nil, err
	}

	return s.repo.GetAdminUserStatistics(ctx, targetUserID)
}

func (s *Service) ensureAdmin(ctx context.Context, requesterID uuid.UUID) error {
	if s.adminChecker == nil {
		return fmt.Errorf("admin checker is not configured")
	}

	isAdmin, err := s.adminChecker.IsAdmin(ctx, requesterID)
	if err != nil {
		return fmt.Errorf("adminChecker.IsAdmin: %w", err)
	}
	if !isAdmin {
		return domain.ErrForbidden
	}

	return nil
}

func (s *Service) ensureUserExists(ctx context.Context, targetUserID uuid.UUID) error {
	if s.authClient == nil {
		return fmt.Errorf("auth client is not configured")
	}

	_, err := s.authClient.GetMe(ctx, &authpb.GetMeRequest{Id: targetUserID.String()})
	if err == nil {
		return nil
	}
	if status.Code(err) == codes.NotFound {
		return domain.ErrUserNotFound
	}

	return fmt.Errorf("auth grpc get me: %w", err)
}
