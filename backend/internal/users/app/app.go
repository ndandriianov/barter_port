package app

import (
	"barter-port/internal/libs/bootstrap"
	"barter-port/internal/libs/kafkax"
	"barter-port/internal/libs/platform/logger"
	"barter-port/internal/users/infrastructure/kafka"
	"barter-port/internal/users/infrastructure/repository/inbox"
	"barter-port/internal/users/infrastructure/repository/user"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	inboxP "barter-port/internal/users/application/inbox-processor"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/run"
	"golang.org/x/net/context"
)

type App struct {
	log             *slog.Logger
	db              *pgxpool.Pool
	inboxProcessor  *inboxP.Processor
	ucEventConsumer *kafka.UserCreationInboxConsumer
}

func NewApp(cfg bootstrap.Config) (*App, error) {
	db, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	log := logger.NewJSONLogger(slog.LevelDebug, "users", "")

	inboxRepo := inbox.NewRepository()
	userRepo := user.NewRepository(db)

	inboxProcessor := inboxP.NewProcessor(inboxP.Params{
		InboxRepo:    inboxRepo,
		UserRepo:     userRepo,
		Db:           db,
		Log:          log,
		BatchSize:    cfg.Kafka.BatchSize,
		PollInterval: cfg.Kafka.PollInterval,
	})

	ucEventConsumer := kafka.NewUserCreationInboxConsumer(kafka.Params{
		Log:          log,
		Reader:       kafkax.NewMessageReader(cfg.Kafka.Brokers, cfg.Kafka.UserCreationTopic, cfg.Kafka.UserCreationGroup),
		DB:           db,
		InboxRepo:    inboxRepo,
		PollInterval: cfg.Kafka.PollInterval,
	})

	return &App{
		log:             log,
		db:              db,
		inboxProcessor:  inboxProcessor,
		ucEventConsumer: ucEventConsumer,
	}, nil
}

func (app *App) Run() error {
	var g run.Group
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// группа, отвечающая за graceful shutdown при получении сигнала прерывания
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	g.Add(func() error {
		<-stop
		return nil
	}, func(error) {
		signal.Stop(stop)
	})

	err := g.Run()
	if err != nil {
		return err
	}

	return nil
}
