package producer

import (
	"barter-port/internal/auth/infrastructure/repository/outbox"
	"barter-port/internal/contracts/kafka/messages/auth-users"
	"barter-port/pkg/db"
	kafkax2 "barter-port/pkg/kafkax"
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserCreationOutboxPublisher struct {
	db        *pgxpool.Pool
	repo      *outbox.Repository
	logger    *slog.Logger
	publisher *kafkax2.OutboxPublisher
}

func NewUserCreationOutboxPublisher(
	dbPool *pgxpool.Pool,
	repo *outbox.Repository,
	logger *slog.Logger,
	publisher *kafkax2.OutboxPublisher,
) *UserCreationOutboxPublisher {
	return &UserCreationOutboxPublisher{
		db:        dbPool,
		repo:      repo,
		logger:    logger,
		publisher: publisher,
	}
}

func (p *UserCreationOutboxPublisher) Run(ctx context.Context) error {
	return p.publisher.Run(ctx, p.publishBatch, "failed to publish user creation outbox batch")
}

func (p *UserCreationOutboxPublisher) Close() error {
	return p.publisher.Close()
}

func (p *UserCreationOutboxPublisher) publishBatch(ctx context.Context) (int, error) {
	var messages []auth_users.UserCreationMessage

	err := db.RunInTx(ctx, p.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		messages, err = p.repo.ReadUserCreationMessagesForUpdate(ctx, tx, p.publisher.BatchSize())
		if err != nil {
			return fmt.Errorf("read outbox messages: %w", err)
		}
		if len(messages) == 0 {
			return nil
		}

		kafkaMessages, err := kafkax2.BuildMessages(messages)
		if err != nil {
			return fmt.Errorf("build kafka messages: %w", err)
		}

		p.logger.Debug("writing outbox uc messages to kafka", slog.Any("messages", messages))

		if err = p.publisher.WriteMessages(ctx, kafkaMessages, "ensure topic after publish failure"); err != nil {
			return err
		}

		for _, message := range messages {
			if err = p.repo.DeleteUserCreationMessage(ctx, tx, message.ID); err != nil {
				return fmt.Errorf("delete outbox message %s: %w", message.ID, err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}
	p.logger.Info("published user creation messages", slog.Int("count", len(messages)))

	return len(messages), nil
}
