package app

import (
	"barter-port/internal/users/infrastructure/kafka/consumer"
	"barter-port/internal/users/infrastructure/repository/inbox"
	"barter-port/internal/users/infrastructure/repository/user"
	"barter-port/pkg/bootstrap"
	"barter-port/pkg/logger"
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
	inboxRepository *inbox.Repository
	inboxProcessor  *inboxP.Processor
	ucEventConsumer *consumer.UserCreationInboxConsumer
}

func NewApp(cfg bootstrap.Config) (*App, error) {
	app := &App{}

	if err := app.initDatabase(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	app.log = logger.NewJSONLogger(slog.LevelDebug, "users", "")

	app.inboxRepository = inbox.NewRepository()
	userRepo := user.NewRepository(app.db)

	app.inboxProcessor = inboxP.NewProcessor(inboxP.Params{
		InboxRepo:    app.inboxRepository,
		UserRepo:     userRepo,
		Db:           app.db,
		Log:          app.log,
		BatchSize:    cfg.Kafka.BatchSize,
		PollInterval: cfg.Kafka.PollInterval,
	})

	err := app.initUCEventConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize user creation event consumer: %w", err)
	}

	return app, nil
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
