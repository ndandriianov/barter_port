package main

import (
	"barter-port/internal/auth/application"
	authkafka "barter-port/internal/auth/infrastructure/kafka"
	"barter-port/internal/auth/infrastructure/repository/email_token"
	"barter-port/internal/auth/infrastructure/repository/outbox"
	"barter-port/internal/auth/infrastructure/repository/refresh_token"
	"barter-port/internal/auth/infrastructure/repository/user"
	"barter-port/internal/auth/infrastructure/transport"
	"barter-port/internal/libs/bootstrap"
	"barter-port/internal/libs/kafkax"
	"barter-port/internal/libs/platform/logger"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

//go:generate bash ../../scripts/generate-swagger-auth.sh

// @title Barter Port API
// @version 1.0.0
// @description API for Barter Port
// @host localhost:80
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type app struct {
	logger          *slog.Logger
	db              *pgxpool.Pool
	server          *http.Server
	outboxPublisher *authkafka.UserCreationOutboxPublisher
}

func run() error {
	_ = godotenv.Load()

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	app, err := newApp(cfg)
	if err != nil {
		return err
	}
	defer app.Close()

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return app.Run(rootCtx)
}

func loadConfig() (bootstrap.Config, error) {
	//serviceName := bootstrap.GetEnv("SERVICE_NAME", "auth")
	serviceConfigPath := "" //fmt.Sprintf("./config/%s.yaml", serviceName)

	cfg, err := bootstrap.LoadConfig(bootstrap.ConfigOptions{
		CommonPath:  os.Getenv("CONFIG_COMMON"),
		ServicePath: serviceConfigPath,
		AppEnv:      os.Getenv("APP_ENV"),
	})
	if err != nil {
		return bootstrap.Config{}, errors.New("failed to load config: " + err.Error())
	}

	return cfg, nil
}

func newApp(cfg bootstrap.Config) (*app, error) {
	db, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		return nil, errors.New("failed to initialize database: " + err.Error())
	}

	frontendURL := cfg.Frontend.URL
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	userRepo := user.NewRepository()
	emailTokenRepo := email_token.NewRepository()
	refreshTokenRepo := refresh_token.NewRepository()
	outboxRepo := &outbox.Repository{}

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
	topicInitCtx, cancelTopicInit := context.WithTimeout(context.Background(), cfg.Kafka.WriteTimeout)
	if err = kafkax.EnsureTopic(topicInitCtx, cfg.Kafka.Brokers, cfg.Kafka.UserCreationTopic, 1, 1); err != nil {
		cancelTopicInit()
		db.Close()
		return nil, errors.New("failed to ensure kafka topic: " + err.Error())
	}
	cancelTopicInit()

	outboxPublisher := authkafka.NewUserCreationOutboxPublisher(
		db,
		outboxRepo,
		kafkaWriter,
		infrastructureLogger,
		cfg.Kafka.Brokers,
		cfg.Kafka.UserCreationTopic,
		cfg.Kafka.BatchSize,
		cfg.Kafka.PollInterval,
		cfg.Kafka.WriteTimeout,
	)

	authService := application.NewService(db, userRepo, emailTokenRepo, m, infrastructureLogger, outboxRepo, frontendURL, re)
	handlers := transport.NewHandlers(logg, authService, jwtManager, db, refreshTokenRepo)
	router := transport.NewRouter(logg, validator, handlers)
	server := &http.Server{
		Addr:    bootstrap.InitPortStringFromConfig(cfg, 8081),
		Handler: router,
	}

	return &app{
		logger:          logg,
		db:              db,
		server:          server,
		outboxPublisher: outboxPublisher,
	}, nil
}

func (a *app) Run(rootCtx context.Context) error {
	publisherDone := make(chan struct{})
	go func() {
		defer close(publisherDone)
		if err := a.outboxPublisher.Run(rootCtx); err != nil {
			a.logger.Error("outbox publisher stopped with error", slog.Any("error", err))
		}
	}()

	go func() {
		<-rootCtx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("failed to shutdown auth server", slog.Any("error", err))
		}
	}()

	log.Println("backend listening on", a.server.Addr)
	err := a.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-publisherDone
	return nil
}

func (a *app) Close() {
	if err := a.outboxPublisher.Close(); err != nil {
		a.logger.Error("failed to close kafka writer", slog.Any("error", err))
	}
	a.db.Close()
}
