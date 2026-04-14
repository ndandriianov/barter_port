package consumer

import (
	dealsusers "barter-port/contracts/kafka/messages/deals-users"
	reputation_inbox "barter-port/internal/users/infrastructure/repository/reputation-inbox"
	"barter-port/pkg/db"
	"barter-port/pkg/kafkax"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReputationInboxConsumer struct {
	db        *pgxpool.Pool
	inboxRepo reputationInboxWriter
	consumer  *kafkax.InboxConsumer[dealsusers.OfferReportPenaltyMessage]
}

type reputationInboxWriter interface {
	WriteReputationInboxMessage(context.Context, db.DB, dealsusers.OfferReportPenaltyMessage) error
}

func NewReputationInboxConsumer(
	db *pgxpool.Pool,
	inboxRepo reputationInboxWriter,
	consumer *kafkax.InboxConsumer[dealsusers.OfferReportPenaltyMessage],
) *ReputationInboxConsumer {
	return &ReputationInboxConsumer{
		db:        db,
		inboxRepo: inboxRepo,
		consumer:  consumer,
	}
}

func (c *ReputationInboxConsumer) Run(ctx context.Context) error {
	return c.consumer.Run(ctx, c.consumeMessage)
}

func (c *ReputationInboxConsumer) consumeMessage(ctx context.Context, message dealsusers.OfferReportPenaltyMessage) error {
	err := c.inboxRepo.WriteReputationInboxMessage(ctx, c.db, message)
	if err != nil {
		if errors.Is(err, reputation_inbox.ErrReputationEventAlreadyExists) {
			return kafkax.ErrDuplicate
		}
		return fmt.Errorf("failed to write reputation inbox message: %w", err)
	}
	return nil
}
