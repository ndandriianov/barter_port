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
	"barter-port/internal/libs/platform/logger"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"log"
	"net/http"
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
	_ = godotenv.Load()

	//serviceName := bootstrap.GetEnv("SERVICE_NAME", "auth")
	serviceConfigPath := "" //fmt.Sprintf("./config/%s.yaml", serviceName)

	cfg, err := bootstrap.LoadConfig(bootstrap.ConfigOptions{
		CommonPath:  os.Getenv("CONFIG_COMMON"),
		ServicePath: serviceConfigPath,
		AppEnv:      os.Getenv("APP_ENV"),
	})
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer db.Close()

	frontendURL := cfg.Frontend.URL
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	userRepo := user.NewRepository()
	emailTokenRepo := email_token.NewRepository()
	refreshTokenRepo := refresh_token.NewRepository()
	outboxRepo := &outbox.Repository{}

	m := bootstrap.InitMailerFromConfig(cfg)
	if err = bootstrap.ValidateMailConfig(cfg); err != nil {
		log.Fatal("failed to initialize mailer:", err)
	}

	logg := logger.NewJSONLogger(slog.LevelDebug, "auth-service", "")
	infrastructureLogger := logger.NewJSONLogger(slog.LevelDebug, "", "infrastructure")

	jwtManager, err := bootstrap.InitJWTManagerFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize JWT manager:", err)
	}

	validator, err := bootstrap.InitLocalJWTFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize JWT validator:", err)
	}
	if len(cfg.Kafka.Brokers) == 0 {
		log.Fatal("failed to initialize kafka writer: kafka brokers are not configured")
	}
	if cfg.Kafka.UserCreationTopic == "" {
		log.Fatal("failed to initialize kafka writer: user creation topic is not configured")
	}

	kafkaWriter := authkafka.NewWriter(cfg.Kafka.Brokers, cfg.Kafka.UserCreationTopic)
	topicInitCtx, cancelTopicInit := context.WithTimeout(context.Background(), cfg.Kafka.WriteTimeout)
	if err = authkafka.EnsureTopic(topicInitCtx, cfg.Kafka.Brokers, cfg.Kafka.UserCreationTopic, 1, 1); err != nil {
		cancelTopicInit()
		log.Fatal("failed to ensure kafka topic:", err)
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
	defer func() {
		if err := outboxPublisher.Close(); err != nil {
			logg.Error("failed to close kafka writer", slog.Any("error", err))
		}
	}()

	authService := application.NewService(db, userRepo, emailTokenRepo, m, infrastructureLogger, outboxRepo, frontendURL, re)
	handlers := transport.NewHandlers(logg, authService, jwtManager, db, refreshTokenRepo)
	router := transport.NewRouter(logg, validator, handlers)
	server := &http.Server{
		Addr:    bootstrap.InitPortStringFromConfig(cfg, 8081),
		Handler: router,
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publisherDone := make(chan struct{})
	go func() {
		defer close(publisherDone)
		if err := outboxPublisher.Run(rootCtx); err != nil {
			logg.Error("outbox publisher stopped with error", slog.Any("error", err))
		}
	}()

	go func() {
		<-rootCtx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logg.Error("failed to shutdown auth server", slog.Any("error", err))
		}
	}()

	log.Println("backend listening on", server.Addr)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	<-publisherDone
}
