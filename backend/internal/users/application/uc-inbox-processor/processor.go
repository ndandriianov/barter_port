package uc_inbox_processor

import (
	authusers "barter-port/contracts/kafka/messages/auth-users"
	usersauth "barter-port/contracts/kafka/messages/users-auth"
	statusUpdate "barter-port/contracts/kafka/messages/users-auth/status-update"
	ucinbox "barter-port/internal/users/infrastructure/repository/uc-inbox"
	ucroutbox "barter-port/internal/users/infrastructure/repository/uc-result-outbox"
	"barter-port/internal/users/infrastructure/repository/user"
	"barter-port/internal/users/model"
	"barter-port/pkg/db"
	"barter-port/pkg/errorx"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Processor struct {
	inboxRepo  *ucinbox.Repository
	outboxRepo *ucroutbox.Repository
	userRepo   *user.Repository
	db         *pgxpool.Pool
	log        *slog.Logger

	batchSize    int
	pollInterval time.Duration
}

func NewProcessor(
	inboxRepo *ucinbox.Repository,
	outboxRepo *ucroutbox.Repository,
	userRepo *user.Repository,
	db *pgxpool.Pool,
	log *slog.Logger,
	batchSize int,
	pollInterval time.Duration,
) *Processor {
	return &Processor{
		inboxRepo:    inboxRepo,
		outboxRepo:   outboxRepo,
		userRepo:     userRepo,
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

			p.log.Error("failed to process user creation messages", slog.Any("error", err))
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
	var messages []authusers.UserCreationMessage

	err := db.RunInTx(ctx, p.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		messages, err = p.inboxRepo.ReadUserCreationMessagesForUpdate(ctx, tx, p.batchSize)
		if err != nil {
			return fmt.Errorf("failed to read user creation messages: %w", err)
		}
		if len(messages) == 0 {
			return nil
		}

		for _, message := range messages {
			err = p.userRepo.AddUser(ctx, tx, message.UserID)
			if err != nil {
				if errors.Is(err, model.ErrUserAlreadyExists) {
					err = p.outboxRepo.WriteUCResultMessage(ctx, tx, usersauth.UCResultMessage{
						ID:        uuid.New(),
						UserID:    message.UserID,
						Status:    statusUpdate.Failed.String(),
						CreatedAt: time.Now(),
					})
					if err != nil {
						return fmt.Errorf("failed to write user creation result: %w", err)
					}
				}

				return fmt.Errorf("failed to add user: %w", err)
			}

			// TODO: отправить событие об успехе
			err = p.outboxRepo.WriteUCResultMessage(ctx, tx, usersauth.UCResultMessage{
				ID:        uuid.New(),
				UserID:    message.UserID,
				Status:    statusUpdate.Success.String(),
				CreatedAt: time.Now(),
			})
			if err != nil {
				return fmt.Errorf("failed to write user creation result: %w", err)
			}

			err = p.inboxRepo.DeleteUserCreationMessage(ctx, tx, message.ID)
			if err != nil {
				return fmt.Errorf("failed to delete user creation message: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	if len(messages) > 0 {
		p.log.Debug("processed user creation messages", slog.Int("count", len(messages)))
	}

	return len(messages), nil
}
