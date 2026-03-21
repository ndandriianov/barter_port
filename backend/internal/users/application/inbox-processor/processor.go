package inbox_processor

import (
	authusers "barter-port/internal/contracts/kafka/auth-users"
	"barter-port/internal/libs/db"
	"barter-port/internal/libs/errorx"
	"barter-port/internal/users/infrastructure/repository/inbox"
	"barter-port/internal/users/infrastructure/repository/user"
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

			p.log.Error("failed to process user creation events", slog.Any("error", err))
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
	var events []authusers.UserCreationEvent

	err := db.RunInTx(ctx, p.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		events, err = p.inboxRepo.ReadUserCreationEventsForUpdate(ctx, tx, p.batchSize)
		if err != nil {
			return fmt.Errorf("failed to read user creation events: %w", err)
		}
		if len(events) == 0 {
			return nil
		}

		for _, event := range events {
			err = p.userRepo.AddUser(ctx, tx, event.UserID)
			if err != nil {
				return fmt.Errorf("failed to add user: %w", err)
			}

			err = p.inboxRepo.DeleteUserCreationEvent(ctx, tx, event.ID)
			if err != nil {
				return fmt.Errorf("failed to delete user creation event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	if len(events) > 0 {
		p.log.Debug("processed %d user creation events", len(events))
	}

	return len(events), nil
}
