package reputation_processor

import (
	reputation_events "barter-port/internal/users/infrastructure/repository/reputation-events"
	reputation_inbox "barter-port/internal/users/infrastructure/repository/reputation-inbox"
	"barter-port/pkg/db"
	"barter-port/pkg/errorx"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Processor struct {
	inboxRepo  *reputation_inbox.Repository
	eventsRepo *reputation_events.Repository
	db         *pgxpool.Pool
	log        *slog.Logger

	batchSize    int
	pollInterval time.Duration
}

func NewProcessor(
	inboxRepo *reputation_inbox.Repository,
	eventsRepo *reputation_events.Repository,
	dbPool *pgxpool.Pool,
	log *slog.Logger,
	batchSize int,
	pollInterval time.Duration,
) *Processor {
	return &Processor{
		inboxRepo:    inboxRepo,
		eventsRepo:   eventsRepo,
		db:           dbPool,
		log:          log,
		batchSize:    batchSize,
		pollInterval: pollInterval,
	}
}

func (p *Processor) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		processed, err := p.processNext(ctx)
		if err != nil {
			if errorx.IsShutdownError(ctx, err) {
				return nil
			}
			p.log.Error("failed to process reputation messages", slog.Any("error", err))
		}

		if processed == p.batchSize {
			continue
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return nil
		}
	}
}

func (p *Processor) processNext(ctx context.Context) (int, error) {
	var messages []reputation_inbox.InboxMessage

	err := db.RunInTx(ctx, p.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		messages, err = p.inboxRepo.ReadMessagesForUpdate(ctx, tx, p.batchSize)
		if err != nil {
			return fmt.Errorf("read reputation inbox: %w", err)
		}
		if len(messages) == 0 {
			return nil
		}

		for _, msg := range messages {
			inserted, err := p.eventsRepo.WriteReputationEvent(ctx, tx, msg)
			if err != nil {
				return fmt.Errorf("write reputation event: %w", err)
			}

			if inserted {
				if err = p.eventsRepo.ApplyReputationDelta(ctx, tx, msg.UserID, msg.Delta); err != nil {
					return fmt.Errorf("apply reputation delta: %w", err)
				}
			}

			if err = p.inboxRepo.DeleteMessage(ctx, tx, msg.ID); err != nil {
				return fmt.Errorf("delete reputation inbox message: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	if len(messages) > 0 {
		p.log.Debug("processed reputation messages", slog.Int("count", len(messages)))
	}

	return len(messages), nil
}
