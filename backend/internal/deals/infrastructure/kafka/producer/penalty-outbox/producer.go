package penalty_outbox

import (
	offer_report_outbox "barter-port/internal/deals/infrastructure/repository/offer-report-outbox"
	"barter-port/pkg/db"
	"barter-port/pkg/kafkax"
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PenaltyOutboxProducer struct {
	db        *pgxpool.Pool
	repo      *offer_report_outbox.Repository
	logger    *slog.Logger
	publisher *kafkax.OutboxPublisher
}

func NewPenaltyOutboxProducer(
	dbPool *pgxpool.Pool,
	repo *offer_report_outbox.Repository,
	logger *slog.Logger,
	publisher *kafkax.OutboxPublisher,
) *PenaltyOutboxProducer {
	return &PenaltyOutboxProducer{
		db:        dbPool,
		repo:      repo,
		logger:    logger,
		publisher: publisher,
	}
}

func (p *PenaltyOutboxProducer) Run(ctx context.Context) error {
	return p.publisher.Run(ctx, p.publishBatch, "failed to publish offer report penalty outbox batch")
}

func (p *PenaltyOutboxProducer) Close() error {
	return p.publisher.Close()
}

func (p *PenaltyOutboxProducer) publishBatch(ctx context.Context) (int, error) {
	var count int

	err := db.RunInTx(ctx, p.db, func(ctx context.Context, tx pgx.Tx) error {
		messages, err := p.repo.ReadOutboxMessagesForUpdate(ctx, tx, p.publisher.BatchSize())
		if err != nil {
			return fmt.Errorf("read outbox messages: %w", err)
		}
		if len(messages) == 0 {
			return nil
		}

		kafkaMessages, err := kafkax.BuildMessages(messages)
		if err != nil {
			return fmt.Errorf("build kafka messages: %w", err)
		}

		p.logger.Debug("publishing offer report penalty messages to kafka", slog.Int("count", len(messages)))

		if err = p.publisher.WriteMessages(ctx, kafkaMessages, "ensure offer-report-penalty topic"); err != nil {
			return err
		}

		for _, msg := range messages {
			if err = p.repo.DeleteOutboxMessage(ctx, tx, msg.ID); err != nil {
				return fmt.Errorf("delete outbox message %s: %w", msg.ID, err)
			}
		}

		count = len(messages)
		return nil
	})
	if err != nil {
		return 0, err
	}

	if count > 0 {
		p.logger.Debug("published offer report penalty messages to kafka", slog.Int("count", count))
	}

	return count, nil
}
