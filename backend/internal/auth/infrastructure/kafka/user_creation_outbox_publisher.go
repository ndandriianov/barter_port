package kafka

import (
	"barter-port/internal/auth/domain"
	"barter-port/internal/auth/infrastructure/repository/outbox"
	"barter-port/internal/libs/db"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
)

const userCreationEventType = "auth.user.created"

type messageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafkago.Message) error
	Close() error
}

type UserCreationOutboxPublisher struct {
	db           *pgxpool.Pool
	repo         *outbox.Repository
	writer       messageWriter
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
	writer messageWriter,
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

func NewWriter(brokers []string, topic string) *kafkago.Writer {
	return &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequireOne,
		Async:        false,
	}
}

func (p *UserCreationOutboxPublisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		published, err := p.publishBatch(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
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
	var events []domain.UserCreationEvent

	err := db.RunInTx(ctx, p.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		events, err = p.repo.ReadUserCreationEventsForUpdate(ctx, tx, p.batchSize)
		if err != nil {
			return fmt.Errorf("read outbox events: %w", err)
		}
		if len(events) == 0 {
			return nil
		}

		messages, err := buildMessages(events)
		if err != nil {
			return fmt.Errorf("build kafka messages: %w", err)
		}

		writeCtx, cancel := context.WithTimeout(ctx, p.writeTimeout)
		defer cancel()

		p.logger.Debug("writing outbox events to kafka", slog.Any("events", events))

		if err = p.writer.WriteMessages(writeCtx, messages...); err != nil {
			if isUnknownTopicOrPartition(err) {
				if ensureErr := EnsureTopic(writeCtx, p.brokers, p.topic, 1, 1); ensureErr != nil {
					return fmt.Errorf("ensure topic %q after publish failure: %w", p.topic, ensureErr)
				}
			}
			return fmt.Errorf("write kafka messages: %w", err)
		}

		for _, event := range events {
			if err = p.repo.DeleteUserCreationEvent(ctx, tx, event.ID); err != nil {
				return fmt.Errorf("delete outbox event %s: %w", event.ID, err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	if len(events) > 0 {
		p.logger.Info("published user creation events", slog.Int("count", len(events)))
	}

	return len(events), nil
}

func isUnknownTopicOrPartition(err error) bool {
	return errors.Is(err, kafkago.UnknownTopicOrPartition) ||
		strings.Contains(err.Error(), "Unknown Topic Or Partition")
}

func buildMessages(events []domain.UserCreationEvent) ([]kafkago.Message, error) {
	messages := make([]kafkago.Message, 0, len(events))

	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return nil, fmt.Errorf("marshal event %s: %w", event.ID, err)
		}

		messages = append(messages, kafkago.Message{
			Key:   []byte(event.UserID.String()),
			Value: payload,
			Time:  event.CreatedAt,
			Headers: []kafkago.Header{
				{Key: "event_id", Value: []byte(event.ID.String())},
				{Key: "event_type", Value: []byte(userCreationEventType)},
			},
		})
	}

	return messages, nil
}
