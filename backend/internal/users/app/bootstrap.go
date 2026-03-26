package app

import (
	authusers "barter-port/internal/contracts/kafka/messages/auth-users"
	"barter-port/internal/users/infrastructure/kafka/consumer"
	bootstrap2 "barter-port/pkg/bootstrap"
	kafkax2 "barter-port/pkg/kafkax"
	"errors"
)

func (app *App) initDatabase(cfg bootstrap2.Config) error {
	db, err := bootstrap2.InitDatabaseFromConfig(cfg)
	if err != nil {
		return errors.New("failed to initialize database: " + err.Error())
	}
	app.db = db
	return nil
}

func (app *App) initUCEventConsumer(cfg bootstrap2.Config) error {
	if app.log == nil {
		return errors.New("log is not initialized")
	}
	if app.db == nil {
		return errors.New("db is not initialized")
	}
	if app.inboxRepository == nil {
		return errors.New("inboxRepository is not initialized")
	}

	reader := kafkax2.NewMessageReader(cfg.Kafka.Brokers, cfg.Kafka.UserCreationTopic, cfg.Kafka.UserCreationGroup)
	kafkaConsumer := kafkax2.NewInboxConsumer[authusers.UserCreationMessage](app.log, reader, cfg.Kafka.PollInterval)
	app.ucEventConsumer = consumer.NewUserCreationInboxConsumer(app.db, app.inboxRepository, kafkaConsumer)

	return nil
}
