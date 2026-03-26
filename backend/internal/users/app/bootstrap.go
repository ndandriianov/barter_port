package app

import (
	authusers "barter-port/contracts/kafka/messages/auth-users"
	"barter-port/internal/users/infrastructure/kafka/consumer"
	"barter-port/internal/users/infrastructure/kafka/producer"
	"barter-port/pkg/bootstrap"
	"barter-port/pkg/kafkax"
	"errors"
)

func (app *App) initDatabase(cfg bootstrap.Config) error {
	db, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		return errors.New("failed to initialize database: " + err.Error())
	}
	app.db = db
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
