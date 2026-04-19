package app

import (
	authpb "barter-port/contracts/grpc/auth/v1"
	userspb "barter-port/contracts/grpc/users/v1"
	authusers "barter-port/contracts/kafka/messages/auth-users"
	dealsusers "barter-port/contracts/kafka/messages/deals-users"
	"barter-port/internal/users/application/user"
	userservice "barter-port/internal/users/application/user"
	"barter-port/internal/users/infrastructure/kafka/consumer"
	"barter-port/internal/users/infrastructure/kafka/producer"
	grpctransport "barter-port/internal/users/infrastructure/transport/grpc"
	httptransport "barter-port/internal/users/infrastructure/transport/http"
	"barter-port/pkg/bootstrap"
	"barter-port/pkg/db"
	"barter-port/pkg/kafkax"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (app *App) initAuthGRPCClient(cfg bootstrap.Config) (authpb.AuthServiceClient, error) {
	if cfg.AuthGRPCAddr == "" {
		return nil, fmt.Errorf("failed to initialize grpc server: auth grpc address is not configured")
	}

	conn, err := grpc.NewClient(cfg.AuthGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth grpc connection: %w", err)
	}

	app.authGRPCConn = conn
	return authpb.NewAuthServiceClient(conn), nil
}

func (app *App) initDatabase(cfg bootstrap.Config) error {
	pool, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		return errors.New("failed to initialize database: " + err.Error())
	}
	app.db = pool

	authDB, err := db.NewPostgres(db.Config{
		DBUser:     cfg.DB.User,
		DBPassword: cfg.DB.Password,
		DBHost:     cfg.DB.Host,
		DBPort:     cfg.DB.Port,
		DBName:     "auth_db",
	})
	if err != nil {
		return errors.New("failed to initialize auth database: " + err.Error())
	}
	app.authDB = authDB

	return nil
}

func (app *App) initUCEventConsumer(cfg bootstrap.Config) error {
	if app.log == nil {
		return errors.New("log is not initialized")
	}
	if app.db == nil {
		return errors.New("db is not initialized")
	}
	if app.inboxRepository == nil {
		return errors.New("inboxRepository is not initialized")
	}

	reader := kafkax.NewMessageReader(cfg.Kafka.Brokers, cfg.Kafka.UserCreationTopic, cfg.Kafka.UserCreationGroup)
	kafkaConsumer := kafkax.NewInboxConsumer[authusers.UserCreationMessage](app.log, reader, cfg.Kafka.PollInterval)
	app.ucEventConsumer = consumer.NewUserCreationInboxConsumer(app.db, app.inboxRepository, kafkaConsumer)

	return nil
}

func (app *App) initUCREventProducer(cfg bootstrap.Config) error {
	if app.log == nil {
		return errors.New("log is not initialized")
	}
	if app.db == nil {
		return errors.New("db is not initialized")
	}
	if app.outboxRepository == nil {
		return errors.New("outboxRepository is not initialized")
	}

	kafkaWriter := kafkax.NewWriter(cfg.Kafka.Brokers, cfg.Kafka.UserCreationResultTopic)

	kafkaPublisher := kafkax.NewOutboxPublisher(
		kafkaWriter,
		app.log,
		cfg.Kafka.Brokers,
		cfg.Kafka.UserCreationResultTopic,
		cfg.Kafka.BatchSize,
		cfg.Kafka.PollInterval,
		cfg.Kafka.WriteTimeout,
	)

	app.ucrEventProducer = producer.NewUCResultOutbox(app.db, app.outboxRepository, app.log, kafkaPublisher)

	return nil
}

func (app *App) initReputationEventConsumer(cfg bootstrap.Config) error {
	if app.log == nil {
		return errors.New("log is not initialized")
	}
	if app.db == nil {
		return errors.New("db is not initialized")
	}
	if app.reputationInboxRepo == nil {
		return errors.New("reputationInboxRepo is not initialized")
	}
	if cfg.Kafka.OfferReportPenaltyTopic == "" {
		return errors.New("offer report penalty topic is not configured")
	}
	if cfg.Kafka.OfferReportPenaltyGroup == "" {
		return errors.New("offer report penalty group is not configured")
	}

	topicInitCtx, cancelTopicInit := context.WithTimeout(context.Background(), cfg.Kafka.WriteTimeout)
	defer cancelTopicInit()

	if err := kafkax.EnsureTopic(topicInitCtx, cfg.Kafka.Brokers, cfg.Kafka.OfferReportPenaltyTopic, 1, 1); err != nil {
		return fmt.Errorf("failed to ensure offer report penalty topic: %w", err)
	}

	reader := kafkax.NewMessageReader(cfg.Kafka.Brokers, cfg.Kafka.OfferReportPenaltyTopic, cfg.Kafka.OfferReportPenaltyGroup)
	kafkaConsumer := kafkax.NewInboxConsumer[dealsusers.ReputationMessage](app.log, reader, cfg.Kafka.PollInterval)
	app.reputationEventConsumer = consumer.NewReputationInboxConsumer(app.db, app.reputationInboxRepo, kafkaConsumer)

	return nil
}

func (app *App) initHTTPServer(cfg bootstrap.Config, userService *userservice.Service) error {
	if app.log == nil {
		return errors.New("log is not initialized")
	}
	if userService == nil {
		return errors.New("userService is not initialized")
	}

	validator, err := bootstrap.InitLocalJWTFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize jwt validator: %w", err)
	}

	handlers := httptransport.NewHandlers(userService)
	router := httptransport.NewRouter(app.log, validator, handlers)
	app.server = &http.Server{
		Addr:    bootstrap.InitPortStringFromConfig(cfg, 8082),
		Handler: router,
	}

	return nil
}

func (app *App) initGRPCServer(cfg bootstrap.Config, usersService *user.Service) error {
	if cfg.UsersGRPCListenAddr == "" {
		return fmt.Errorf("failed to initialize grpc server: users grpc listen address is not configured")
	}

	listener, err := net.Listen("tcp", cfg.UsersGRPCListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen grpc: %w", err)
	}

	server := grpc.NewServer()
	userspb.RegisterUsersServiceServer(server, grpctransport.NewServer(usersService))

	app.grpcServer = server
	app.grpcListener = listener

	return nil
}
