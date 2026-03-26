package inbox_processor

import (
	authusers "barter-port/internal/contracts/kafka/messages/auth-users"
	"barter-port/internal/users/infrastructure/repository/inbox"
	"barter-port/internal/users/infrastructure/repository/user"
	"barter-port/internal/users/model"
	"barter-port/pkg/db"
	"barter-port/pkg/errorx"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Processor struct {
	inboxRepo *inbox.Repository
	userRepo  *user.Repository
	db        *pgxpool.Pool
	log       *slog.Logger

	batchSize    int
	pollInterval time.Duration
}

type Params struct {
	InboxRepo *inbox.Repository
	UserRepo  *user.Repository
	Db        *pgxpool.Pool
	Log       *slog.Logger

	BatchSize    int
	PollInterval time.Duration
}

func NewProcessor(params Params) *Processor {
	return &Processor{
		inboxRepo:    params.InboxRepo,
		userRepo:     params.UserRepo,
		db:           params.Db,
		log:          params.Log,
		batchSize:    params.BatchSize,
		pollInterval: params.PollInterval,
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
					// TODO: отправить событие об уже существующем пользователе
				}

				return fmt.Errorf("failed to add user: %w", err)
			}

			// TODO: отправить событие об успехе

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
		p.log.Debug("processed %d user creation messages", len(messages))
	}

	return len(messages), nil
}
