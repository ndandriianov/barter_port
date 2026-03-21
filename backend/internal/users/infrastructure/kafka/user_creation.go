package kafka

import (
	authusers "barter-port/internal/contracts/kafka/auth-users"
	"barter-port/internal/users/infrastructure/repository/inbox"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
)

type UserCreationInboxConsumer struct {
	log          *slog.Logger
	reader       *kafkago.Reader
	db           *pgxpool.Pool
	inboxRepo    *inbox.Repository
	pollInterval time.Duration
}

func (c *UserCreationInboxConsumer) Run(ctx context.Context) error {
	defer func() {
		if err := c.reader.Close(); err != nil {
			c.log.Error("failed to close Kafka reader", slog.Any("error", err))
		}
	}()

	for {
		if ctx.Err() != nil {
			return nil
		}

		err := c.consumeMessage(ctx)
		if err == nil {
			continue
		}

		if isShutdownError(ctx, err) {
			return nil
		}

		c.log.Error("failed to consume message", "error", err)

		select {
		case <-time.After(c.pollInterval):
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *UserCreationInboxConsumer) consumeMessage(ctx context.Context) error {
	message, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}

	var event authusers.UserCreationEvent
	err = json.Unmarshal(message.Value, &event)
	if err != nil {
		if commitErr := c.reader.CommitMessages(ctx, message); commitErr != nil {
			return fmt.Errorf(
				"failed to unmarshal message id=%s: %w; additionally failed to commit bad message: %w",
				string(message.Key), err, commitErr,
			)
		}
		return fmt.Errorf("failed to unmarshal message id: %s: %w", string(message.Key), err)
	}

	err = c.inboxRepo.WriteUserCreationEvent(ctx, c.db, event)
	// TODO: проверка и skip при unique violation
	if err != nil {
		return fmt.Errorf("failed to write user creation event to inbox: %w", err)
	}

	err = c.reader.CommitMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to commit message id: %s: %w", string(message.Key), err)
	}

	return nil
}

func isShutdownError(ctx context.Context, err error) bool {
	return ctx.Err() != nil ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}
