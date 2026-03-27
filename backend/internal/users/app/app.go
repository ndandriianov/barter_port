package app

import (
	"barter-port/internal/users/infrastructure/kafka/consumer"
	"barter-port/internal/users/infrastructure/kafka/producer"
	ucinbox "barter-port/internal/users/infrastructure/repository/uc-inbox"
	ucroutbox "barter-port/internal/users/infrastructure/repository/uc-result-outbox"
	"barter-port/internal/users/infrastructure/repository/user"
	"barter-port/pkg/bootstrap"
	"barter-port/pkg/logger"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	ucinboxP "barter-port/internal/users/application/uc-inbox-processor"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/run"
)

type App struct {
	log              *slog.Logger
	db               *pgxpool.Pool
	inboxRepository  *ucinbox.Repository
	inboxProcessor   *ucinboxP.Processor
	ucEventConsumer  *consumer.UserCreationInboxConsumer
	outboxRepository *ucroutbox.Repository
	ucrEventProducer *producer.UCResultOutbox
}

func NewApp(cfg bootstrap.Config) (*App, error) {
	app := &App{}
	var err error
	defer func() {
		if err != nil {
			_ = app.Close()
		}
	}()

	if err = app.initDatabase(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	app.log = logger.NewJSONLogger(slog.LevelDebug, "users", "")

	// Repositories
	app.inboxRepository = ucinbox.NewRepository()
	app.outboxRepository = ucroutbox.NewRepository()
	userRepo := user.NewRepository(app.db)

	// Kafka
	app.inboxProcessor = ucinboxP.NewProcessor(
		app.inboxRepository,
		app.outboxRepository,
		userRepo,
		app.db,
		app.log,
		cfg.Kafka.BatchSize,
		cfg.Kafka.PollInterval,
	)

	err = app.initUCEventConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize user creation event consumer: %w", err)
	}

	err = app.initUCREventProducer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize user creation result event producer: %w", err)
	}

	return app, nil
}

func (app *App) Close() error {
	var err error

	if app.ucrEventProducer != nil {
		err = errors.Join(err, app.ucrEventProducer.Close())
	}
	if app.db != nil {
		app.db.Close()
	}

	return err
}

func (app *App) Run() error {
	var g run.Group
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		if err := app.Close(); err != nil && app.log != nil {
			app.log.Error("failed to close app resources", slog.Any("error", err))
		}
	}()

	g.Add(func() error {
		return app.inboxProcessor.Run(ctx)
	}, func(error) {
		cancel()
	})

	g.Add(func() error {
		return app.ucEventConsumer.Run(ctx)
	}, func(error) {
		cancel()
	})

	g.Add(func() error {
		return app.ucrEventProducer.Run(ctx)
	}, func(err error) {
		cancel()
	})

	// группа, отвечающая за graceful shutdown при получении сигнала прерывания
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	g.Add(func() error {
		select {
		case <-stop:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}, func(error) {
		signal.Stop(stop)
	})

	return g.Run()
}
