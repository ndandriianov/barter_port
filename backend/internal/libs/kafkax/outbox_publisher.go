package kafkax

import (
	"barter-port/internal/libs/errorx"
	"context"
	"fmt"
	"log/slog"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type OutboxPublisher struct {
	writer       MessageWriter
	logger       *slog.Logger
	brokers      []string
	topic        string
	batchSize    int
	pollInterval time.Duration
	writeTimeout time.Duration
}

func NewOutboxPublisher(
	writer MessageWriter,
	logger *slog.Logger,
	brokers []string,
	topic string,
	batchSize int,
	pollInterval time.Duration,
	writeTimeout time.Duration,
) *OutboxPublisher {
	if batchSize <= 0 {
		batchSize = 100
	}
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	if writeTimeout <= 0 {
		writeTimeout = 10 * time.Second
	}

	return &OutboxPublisher{
		writer:       writer,
		logger:       logger,
		brokers:      append([]string(nil), brokers...),
		topic:        topic,
		batchSize:    batchSize,
		pollInterval: pollInterval,
		writeTimeout: writeTimeout,
	}
}

func (p *OutboxPublisher) BatchSize() int {
	return p.batchSize
}

func (p *OutboxPublisher) Run(ctx context.Context, publishBatch func(context.Context) (int, error), failureMessage string) error {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		published, err := publishBatch(ctx)
		if err != nil {
			if errorx.IsShutdownError(ctx, err) {
				return nil
			}

			p.logger.Error(failureMessage, slog.Any("error", err))
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

func (p *OutboxPublisher) Close() error {
	return p.writer.Close()
}

func (p *OutboxPublisher) WriteMessages(ctx context.Context, messages []kafkago.Message, ensureTopicMessage string) error {
	writeCtx, cancel := context.WithTimeout(ctx, p.writeTimeout)
	defer cancel()

	if err := p.writer.WriteMessages(writeCtx, messages...); err != nil {
		if IsUnknownTopicOrPartition(err) {
			if ensureErr := EnsureTopic(writeCtx, p.brokers, p.topic, 1, 1); ensureErr != nil {
				return fmt.Errorf("%s: %w", ensureTopicMessage, ensureErr)
			}
		}

		return fmt.Errorf("write kafka messages: %w", err)
	}

	return nil
}
