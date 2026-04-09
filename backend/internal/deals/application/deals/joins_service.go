package deals

import (
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// JoinDeal creates a join request for the specified user and deal.
func (s *Service) JoinDeal(ctx context.Context, dealID, userID uuid.UUID) error {
	return db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		deal, err := s.dealsRepository.GetDealByID(ctx, tx, dealID)
		if err != nil {
			return err
		}

		if deal.Status != enums.DealStatusLookingForParticipants {
			return domain.ErrInvalidDealStatus
		}

		for _, participantID := range deal.Participants {
			if participantID == userID {
				return domain.ErrForbidden
			}
		}

		err = s.joinsRepository.CreateJoinRequest(ctx, tx, userID, dealID)
		if repox.IsUniqueViolation(err) {
			return nil
		}
		return err
	})
}

// LeaveDeal removes user's join request from the deal.
func (s *Service) LeaveDeal(ctx context.Context, dealID, userID uuid.UUID) error {
	return db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		deal, err := s.dealsRepository.GetDealByID(ctx, tx, dealID)
		if err != nil {
			return err
		}

		if deal.Status != enums.DealStatusLookingForParticipants {
			return domain.ErrInvalidDealStatus
		}

		userParticipating := false
		for _, participantID := range deal.Participants {
			if participantID == userID {
				userParticipating = true
			}
		}

		if userParticipating {
			return s.dealsRepository.DeleteParticipant(ctx, tx, dealID, userID)
		}

		return s.joinsRepository.DeleteJoinRequest(ctx, tx, userID, dealID)
	})
}

// GetDealJoinRequests returns deal join requests with IDs of users that voted in favor.
func (s *Service) GetDealJoinRequests(ctx context.Context, dealID, userID uuid.UUID) ([]htypes.JoinRequestWithVoters, error) {
	var result []htypes.JoinRequestWithVoters

	txErr := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		deal, err := s.dealsRepository.GetDealByID(ctx, tx, dealID)
		if err != nil {
			return err
		}

		if !containsUserID(deal.Participants, userID) {
			return domain.ErrForbidden
		}

		requests, err := s.joinsRepository.GetJoinRequestsByDealID(ctx, tx, dealID)
		if err != nil {
			return err
		}
		votes, err := s.joinsRepository.GetJoinRequestVotesByDealID(ctx, tx, dealID)
		if err != nil {
			return err
		}

		votersByUser := make(map[uuid.UUID][]uuid.UUID, len(requests))
		for _, v := range votes {
			votersByUser[v.UserID] = append(votersByUser[v.UserID], v.VoterID)
		}

		result = make([]htypes.JoinRequestWithVoters, 0, len(requests))
		for _, req := range requests {
			result = append(result, htypes.JoinRequestWithVoters{
				UserID:   req.UserID,
				DealID:   req.DealID,
				VoterIDs: votersByUser[req.UserID],
			})
		}

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return result, nil
}

// ProcessJoinRequest saves participant's vote and applies the final decision when all participants voted.
func (s *Service) ProcessJoinRequest(ctx context.Context, dealID, requestedUserID, voterID uuid.UUID, accept bool) error {
	return db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		deal, err := s.dealsRepository.GetDealByID(ctx, tx, dealID)
		if err != nil {
			return err
		}
		if deal.Status != enums.DealStatusLookingForParticipants {
			return domain.ErrInvalidDealStatus
		}

		if !containsUserID(deal.Participants, voterID) {
			return domain.ErrForbidden
		}

		requests, err := s.joinsRepository.GetJoinRequestsByDealID(ctx, tx, dealID)
		if err != nil {
			return err
		}
		if !containsJoinRequest(requests, requestedUserID) {
			return domain.ErrJoinRequestNotFound
		}

		if !accept {
			err = s.joinsRepository.DeleteJoinRequest(ctx, tx, requestedUserID, dealID)
			if err != nil {
				return fmt.Errorf("delete join request: %w", err)
			}
			return nil
		}

		if err = s.joinsRepository.UpsertJoinRequestVote(ctx, tx, requestedUserID, dealID, voterID); err != nil {
			return fmt.Errorf("upsert join request vote: %w", err)
		}

		votes, err := s.joinsRepository.GetJoinRequestVotesByDealID(ctx, tx, dealID)
		if err != nil {
			return err
		}

		if len(votes) == len(deal.Participants) {
			if err = s.joinsRepository.AddParticipant(ctx, tx, dealID, requestedUserID); err != nil {
				return fmt.Errorf("add participant from join request: %w", err)
			}
			return s.joinsRepository.DeleteJoinRequest(ctx, tx, requestedUserID, dealID)
		}

		return nil
	})
}

func containsUserID(items []uuid.UUID, userID uuid.UUID) bool {
	for _, item := range items {
		if item == userID {
			return true
		}
	}
	return false
}

func containsJoinRequest(items []htypes.JoinRequest, userID uuid.UUID) bool {
	for _, item := range items {
		if item.UserID == userID {
			return true
		}
	}
	return false
}
