package producer

import (
	usersauth "barter-port/internal/contracts/kafka/messages/users-auth"
	"barter-port/internal/users/infrastructure/repository/outbox"
	"barter-port/pkg/db"
	kafkax2 "barter-port/pkg/kafkax"
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UCResultOutbox struct {
	db        *pgxpool.Pool
	repo      *outbox.Repository
	logger    *slog.Logger
	publisher *kafkax2.OutboxPublisher
}

func NewUCResultOutbox(
	db *pgxpool.Pool,
	repo *outbox.Repository,
	logger *slog.Logger,
	publisher *kafkax2.OutboxPublisher,
) *UCResultOutbox {
	return &UCResultOutbox{
		db:        db,
		repo:      repo,
		logger:    logger,
		publisher: publisher,
	}
}

func (p *UCResultOutbox) Run(ctx context.Context) error {
	return p.publisher.Run(ctx, p.publishBatch, "failed to publish user creation outbox batch")
}

func (p *UCResultOutbox) Close() error {
	return p.publisher.Close()
}

func (p *UCResultOutbox) publishBatch(ctx context.Context) (int, error) {
	var messages []usersauth.UCResultMessage

	err := db.RunInTx(ctx, p.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		messages, err = p.repo.ReadUCResultMessagesForUpdate(ctx, tx, p.publisher.BatchSize())
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return nil
		}

		kafkaMessages, err := kafkax2.BuildMessages(messages)
		if err != nil {
			return fmt.Errorf("build Kafka messages: %w", err)
		}

		p.logger.Debug("writing outbox uc_result messages to kafka", slog.Any("messages", messages))

		if err = p.publisher.WriteMessages(ctx, kafkaMessages, "ensure topic exists"); err != nil {
			return err
		}

		for _, message := range messages {
			if err = p.repo.DeleteUCResultMessage(ctx, tx, message.ID); err != nil {
				return fmt.Errorf("delete outbox message %s: %w", message.ID, err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}
	p.logger.Debug("published outbox uc_result messages to kafka", slog.Int("count", len(messages)))

	return len(messages), nil
}
