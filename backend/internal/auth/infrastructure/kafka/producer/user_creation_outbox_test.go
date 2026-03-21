package producer

import (
	"barter-port/internal/auth/infrastructure/repository/outbox"
	authusers "barter-port/internal/contracts/kafka/messages/auth-users"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewUserCreationOutboxPublisher(t *testing.T) {
	t.Parallel()

	t.Run("applies defaults and copies brokers", func(t *testing.T) {
		t.Parallel()

		brokers := []string{"broker-1", "broker-2"}
		publisher := NewUserCreationOutboxPublisher(
			nil,
			&outbox.Repository{},
			nil,
			testOutboxLogger(),
			brokers,
			"users.created",
			0,
			0,
			0,
		)

		brokers[0] = "changed"

		require.Equal(t, 100, publisher.batchSize)
		require.Equal(t, 2*time.Second, publisher.pollInterval)
		require.Equal(t, 10*time.Second, publisher.writeTimeout)
		require.Equal(t, []string{"broker-1", "broker-2"}, publisher.brokers)
	})

	t.Run("uses provided values", func(t *testing.T) {
		t.Parallel()

		publisher := NewUserCreationOutboxPublisher(
			nil,
			&outbox.Repository{},
			nil,
			testOutboxLogger(),
			[]string{"broker-1"},
			"users.created",
			5,
			3*time.Second,
			4*time.Second,
		)

		require.Equal(t, 5, publisher.batchSize)
		require.Equal(t, 3*time.Second, publisher.pollInterval)
		require.Equal(t, 4*time.Second, publisher.writeTimeout)
	})
}

func TestBuildMessages(t *testing.T) {
	t.Parallel()

	message := authusers.UserCreationMessage{
		ID:        uuid.New(),
		EventID:   uuid.New(),
		UserID:    uuid.New(),
		CreatedAt: time.Date(2026, time.March, 21, 14, 30, 0, 0, time.UTC),
	}

	got, err := buildMessages([]authusers.UserCreationMessage{message})
	require.NoError(t, err)
	require.Len(t, got, 1)

	payload, err := json.Marshal(message)
	require.NoError(t, err)

	require.Equal(t, []byte(message.UserID.String()), got[0].Key)
	require.Equal(t, payload, got[0].Value)
	require.Equal(t, message.CreatedAt, got[0].Time)
	require.Equal(t, []kafkago.Header{
		{Key: "message_id", Value: []byte(message.ID.String())},
		{Key: "message_type", Value: []byte(userCreationEventType)},
	}, got[0].Headers)
}

func TestUserCreationOutboxPublisherClose(t *testing.T) {
	t.Parallel()

	t.Run("closes writer", func(t *testing.T) {
		t.Parallel()

		writer := newMockOutboxWriter(t)
		writer.On("Close").Return(nil).Once()

		publisher := &UserCreationOutboxPublisher{writer: writer}
		require.NoError(t, publisher.Close())
	})

	t.Run("returns writer close error", func(t *testing.T) {
		t.Parallel()

		writer := newMockOutboxWriter(t)
		writer.On("Close").Return(errors.New("close failed")).Once()

		publisher := &UserCreationOutboxPublisher{writer: writer}
		require.EqualError(t, publisher.Close(), "close failed")
	})
}

func testOutboxLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type mockOutboxWriter struct {
	mock.Mock
}

func newMockOutboxWriter(t *testing.T) *mockOutboxWriter {
	t.Helper()

	writer := &mockOutboxWriter{}
	t.Cleanup(func() {
		writer.AssertExpectations(t)
	})

	return writer
}

func (m *mockOutboxWriter) WriteMessages(ctx context.Context, msgs ...kafkago.Message) error {
	callArgs := []interface{}{ctx}
	for _, msg := range msgs {
		callArgs = append(callArgs, msg)
	}

	args := m.Called(callArgs...)
	return args.Error(0)
}

func (m *mockOutboxWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}
