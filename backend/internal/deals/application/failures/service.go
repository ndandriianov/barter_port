package failures

import (
	chatspb "barter-port/contracts/grpc/chats/v1"
	dealsusers "barter-port/contracts/kafka/messages/deals-users"
	appdeals "barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	failuresrepo "barter-port/internal/deals/infrastructure/repository/failures"
	outbox "barter-port/internal/deals/infrastructure/repository/offer-report-outbox"
	"barter-port/pkg/db"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type Service struct {
	*appdeals.Service
	repository *failuresrepo.Repository
	outbox     *outbox.Repository
}

func NewService(base *appdeals.Service, repository *failuresrepo.Repository) *Service {
	return &Service{Service: base, repository: repository}
}

func (s *Service) GetDealsForFailureReview(
	ctx context.Context,
	userID uuid.UUID,
) ([]htypes.DealIDWithParticipantIDs, error) {
	isAdmin, err := s.isAdmin(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, domain.ErrAdminOnly
	}

	return s.repository.GetFailureReviewDeals(ctx, s.DB())
}

func (s *Service) VoteForFailure(
	ctx context.Context,
	dealID, voterID, votedForID uuid.UUID,
) error {
	return db.RunInTx(ctx, s.DB(), func(ctx context.Context, tx pgx.Tx) error {
		deal, err := s.lockedDealForFailureVote(ctx, tx, dealID)
		if err != nil {
			return err
		}
		if !appdeals.ContainsUserID(deal.Participants, voterID) || !appdeals.ContainsUserID(deal.Participants, votedForID) {
			return domain.ErrForbidden
		}

		if err = s.ensureFailureVotingOpen(ctx, tx, dealID); err != nil {
			return err
		}

		if err = s.repository.SetFailureVote(ctx, tx, dealID, voterID, votedForID); err != nil {
			return err
		}

		votes, err := s.repository.GetFailureVotes(ctx, tx, dealID)
		if err != nil {
			return err
		}

		threshold := failureVoteThreshold(len(deal.Participants))
		if len(votes) < threshold {
			return nil
		}

		blamedUserID := failureVoteWinner(votes, threshold)
		return s.repository.CreateFailureRecord(ctx, tx, dealID, blamedUserID)
	})
}

func (s *Service) RevokeVoteForFailure(ctx context.Context, dealID, userID uuid.UUID) error {
	return db.RunInTx(ctx, s.DB(), func(ctx context.Context, tx pgx.Tx) error {
		deal, err := s.lockedDealForFailureVote(ctx, tx, dealID)
		if err != nil {
			return err
		}
		if !appdeals.ContainsUserID(deal.Participants, userID) {
			return domain.ErrForbidden
		}

		if err = s.ensureFailureVotingOpen(ctx, tx, dealID); err != nil {
			return err
		}

		return s.repository.ClearFailureVote(ctx, tx, dealID, userID)
	})
}

func (s *Service) GetFailureVotes(
	ctx context.Context,
	dealID, userID uuid.UUID,
) ([]htypes.FailureVote, error) {
	deal, err := s.GetDealByID(ctx, dealID)
	if err != nil {
		return nil, err
	}

	if !appdeals.ContainsUserID(deal.Participants, userID) {
		isAdmin, adminErr := s.isAdmin(ctx, userID)
		if adminErr != nil {
			return nil, adminErr
		}
		if !isAdmin {
			return nil, domain.ErrForbidden
		}
	}

	return s.repository.GetFailureVotes(ctx, s.DB(), dealID)
}

func (s *Service) GetFailureMaterials(
	ctx context.Context,
	dealID, userID uuid.UUID,
) (htypes.FailureMaterials, error) {
	isAdmin, err := s.isAdmin(ctx, userID)
	if err != nil {
		return htypes.FailureMaterials{}, err
	}
	if !isAdmin {
		return htypes.FailureMaterials{}, domain.ErrAdminOnly
	}

	deal, err := s.GetDealByID(ctx, dealID)
	if err != nil {
		return htypes.FailureMaterials{}, err
	}

	hasFailure, err := s.repository.HasFailureRecord(ctx, s.DB(), dealID)
	if err != nil {
		return htypes.FailureMaterials{}, err
	}
	if !hasFailure {
		return htypes.FailureMaterials{}, domain.ErrForbidden
	}

	if deal.Status == enums.DealStatusCompleted || deal.Status == enums.DealStatusCancelled {
		return htypes.FailureMaterials{}, domain.ErrForbidden
	}

	result := htypes.FailureMaterials{Deal: deal}
	if s.ChatsClient() == nil {
		return result, nil
	}

	resp, err := s.ChatsClient().GetDealChatId(ctx, &chatspb.GetDealChatIdRequest{DealId: dealID.String()})
	if err != nil {
		switch grpcstatus.Code(err) {
		case codes.NotFound, codes.Unavailable:
			return result, nil
		default:
			return htypes.FailureMaterials{}, fmt.Errorf("get deal chat id: %w", err)
		}
	}
	if chatID := resp.GetChatId(); chatID != "" {
		parsedID, parseErr := uuid.Parse(chatID)
		if parseErr != nil {
			return htypes.FailureMaterials{}, fmt.Errorf("parse deal chat id: %w", parseErr)
		}
		result.ChatID = &parsedID
	}

	return result, nil
}

func (s *Service) ModeratorResolutionForFailure(
	ctx context.Context,
	dealID, userID uuid.UUID,
	confirmed bool,
	failureUserID *uuid.UUID,
	punishmentPoints *int,
	comment *string,
) (htypes.FailureRecord, error) {
	isAdmin, err := s.isAdmin(ctx, userID)
	if err != nil {
		return htypes.FailureRecord{}, err
	}
	if !isAdmin {
		return htypes.FailureRecord{}, domain.ErrAdminOnly
	}

	if !confirmed && (failureUserID != nil || punishmentPoints != nil) {
		return htypes.FailureRecord{}, domain.ErrInvalidFailureDecision
	}
	if confirmed && (punishmentPoints == nil || *punishmentPoints < 0 || failureUserID == nil) {
		return htypes.FailureRecord{}, domain.ErrInvalidFailureDecision
	}

	now := time.Now()

	var record htypes.FailureRecord
	err = db.RunInTx(ctx, s.DB(), func(ctx context.Context, tx pgx.Tx) error {
		if _, err = s.repository.LockDeal(ctx, tx, dealID); err != nil {
			return err
		}

		deal, err := s.DealsRepository().GetDealByID(ctx, tx, dealID)
		if err != nil {
			return err
		}

		if failureUserID != nil && !appdeals.ContainsUserID(deal.Participants, *failureUserID) {
			return domain.ErrForbidden
		}

		record, err = s.repository.ResolveFailureRecord(
			ctx,
			tx,
			dealID,
			confirmed,
			failureUserID,
			punishmentPoints,
			comment,
		)
		if err != nil {
			return err
		}

		targetStatus := enums.DealStatusFailed
		if !confirmed {
			if deal.Status == enums.DealStatusConfirmed {
				targetStatus = enums.DealStatusCompleted
			} else {
				targetStatus = enums.DealStatusCancelled
			}
		}

		outboxMsg := dealsusers.ReputationMessage{
			ID:         uuid.New(),
			SourceType: dealsusers.DealFailureResponsibleMessageType,
			SourceID:   dealID,
			UserID:     *failureUserID,
			Delta:      -*punishmentPoints,
			CreatedAt:  now,
			Comment:    comment,
		}

		err = s.outbox.WriteOutboxMessage(ctx, tx, outboxMsg)
		if err != nil {
			return err
		}

		if err = s.DealsRepository().UpdateDealStatus(ctx, tx, dealID, targetStatus); err != nil {
			return err
		}
		return s.DealsRepository().DeleteStatusVotes(ctx, tx, dealID)
	})
	if err != nil {
		return htypes.FailureRecord{}, err
	}

	return record, nil
}

func (s *Service) GetModeratorResolutionForFailure(
	ctx context.Context,
	dealID, userID uuid.UUID,
) (htypes.FailureRecord, error) {
	deal, err := s.GetDealByID(ctx, dealID)
	if err != nil {
		return htypes.FailureRecord{}, err
	}

	if !appdeals.ContainsUserID(deal.Participants, userID) {
		isAdmin, adminErr := s.isAdmin(ctx, userID)
		if adminErr != nil {
			return htypes.FailureRecord{}, adminErr
		}
		if !isAdmin {
			return htypes.FailureRecord{}, domain.ErrForbidden
		}
	}

	record, err := s.repository.GetFailureRecord(ctx, s.DB(), dealID)
	if errors.Is(err, domain.ErrFailureNotFound) {
		return htypes.FailureRecord{}, domain.ErrForbidden
	}
	if err != nil {
		return htypes.FailureRecord{}, err
	}

	return record, nil
}

func (s *Service) ensureFailureVotingOpen(ctx context.Context, tx pgx.Tx, dealID uuid.UUID) error {
	record, err := s.repository.GetFailureRecord(ctx, tx, dealID)
	if err == nil {
		if record.ConfirmedByAdmin != nil {
			return domain.ErrFailureAlreadyResolved
		}
		return domain.ErrFailureReviewRequired
	}
	if errors.Is(err, domain.ErrFailureNotFound) {
		return nil
	}
	return err
}

func (s *Service) lockedDealForFailureVote(ctx context.Context, tx pgx.Tx, dealID uuid.UUID) (domain.Deal, error) {
	status, err := s.repository.LockDeal(ctx, tx, dealID)
	if err != nil {
		return domain.Deal{}, err
	}
	if status != enums.DealStatusDiscussion &&
		status != enums.DealStatusConfirmed {
		return domain.Deal{}, domain.ErrInvalidDealStatus
	}

	return s.DealsRepository().GetDealByID(ctx, tx, dealID)
}

func (s *Service) isAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	if s.AdminChecker() == nil {
		return false, fmt.Errorf("admin checker is not configured")
	}

	return s.AdminChecker().IsAdmin(ctx, userID)
}

func failureVoteThreshold(participantsCount int) int {
	threshold := participantsCount / 2
	if threshold < 1 {
		return 1
	}
	return threshold
}

func failureVoteWinner(votes []htypes.FailureVote, threshold int) *uuid.UUID {
	counts := make(map[uuid.UUID]int, len(votes))
	for _, vote := range votes {
		counts[vote.Vote]++
		if counts[vote.Vote] >= threshold {
			return new(vote.Vote)
		}
	}

	return nil
}
