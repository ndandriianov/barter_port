package producer

import (
	usersauth "barter-port/internal/contracts/kafka/messages/users-auth"
	"barter-port/internal/libs/db"
	"barter-port/internal/libs/errorx"
	"barter-port/internal/libs/kafkax"
	"barter-port/internal/users/infrastructure/repository/outbox"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
)

const ucResultMessageType = "users.auth.uc_result"

type UCResultOutbox struct {
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

func NewUCResultOutbox(
	db *pgxpool.Pool,
	repo *outbox.Repository,
	writer kafkax.MessageWriter,
	logger *slog.Logger,
	brokers []string,
	topic string,
	batchSize int,
	pollInterval time.Duration,
	writeTimeout time.Duration,
) *UCResultOutbox {
	return &UCResultOutbox{
		db:           db,
		repo:         repo,
		writer:       writer,
		logger:       logger,
		brokers:      brokers,
		topic:        topic,
		batchSize:    batchSize,
		pollInterval: pollInterval,
		writeTimeout: writeTimeout,
	}
}

func (p *UCResultOutbox) Run(ctx context.Context) error {
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

func (p *UCResultOutbox) Close() error {
	return p.writer.Close()
}

func (p *UCResultOutbox) publishBatch(ctx context.Context) (int, error) {
	var messages []usersauth.UCResultMessage

	err := db.RunInTx(ctx, p.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		messages, err = p.repo.ReadUCResultMessagesForUpdate(ctx, tx, p.batchSize)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return nil
		}

		kafkaMessages, err := buildMessages(messages)
		if err != nil {
			return fmt.Errorf("build Kafka messages: %w", err)
		}

		writeCtx, cancel := context.WithTimeout(ctx, p.writeTimeout)
		defer cancel()

		p.logger.Debug("writing outbox uc_result messages to kafka", slog.Any("messages", messages))

		if err = p.writer.WriteMessages(writeCtx, kafkaMessages...); err != nil {
			if kafkax.IsUnknownTopicOrPartition(err) {
				if ensureErr := kafkax.EnsureTopic(writeCtx, p.brokers, p.topic, 1, 1); ensureErr != nil {
					return fmt.Errorf("ensure topic exists: %w", ensureErr)
				}
				return fmt.Errorf("write messages to Kafka: %w", err)
			}
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

func buildMessages(messages []usersauth.UCResultMessage) ([]kafkago.Message, error) {
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
				{Key: "message_type", Value: []byte(ucResultMessageType)},
			},
		})
	}

	return kafkaMessages, nil
}
