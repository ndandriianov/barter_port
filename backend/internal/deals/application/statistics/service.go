package statistics

import (
	"context"

	statsrepo "barter-port/internal/deals/infrastructure/repository/statistics"

	"github.com/google/uuid"
)

type Service struct {
	repo *statsrepo.Repository
}

func NewService(repo *statsrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMyStatistics(ctx context.Context, userID uuid.UUID) (*statsrepo.Result, error) {
	return s.repo.GetMyStatistics(ctx, userID)
}
