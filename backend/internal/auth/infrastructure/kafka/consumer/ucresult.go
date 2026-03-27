package consumer

import (
	usersauth "barter-port/contracts/kafka/messages/users-auth"
	ucrinbox "barter-port/internal/auth/infrastructure/repository/uc-result-inbox"
	"barter-port/pkg/db"
	"barter-port/pkg/kafkax"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UCResultInboxConsumer struct {
	db        *pgxpool.Pool
	inboxRepo inboxWriter
	consumer  *kafkax.InboxConsumer[usersauth.UCResultMessage]
}

type inboxWriter interface {
	WriteUCResultMessage(context.Context, db.DB, usersauth.UCResultMessage) error
}

func NewUCResultInboxConsumer(
	db *pgxpool.Pool,
	inboxRepo inboxWriter,
	consumer *kafkax.InboxConsumer[usersauth.UCResultMessage],
) *UCResultInboxConsumer {
	return &UCResultInboxConsumer{
		db:        db,
		inboxRepo: inboxRepo,
		consumer:  consumer,
	}
}

func (c *UCResultInboxConsumer) Run(ctx context.Context) error {
	return c.consumer.Run(ctx, c.consumeMessage)
}

func (c *UCResultInboxConsumer) consumeMessage(ctx context.Context, message usersauth.UCResultMessage) error {
	err := c.inboxRepo.WriteUCResultMessage(ctx, c.db, message)
	if err != nil {
		if errors.Is(err, ucrinbox.ErrUCResultEventAlreadyExists) {
			return kafkax.ErrDuplicate
		}
		return fmt.Errorf("failed to write uc-result message to inbox: %w", err)
	}

	return nil
}
