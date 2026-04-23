package app

import (
	authpb "barter-port/contracts/grpc/auth/v1"
	usersauth "barter-port/contracts/kafka/messages/users-auth"
	"barter-port/internal/auth/application"
	ucrprocessor "barter-port/internal/auth/application/uc-result-inbox-processor"
	authconsumer "barter-port/internal/auth/infrastructure/kafka/consumer"
	authkafka "barter-port/internal/auth/infrastructure/kafka/producer"
	"barter-port/internal/auth/infrastructure/repository/email_token"
	"barter-port/internal/auth/infrastructure/repository/password_reset_token"
	"barter-port/internal/auth/infrastructure/repository/refresh_token"
	ucevent "barter-port/internal/auth/infrastructure/repository/uc-event"
	ucoutbox "barter-port/internal/auth/infrastructure/repository/uc-outbox"
	ucrinbox "barter-port/internal/auth/infrastructure/repository/uc-result-inbox"
	"barter-port/internal/auth/infrastructure/repository/user"
	grpctransport "barter-port/internal/auth/infrastructure/transport/grpc"
	httptransport "barter-port/internal/auth/infrastructure/transport/http"
	"barter-port/pkg/authkit/validators"
	"barter-port/pkg/bootstrap"
	"barter-port/pkg/jwt"
	"barter-port/pkg/kafkax"
	"barter-port/pkg/logger"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"regexp"

	"google.golang.org/grpc"
)

func (a *App) initDatabase(cfg bootstrap.Config) error {
	db, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	a.db = db
	return nil
}

func (a *App) initServices(cfg bootstrap.Config) error {
	a.initLoggers()
	infraLogger := a.infrastructureLogger()

	mailer, err := a.initMailer(cfg)
	if err != nil {
		return err
	}

	jwtManager, validator, err := a.initJWT(cfg)
	if err != nil {
		return err
	}

	userRepo := user.NewRepository()
	ucEventRepo := ucevent.NewRepository()
	ucResultInboxRepo := ucrinbox.NewRepository()
	emailTokenRepo := email_token.NewRepository()
	passwordResetTokenRepo := password_reset_token.NewRepository()
	refreshTokenRepo := refresh_token.NewRepository()
	outboxRepo := &ucoutbox.Repository{}

	if err = a.initKafka(cfg, infraLogger, ucResultInboxRepo, ucEventRepo, outboxRepo); err != nil {
		return err
	}

	authService := application.NewService(
		a.db,
		userRepo,
		ucEventRepo,
		emailTokenRepo,
		passwordResetTokenRepo,
		refreshTokenRepo,
		mailer,
		infraLogger,
		outboxRepo,
		cfg.Mailer.Bypass,
		cfg.Frontend.URL,
		cfg.Admin.Email,
		regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`),
	)
	a.authService = authService
	handlers := httptransport.NewHandlers(a.logger, authService, jwtManager, a.db, refreshTokenRepo)
	router := httptransport.NewRouter(a.logger, validator, handlers)
	a.server = &http.Server{
		Addr:    bootstrap.InitPortStringFromConfig(cfg, 8081),
		Handler: router,
	}

	if err = a.initGRPCServer(cfg, authService); err != nil {
		return err
	}

	return nil
}

func (a *App) initGRPCServer(cfg bootstrap.Config, authService *application.Service) error {
	if cfg.AuthGRPCListenAddr == "" {
		return fmt.Errorf("failed to initialize grpc server: auth grpc listen address is not configured")
	}

	listener, err := net.Listen("tcp", cfg.AuthGRPCListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen grpc: %w", err)
	}

	server := grpc.NewServer()
	authpb.RegisterAuthServiceServer(server, grpctransport.NewServer(authService))

	a.grpcServer = server
	a.grpcListener = listener

	return nil
}

func (a *App) initLoggers() {
	a.logger = logger.NewJSONLogger(slog.LevelDebug, "auth-service", "")
}

func (a *App) infrastructureLogger() *slog.Logger {
	return logger.NewJSONLogger(slog.LevelDebug, "", "infrastructure")
}

func (a *App) initMailer(cfg bootstrap.Config) (application.Mailer, error) {
	mailer := bootstrap.InitMailerFromConfig(cfg)
	if err := bootstrap.ValidateMailConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize mailer: %w", err)
	}

	return mailer, nil
}

func (a *App) initJWT(cfg bootstrap.Config) (*jwt.Manager, *validators.LocalJWT, error) {
	jwtManager, err := bootstrap.InitJWTManagerFromConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize JWT manager: %w", err)
	}

	validator, err := bootstrap.InitLocalJWTFromConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize JWT validator: %w", err)
	}

	return jwtManager, validator, nil
}

func (a *App) initKafka(
	cfg bootstrap.Config,
	infraLogger *slog.Logger,
	ucResultInboxRepo *ucrinbox.Repository,
	ucEventRepo *ucevent.Repository,
	outboxRepo *ucoutbox.Repository,
) error {
	if len(cfg.Kafka.Brokers) == 0 {
		return errors.New("failed to initialize kafka writer: kafka brokers are not configured")
	}
	if cfg.Kafka.UserCreationTopic == "" {
		return errors.New("failed to initialize kafka writer: user creation topic is not configured")
	}
	if cfg.Kafka.UserCreationResultTopic == "" {
		return errors.New("failed to initialize kafka consumer: user creation result topic is not configured")
	}
	if cfg.Kafka.UserCreationResultGroup == "" {
		return errors.New("failed to initialize kafka consumer: user creation result group is not configured")
	}

	kafkaWriter := kafkax.NewWriter(cfg.Kafka.Brokers, cfg.Kafka.UserCreationTopic)
	writerOwned := false
	defer func() {
		if !writerOwned {
			_ = kafkaWriter.Close()
		}
	}()

	topicInitCtx, cancelTopicInit := context.WithTimeout(context.Background(), cfg.Kafka.WriteTimeout)
	defer cancelTopicInit()

	if err := kafkax.EnsureTopic(topicInitCtx, cfg.Kafka.Brokers, cfg.Kafka.UserCreationTopic, 1, 1); err != nil {
		return fmt.Errorf("failed to ensure kafka topic: %w", err)
	}
	if err := kafkax.EnsureTopic(topicInitCtx, cfg.Kafka.Brokers, cfg.Kafka.UserCreationResultTopic, 1, 1); err != nil {
		return fmt.Errorf("failed to ensure kafka result topic: %w", err)
	}

	kafkaPublisher := kafkax.NewOutboxPublisher(
		kafkaWriter,
		infraLogger,
		cfg.Kafka.Brokers,
		cfg.Kafka.UserCreationTopic,
		cfg.Kafka.BatchSize,
		cfg.Kafka.PollInterval,
		cfg.Kafka.WriteTimeout,
	)

	a.outboxPublisher = authkafka.NewUserCreationOutboxPublisher(a.db, outboxRepo, infraLogger, kafkaPublisher)

	ucResultReader := kafkax.NewMessageReader(
		cfg.Kafka.Brokers,
		cfg.Kafka.UserCreationResultTopic,
		cfg.Kafka.UserCreationResultGroup,
	)
	ucResultKafkaConsumer := kafkax.NewInboxConsumer[usersauth.UCResultMessage](
		infraLogger,
		ucResultReader,
		cfg.Kafka.PollInterval,
	)
	a.ucResultConsumer = authconsumer.NewUCResultInboxConsumer(a.db, ucResultInboxRepo, ucResultKafkaConsumer)
	a.ucResultProcessor = ucrprocessor.NewProcessor(
		ucResultInboxRepo,
		ucEventRepo,
		a.db,
		infraLogger,
		cfg.Kafka.BatchSize,
		cfg.Kafka.PollInterval,
	)

	writerOwned = true
	return nil
}
