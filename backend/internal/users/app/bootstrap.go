package app

import (
	authusers "barter-port/internal/contracts/kafka/messages/auth-users"
	"barter-port/internal/libs/bootstrap"
	"barter-port/internal/libs/kafkax"
	"barter-port/internal/users/infrastructure/kafka/consumer"
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
