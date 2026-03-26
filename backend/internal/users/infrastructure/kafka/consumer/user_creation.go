package consumer

import (
	authusers "barter-port/internal/contracts/kafka/messages/auth-users"
	"barter-port/internal/users/infrastructure/repository/inbox"
	"barter-port/pkg/db"
	"barter-port/pkg/kafkax"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserCreationInboxConsumer struct {
	db        *pgxpool.Pool
	inboxRepo inboxWriter
	consumer  *kafkax.InboxConsumer[authusers.UserCreationMessage]
}

type inboxWriter interface {
	WriteUserCreationMessage(context.Context, db.DB, authusers.UserCreationMessage) error
}

func NewUserCreationInboxConsumer(
	db *pgxpool.Pool,
	inboxRepo inboxWriter,
	consumer *kafkax.InboxConsumer[authusers.UserCreationMessage],
) *UserCreationInboxConsumer {
	return &UserCreationInboxConsumer{
		db:        db,
		inboxRepo: inboxRepo,
		consumer:  consumer,
	}
}

func (c *UserCreationInboxConsumer) Run(ctx context.Context) error {
	return c.consumer.Run(ctx, c.consumeMessage)
}

func (c *UserCreationInboxConsumer) consumeMessage(ctx context.Context, message authusers.UserCreationMessage) error {
	err := c.inboxRepo.WriteUserCreationMessage(ctx, c.db, message)
	if err != nil {
		if errors.Is(err, inbox.ErrUCEventAlreadyExists) {
			return kafkax.ErrDuplicate
		}
		return fmt.Errorf("failed to write user creation message to inbox: %w", err)
	}

	return nil
}
