package app

import (
	userservice "barter-port/internal/users/application/user"
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
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ucinboxP "barter-port/internal/users/application/uc-inbox-processor"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/run"
	"google.golang.org/grpc"
)

type App struct {
	log              *slog.Logger
	db               *pgxpool.Pool
	authDB           *pgxpool.Pool
	authGRPCConn     *grpc.ClientConn
	server           *http.Server
	grpcServer       *grpc.Server
	grpcListener     net.Listener
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
	authClient, err := app.initAuthGRPCClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth grpc client: %w", err)
	}
	userService := userservice.NewService(userRepo, authClient)

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

	if err = app.initHTTPServer(cfg, userService); err != nil {
		return nil, fmt.Errorf("failed to initialize http server: %w", err)
	}

	err = app.initGRPCServer(cfg, userService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize grpc server: %w", err)
	}

	return app, nil
}

func (app *App) Close() error {
	var err error

	if app.ucrEventProducer != nil {
		err = errors.Join(err, app.ucrEventProducer.Close())
	}
	if app.authDB != nil {
		app.authDB.Close()
	}
	if app.authGRPCConn != nil {
		err = errors.Join(err, app.authGRPCConn.Close())
	}
	if app.db != nil {
		app.db.Close()
	}
	if app.grpcListener != nil {
		_ = app.grpcListener.Close()
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

	g.Add(func() error {
		app.log.Info("users http server listening", slog.String("addr", app.server.Addr))
		return app.server.ListenAndServe()
	}, func(error) {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := app.server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.log.Error("failed to shutdown users server", slog.Any("error", err))
		}
	})

	g.Add(func() error {
		if app.grpcListener == nil || app.grpcServer == nil {
			return errors.New("grpc server or listener is not initialized")
		}

		app.log.Info("users grpc server listening", slog.String("addr", app.grpcListener.Addr().String()))
		return app.grpcServer.Serve(app.grpcListener)
	}, func(error) {
		if app.grpcServer != nil {
			app.grpcServer.GracefulStop()
		}
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

	err := g.Run()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
