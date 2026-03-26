package app

import (
	"barter-port/internal/auth/application"
	authkafka "barter-port/internal/auth/infrastructure/kafka/producer"
	"barter-port/internal/auth/infrastructure/repository/email_token"
	"barter-port/internal/auth/infrastructure/repository/refresh_token"
	ucoutbox "barter-port/internal/auth/infrastructure/repository/uc-outbox"
	"barter-port/internal/auth/infrastructure/repository/user"
	"barter-port/internal/auth/infrastructure/transport"
	"barter-port/pkg/bootstrap"
	"barter-port/pkg/kafkax"
	"barter-port/pkg/logger"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/run"
)

type App struct {
	logger          *slog.Logger
	db              *pgxpool.Pool
	server          *http.Server
	outboxPublisher *authkafka.UserCreationOutboxPublisher
}

func NewApp(cfg bootstrap.Config) (*App, error) {
	db, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		return nil, errors.New("failed to initialize database: " + err.Error())
	}

	frontendURL := cfg.Frontend.URL
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	userRepo := user.NewRepository()
	emailTokenRepo := email_token.NewRepository()
	refreshTokenRepo := refresh_token.NewRepository()
	outboxRepo := &ucoutbox.Repository{}

	m := bootstrap.InitMailerFromConfig(cfg)
	if err = bootstrap.ValidateMailConfig(cfg); err != nil {
		db.Close()
		return nil, errors.New("failed to initialize mailer: " + err.Error())
	}

	logg := logger.NewJSONLogger(slog.LevelDebug, "auth-service", "")
	infrastructureLogger := logger.NewJSONLogger(slog.LevelDebug, "", "infrastructure")

	jwtManager, err := bootstrap.InitJWTManagerFromConfig(cfg)
	if err != nil {
		db.Close()
		return nil, errors.New("failed to initialize JWT manager: " + err.Error())
	}

	validator, err := bootstrap.InitLocalJWTFromConfig(cfg)
	if err != nil {
		db.Close()
		return nil, errors.New("failed to initialize JWT validator: " + err.Error())
	}
	if len(cfg.Kafka.Brokers) == 0 {
		db.Close()
		return nil, errors.New("failed to initialize kafka writer: kafka brokers are not configured")
	}
	if cfg.Kafka.UserCreationTopic == "" {
		db.Close()
		return nil, errors.New("failed to initialize kafka writer: user creation topic is not configured")
	}

	kafkaWriter := kafkax.NewWriter(cfg.Kafka.Brokers, cfg.Kafka.UserCreationTopic)
	writerOwned := false
	defer func() {
		if !writerOwned {
			_ = kafkaWriter.Close()
		}
	}()

	topicInitCtx, cancelTopicInit := context.WithTimeout(context.Background(), cfg.Kafka.WriteTimeout)
	if err = kafkax.EnsureTopic(topicInitCtx, cfg.Kafka.Brokers, cfg.Kafka.UserCreationTopic, 1, 1); err != nil {
		cancelTopicInit()
		db.Close()
		return nil, errors.New("failed to ensure kafka topic: " + err.Error())
	}
	cancelTopicInit()

	kafkaPublisher := kafkax.NewOutboxPublisher(
		kafkaWriter,
		infrastructureLogger,
		cfg.Kafka.Brokers,
		cfg.Kafka.UserCreationTopic,
		cfg.Kafka.BatchSize,
		cfg.Kafka.PollInterval,
		cfg.Kafka.WriteTimeout,
	)

	outboxPublisher := authkafka.NewUserCreationOutboxPublisher(db, outboxRepo, infrastructureLogger, kafkaPublisher)

	authService := application.NewService(db, userRepo, emailTokenRepo, m, infrastructureLogger, outboxRepo, frontendURL, re)
	handlers := transport.NewHandlers(logg, authService, jwtManager, db, refreshTokenRepo)
	router := transport.NewRouter(logg, validator, handlers)
	server := &http.Server{
		Addr:    bootstrap.InitPortStringFromConfig(cfg, 8081),
		Handler: router,
	}

	writerOwned = true

	return &App{
		logger:          logg,
		db:              db,
		server:          server,
		outboxPublisher: outboxPublisher,
	}, nil
}

func (a *App) Run() error {
	var g run.Group
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g.Add(func() error {
		return a.outboxPublisher.Run(ctx)
	}, func(error) {
		cancel()
	})

	g.Add(func() error {
		log.Println("backend listening on", a.server.Addr)
		return a.server.ListenAndServe()
	}, func(error) {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("failed to shutdown auth server", slog.Any("error", err))
		}
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	g.Add(func() error {
		<-stop
		return nil
	}, func(error) {
		signal.Stop(stop)
	})

	err := g.Run()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (a *App) Close() {
	if err := a.outboxPublisher.Close(); err != nil {
		a.logger.Error("failed to close kafka writer", slog.Any("error", err))
	}
	a.db.Close()
}
