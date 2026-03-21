package producer

import (
	"barter-port/internal/auth/infrastructure/repository/outbox"
	"barter-port/internal/contracts/kafka/messages/auth-users"
	"barter-port/internal/libs/db"
	"barter-port/internal/libs/errorx"
	"barter-port/internal/libs/kafkax"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
)

const userCreationEventType = "auth.user.created"

type UserCreationOutboxPublisher struct {
	db           *pgxpool.Pool
	repo         *outbox.Repository
	writer       kafkax.MessageWriter
	logger       *slog.Logger
	brokers      []string
	topic        string
	batchSize    int
	pollInterval time.Duration
	writeTimeout time.Duration
}

func NewUserCreationOutboxPublisher(
	dbPool *pgxpool.Pool,
	repo *outbox.Repository,
	writer kafkax.MessageWriter,
	logger *slog.Logger,
	brokers []string,
	topic string,
	batchSize int,
	pollInterval time.Duration,
	writeTimeout time.Duration,
) *UserCreationOutboxPublisher {
	if batchSize <= 0 {
		batchSize = 100
	}
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	if writeTimeout <= 0 {
		writeTimeout = 10 * time.Second
	}

	return &UserCreationOutboxPublisher{
		db:           dbPool,
		repo:         repo,
		writer:       writer,
		logger:       logger,
		brokers:      append([]string(nil), brokers...),
		topic:        topic,
		batchSize:    batchSize,
		pollInterval: pollInterval,
		writeTimeout: writeTimeout,
	}
}

func (p *UserCreationOutboxPublisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		published, err := p.publishBatch(ctx)
		if err != nil {
			if errorx.IsShutdownError(ctx, err) {
				return nil
			}

			p.logger.Error("failed to publish user creation outbox batch", slog.Any("error", err))
		}

		if published == p.batchSize {
			continue
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (p *UserCreationOutboxPublisher) Close() error {
	return p.writer.Close()
}

func (p *UserCreationOutboxPublisher) publishBatch(ctx context.Context) (int, error) {
	var messages []auth_users.UserCreationMessage

	err := db.RunInTx(ctx, p.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		messages, err = p.repo.ReadUserCreationMessagesForUpdate(ctx, tx, p.batchSize)
		if err != nil {
			return fmt.Errorf("read outbox messages: %w", err)
		}
		if len(messages) == 0 {
			return nil
		}

		kafkaMessages, err := buildMessages(messages)
		if err != nil {
			return fmt.Errorf("build kafka messages: %w", err)
		}

		writeCtx, cancel := context.WithTimeout(ctx, p.writeTimeout)
		defer cancel()

		p.logger.Debug("writing outbox messages to kafka", slog.Any("messages", messages))

		if err = p.writer.WriteMessages(writeCtx, kafkaMessages...); err != nil {
			if kafkax.IsUnknownTopicOrPartition(err) {
				if ensureErr := kafkax.EnsureTopic(writeCtx, p.brokers, p.topic, 1, 1); ensureErr != nil {
					return fmt.Errorf("ensure topic %q after publish failure: %w", p.topic, ensureErr)
				}
			}
			return fmt.Errorf("write kafka messages: %w", err)
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

	if len(messages) > 0 {
		p.logger.Info("published user creation messages", slog.Int("count", len(messages)))
	}

	return len(messages), nil
}

func buildMessages(messages []auth_users.UserCreationMessage) ([]kafkago.Message, error) {
	kafkaMessages := make([]kafkago.Message, 0, len(messages))

	for _, message := range messages {
		payload, err := json.Marshal(message)
		if err != nil {
			return nil, fmt.Errorf("marshal message %s: %w", message.ID, err)
		}

		kafkaMessages = append(kafkaMessages, kafkago.Message{
			Key:   []byte(message.UserID.String()),
			Value: payload,
			Time:  message.CreatedAt,
			Headers: []kafkago.Header{
				{Key: "message_id", Value: []byte(message.ID.String())},
				{Key: "message_type", Value: []byte(userCreationEventType)},
			},
		})
	}

	return kafkaMessages, nil
}
