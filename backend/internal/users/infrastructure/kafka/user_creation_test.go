package kafka

import (
	authusers "barter-port/internal/contracts/kafka/messages/auth-users"
	"barter-port/internal/libs/db"
	"barter-port/internal/users/infrastructure/repository/inbox"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewUserCreationInboxConsumer(t *testing.T) {
	t.Parallel()

	t.Run("uses default poll interval", func(t *testing.T) {
		t.Parallel()

		reader := newMockMessageReader(t)
		repo := &mockInboxWriter{}
		mockDB := newMockDB(t)

		consumer := NewUserCreationInboxConsumer(Params{
			Log:       testLogger(),
			Reader:    reader,
			DB:        mockDB,
			InboxRepo: repo,
		})

		require.Equal(t, 5*time.Second, consumer.pollInterval)
		require.Same(t, reader, consumer.reader)
		require.Same(t, mockDB, consumer.db)
		require.Same(t, repo, consumer.inboxRepo)
	})

	t.Run("uses provided poll interval", func(t *testing.T) {
		t.Parallel()

		reader := newMockMessageReader(t)
		repo := &mockInboxWriter{}
		mockDB := newMockDB(t)

		consumer := NewUserCreationInboxConsumer(Params{
			Log:          testLogger(),
			Reader:       reader,
			DB:           mockDB,
			InboxRepo:    repo,
			PollInterval: 2 * time.Second,
		})

		require.Equal(t, 2*time.Second, consumer.pollInterval)
	})
}

func TestUserCreationInboxConsumerConsumeMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	message := authusers.UserCreationMessage{
		ID:        uuid.New(),
		EventID:   uuid.New(),
		UserID:    uuid.New(),
		CreatedAt: time.Date(2026, time.March, 21, 12, 0, 0, 0, time.UTC),
	}
	messageJSON, err := json.Marshal(message)
	require.NoError(t, err)

	t.Run("returns fetch error", func(t *testing.T) {
		t.Parallel()

		reader := newMockMessageReader(t)
		reader.On("FetchMessage", mock.Anything).Return(kafkago.Message{}, errors.New("fetch failed")).Once()
		consumer := newTestConsumer(t, reader, &mockInboxWriter{})

		err := consumer.consumeMessage(ctx)
		require.EqualError(t, err, "failed to fetch message: fetch failed")
		reader.AssertNotCalled(t, "CommitMessages", mock.Anything, mock.Anything)
	})

	t.Run("commits bad message on unmarshal error", func(t *testing.T) {
		t.Parallel()

		rawMessage := kafkago.Message{
			Key:   []byte("bad-json"),
			Value: []byte("{"),
		}
		reader := newMockMessageReader(t)
		reader.On("FetchMessage", mock.Anything).Return(rawMessage, nil).Once()
		reader.On("CommitMessages", mock.Anything, rawMessage).Return(nil).Once()
		consumer := newTestConsumer(t, reader, &mockInboxWriter{})

		err := consumer.consumeMessage(ctx)
		require.EqualError(t, err, "failed to unmarshal message id: bad-json: unexpected end of JSON input")
	})

	t.Run("returns commit error for bad message", func(t *testing.T) {
		t.Parallel()

		rawMessage := kafkago.Message{
			Key:   []byte("bad-json"),
			Value: []byte("{"),
		}
		reader := newMockMessageReader(t)
		reader.On("FetchMessage", mock.Anything).Return(rawMessage, nil).Once()
		reader.On("CommitMessages", mock.Anything, rawMessage).Return(errors.New("commit failed")).Once()
		consumer := newTestConsumer(t, reader, &mockInboxWriter{})

		err := consumer.consumeMessage(ctx)
		require.EqualError(
			t,
			err,
			"failed to unmarshal message id=bad-json: unexpected end of JSON input; additionally failed to commit bad message: commit failed",
		)
	})

	t.Run("returns repository error", func(t *testing.T) {
		t.Parallel()

		rawMessage := kafkago.Message{
			Key:   []byte("event-1"),
			Value: messageJSON,
		}
		reader := newMockMessageReader(t)
		reader.On("FetchMessage", mock.Anything).Return(rawMessage, nil).Once()
		repo := &mockInboxWriter{writeErr: inbox.ErrUCEventAlreadyExists}
		consumer := newTestConsumer(t, reader, repo)

		err := consumer.consumeMessage(ctx)
		require.EqualError(t, err, "failed to write user creation message to inbox: event already exists")
		require.True(t, repo.called)
		require.Equal(t, message, repo.gotMessage)
		reader.AssertNotCalled(t, "CommitMessages", mock.Anything, mock.Anything)
	})

	t.Run("returns commit error after successful write", func(t *testing.T) {
		t.Parallel()

		rawMessage := kafkago.Message{
			Key:   []byte("event-1"),
			Value: messageJSON,
		}
		reader := newMockMessageReader(t)
		reader.On("FetchMessage", mock.Anything).Return(rawMessage, nil).Once()
		reader.On("CommitMessages", mock.Anything, rawMessage).Return(errors.New("commit failed")).Once()
		repo := &mockInboxWriter{}
		consumer := newTestConsumer(t, reader, repo)

		err := consumer.consumeMessage(ctx)
		require.EqualError(t, err, "failed to commit message id: event-1: commit failed")
		require.True(t, repo.called)
		require.Equal(t, message, repo.gotMessage)
	})

	t.Run("writes and commits valid message", func(t *testing.T) {
		t.Parallel()

		rawMessage := kafkago.Message{
			Key:   []byte("event-1"),
			Value: messageJSON,
		}
		reader := newMockMessageReader(t)
		reader.On("FetchMessage", mock.Anything).Return(rawMessage, nil).Once()
		reader.On("CommitMessages", mock.Anything, rawMessage).Return(nil).Once()
		repo := &mockInboxWriter{}
		consumer := newTestConsumer(t, reader, repo)

		err := consumer.consumeMessage(ctx)
		require.NoError(t, err)
		require.True(t, repo.called)
		require.Equal(t, message, repo.gotMessage)
	})
}

func TestUserCreationInboxConsumerRun(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for canceled context before fetch", func(t *testing.T) {
		t.Parallel()

		reader := newMockMessageReader(t)
		reader.On("Close").Return(nil).Once()
		consumer := newTestConsumer(t, reader, &mockInboxWriter{})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := consumer.Run(ctx)
		require.NoError(t, err)
		reader.AssertNotCalled(t, "FetchMessage", mock.Anything)
	})

	t.Run("returns nil on shutdown fetch error", func(t *testing.T) {
		t.Parallel()

		reader := newMockMessageReader(t)
		reader.On("FetchMessage", mock.Anything).Run(func(args mock.Arguments) {
			<-args.Get(0).(context.Context).Done()
		}).Return(kafkago.Message{}, context.Canceled).Once()
		reader.On("Close").Return(nil).Once()
		consumer := newTestConsumer(t, reader, &mockInboxWriter{})

		ctx, cancel := context.WithCancel(context.Background())
		go cancel()

		err := consumer.Run(ctx)
		require.NoError(t, err)
	})
}

func newTestConsumer(t *testing.T, reader *mockMessageReader, repo *mockInboxWriter) *UserCreationInboxConsumer {
	t.Helper()

	return NewUserCreationInboxConsumer(Params{
		Log:          testLogger(),
		Reader:       reader,
		DB:           newMockDB(t),
		InboxRepo:    repo,
		PollInterval: time.Millisecond,
	})
}

func newMockDB(t *testing.T) pgxmock.PgxConnIface {
	t.Helper()

	mockDB, err := pgxmock.NewConn(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, mockDB.ExpectationsWereMet())
	})

	return mockDB
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type mockMessageReader struct {
	mock.Mock
}

func newMockMessageReader(t *testing.T) *mockMessageReader {
	t.Helper()

	reader := &mockMessageReader{}
	t.Cleanup(func() {
		reader.AssertExpectations(t)
	})

	return reader
}

func (m *mockMessageReader) FetchMessage(ctx context.Context) (kafkago.Message, error) {
	args := m.Called(ctx)
	message, _ := args.Get(0).(kafkago.Message)
	return message, args.Error(1)
}

func (m *mockMessageReader) CommitMessages(ctx context.Context, messages ...kafkago.Message) error {
	callArgs := []interface{}{ctx}
	for _, message := range messages {
		callArgs = append(callArgs, message)
	}

	args := m.Called(callArgs...)
	return args.Error(0)
}

func (m *mockMessageReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockInboxWriter struct {
	writeErr   error
	called     bool
	gotMessage authusers.UserCreationMessage
}

func (m *mockInboxWriter) WriteUserCreationMessage(
	_ context.Context,
	_ db.DB,
	message authusers.UserCreationMessage,
) error {
	m.called = true
	m.gotMessage = message
	return m.writeErr
}
