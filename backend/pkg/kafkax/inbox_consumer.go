package kafkax

import (
	"barter-port/pkg/errorx"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type InboxConsumer[T any] struct {
	log          *slog.Logger
	reader       messageReader
	pollInterval time.Duration
}

type messageReader interface {
	FetchMessage(context.Context) (kafkago.Message, error)
	CommitMessages(context.Context, ...kafkago.Message) error
	Close() error
}

var ErrDuplicate = errors.New("duplicate message")

func NewInboxConsumer[T any](log *slog.Logger, reader messageReader, pollInterval time.Duration) *InboxConsumer[T] {
	return &InboxConsumer[T]{
		log:          log,
		reader:       reader,
		pollInterval: pollInterval,
	}
}

func (c *InboxConsumer[T]) Run(ctx context.Context, processMessage func(context.Context, T) error) error {
	defer func() {
		if err := c.reader.Close(); err != nil {
			c.log.Error("failed to close Kafka consumer", slog.Any("error", err))
		}
	}()

	for {
		if ctx.Err() != nil {
			return nil
		}

		err := c.consumeMessage(ctx, processMessage)
		if err == nil {
			continue
		}

		if errorx.IsShutdownError(ctx, err) {
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

func (c *InboxConsumer[T]) consumeMessage(ctx context.Context, processMessage func(context.Context, T) error) error {
	rawMessage, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}

	var message T
	err = json.Unmarshal(rawMessage.Value, &message)
	if err != nil {
		if commitErr := c.CommitMessage(ctx, rawMessage); commitErr != nil {
			return fmt.Errorf("failed to unmarshal message id=%s: %w; additionally: %w",
				string(rawMessage.Key), err, commitErr,
			)
		}
		return fmt.Errorf("failed to unmarshal message id: %s: %w", string(rawMessage.Key), err)
	}

	err = processMessage(ctx, message)
	if err != nil {
		if errors.Is(err, ErrDuplicate) {
			return c.CommitMessage(ctx, rawMessage)
		}
		return fmt.Errorf("failed to process message: %w", err)
	}

	return c.CommitMessage(ctx, rawMessage)
}

func (c *InboxConsumer[T]) CommitMessage(ctx context.Context, rawMessage kafkago.Message) error {
	if commitErr := c.reader.CommitMessages(ctx, rawMessage); commitErr != nil {
		return fmt.Errorf("commit message id: %s: %w", string(rawMessage.Key), commitErr)
	}
	return nil
}
