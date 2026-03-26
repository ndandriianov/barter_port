package consumer

import (
	authusers "barter-port/internal/contracts/kafka/messages/auth-users"
	"barter-port/internal/libs/db"
	"barter-port/internal/libs/errorx"
	"barter-port/internal/users/infrastructure/repository/inbox"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type UserCreationInboxConsumer struct {
	log          *slog.Logger
	reader       messageReader
	db           db.DB
	inboxRepo    inboxWriter
	pollInterval time.Duration
}

type messageReader interface {
	FetchMessage(context.Context) (kafkago.Message, error)
	CommitMessages(context.Context, ...kafkago.Message) error
	Close() error
}

type inboxWriter interface {
	WriteUserCreationMessage(context.Context, db.DB, authusers.UserCreationMessage) error
}

type Params struct {
	Log          *slog.Logger
	Reader       messageReader
	DB           db.DB
	InboxRepo    inboxWriter
	PollInterval time.Duration
}

func NewUserCreationInboxConsumer(params Params) *UserCreationInboxConsumer {
	if params.PollInterval <= 0 {
		params.PollInterval = time.Second * 5
	}

	return &UserCreationInboxConsumer{
		log:          params.Log,
		reader:       params.Reader,
		db:           params.DB,
		inboxRepo:    params.InboxRepo,
		pollInterval: params.PollInterval,
	}
}

func (c *UserCreationInboxConsumer) Run(ctx context.Context) error {
	defer func() {
		if err := c.reader.Close(); err != nil {
			c.log.Error("failed to close Kafka consumer", slog.Any("error", err))
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

func (c *UserCreationInboxConsumer) consumeMessage(ctx context.Context) error {
	rawMessage, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}

	var message authusers.UserCreationMessage
	err = json.Unmarshal(rawMessage.Value, &message)
	if err != nil {
		if commitErr := c.commitMessage(ctx, rawMessage); commitErr != nil {
			return fmt.Errorf("failed to unmarshal message id=%s: %w; additionally: %w",
				string(rawMessage.Key), err, commitErr,
			)
		}
		return fmt.Errorf("failed to unmarshal message id: %s: %w", string(rawMessage.Key), err)
	}

	err = c.inboxRepo.WriteUserCreationMessage(ctx, c.db, message)
	if err != nil {
		if errors.Is(err, inbox.ErrUCEventAlreadyExists) {
			return c.commitMessage(ctx, rawMessage)
		}
		return fmt.Errorf("failed to write user creation message to inbox: %w", err)
	}

	return c.commitMessage(ctx, rawMessage)
}

func (c *UserCreationInboxConsumer) commitMessage(ctx context.Context, rawMessage kafkago.Message) error {
	if commitErr := c.reader.CommitMessages(ctx, rawMessage); commitErr != nil {
		return fmt.Errorf("commit message id: %s: %w", string(rawMessage.Key), commitErr)
	}
	return nil
}
