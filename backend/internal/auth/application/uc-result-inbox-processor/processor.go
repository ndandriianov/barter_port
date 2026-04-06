package uc_result_inbox_processor

import (
	usersauth "barter-port/contracts/kafka/messages/users-auth"
	statusUpdate "barter-port/contracts/kafka/messages/users-auth/status-update"
	ucevent "barter-port/internal/auth/infrastructure/repository/uc-event"
	ucrinbox "barter-port/internal/auth/infrastructure/repository/uc-result-inbox"
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
	inboxRepo *ucrinbox.Repository
	eventRepo *ucevent.Repository
	db        *pgxpool.Pool
	log       *slog.Logger

	batchSize    int
	pollInterval time.Duration
}

func NewProcessor(
	inboxRepo *ucrinbox.Repository,
	eventRepo *ucevent.Repository,
	db *pgxpool.Pool,
	log *slog.Logger,
	batchSize int,
	pollInterval time.Duration,
) *Processor {
	return &Processor{
		inboxRepo:    inboxRepo,
		eventRepo:    eventRepo,
		db:           db,
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

			p.log.Error("failed to process uc-result messages", slog.Any("error", err))
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
	var messages []usersauth.UCResultMessage

	err := db.RunInTx(ctx, p.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		messages, err = p.inboxRepo.ReadUCResultMessagesForUpdate(ctx, tx, p.batchSize)
		if err != nil {
			return fmt.Errorf("failed to read uc-result messages: %w", err)
		}
		if len(messages) == 0 {
			return nil
		}

		for _, message := range messages {
			status, err := statusUpdate.EnumString(message.Status)
			if err != nil {
				return fmt.Errorf("invalid uc-result status %q: %w", message.Status, err)
			}

			if err = p.eventRepo.SetStatus(ctx, tx, message.UserID, status.String()); err != nil {
				return fmt.Errorf("failed to update user creation status for user %s: %w", message.UserID, err)
			}

			if err = p.inboxRepo.DeleteUCResultMessage(ctx, tx, message.ID); err != nil {
				return fmt.Errorf("failed to delete uc-result message: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	if len(messages) > 0 {
		p.log.Debug("processed uc-result messages", slog.Int("count", len(messages)))
	}

	return len(messages), nil
}
