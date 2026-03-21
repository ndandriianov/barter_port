package outbox

import (
	authusers "barter-port/internal/contracts/kafka/auth-users"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestRepository_WriteUserCreationEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	message := authusers.UserCreationMessage{
		ID:        uuid.New(),
		EventID:   uuid.New(),
		UserID:    uuid.New(),
		CreatedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tx := newMockTx(t)
		repo := &Repository{}
		tx.ExpectExec(`
		INSERT INTO user_creation_outbox (id, event_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)`).
			WithArgs(message.ID, message.EventID, message.UserID, message.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.WriteUserCreationMessage(ctx, tx, message)
		require.NoError(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("exec failed")
		tx := newMockTx(t)
		repo := &Repository{}
		tx.ExpectExec(`
		INSERT INTO user_creation_outbox (id, event_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)`).
			WithArgs(message.ID, message.EventID, message.UserID, message.CreatedAt).
			WillReturnError(wantErr)

		err := repo.WriteUserCreationMessage(ctx, tx, message)
		require.ErrorIs(t, err, wantErr)
	})
}

func TestRepository_ReadUserCreationEventsForUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &Repository{}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		first := authusers.UserCreationMessage{
			ID:        uuid.New(),
			EventID:   uuid.New(),
			UserID:    uuid.New(),
			CreatedAt: time.Date(2026, time.March, 20, 10, 0, 0, 0, time.UTC),
		}
		second := authusers.UserCreationMessage{
			ID:        uuid.New(),
			EventID:   uuid.New(),
			UserID:    uuid.New(),
			CreatedAt: time.Date(2026, time.March, 20, 11, 0, 0, 0, time.UTC),
		}

		rows := pgxmock.NewRows([]string{"id", "event_id", "user_id", "created_at"}).
			AddRow(first.ID, first.EventID, first.UserID, first.CreatedAt).
			AddRow(second.ID, second.EventID, second.UserID, second.CreatedAt)
		tx := newMockTx(t)
		tx.ExpectQuery(`
		SELECT id, event_id, user_id, created_at FROM user_creation_outbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`).
			WithArgs(2).
			WillReturnRows(rows).
			RowsWillBeClosed()

		got, err := repo.ReadUserCreationMessagesForUpdate(ctx, tx, 2)
		require.NoError(t, err)

		want := []authusers.UserCreationMessage{first, second}
		require.Equal(t, want, got)
	})

	t.Run("query error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("query failed")
		tx := newMockTx(t)
		tx.ExpectQuery(`
		SELECT id, event_id, user_id, created_at FROM user_creation_outbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`).
			WithArgs(3).
			WillReturnError(wantErr)

		got, err := repo.ReadUserCreationMessagesForUpdate(ctx, tx, 3)
		require.ErrorIs(t, err, wantErr)
		require.Nil(t, got)
	})

	t.Run("scan error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("scan failed")
		rows := pgxmock.NewRows([]string{"id", "event_id", "user_id", "created_at"}).
			AddRow(uuid.New(), uuid.New(), uuid.New(), time.Now().UTC()).
			RowError(0, wantErr)
		tx := newMockTx(t)
		tx.ExpectQuery(`
		SELECT id, event_id, user_id, created_at FROM user_creation_outbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`).
			WithArgs(1).
			WillReturnRows(rows).
			RowsWillBeClosed()

		got, err := repo.ReadUserCreationMessagesForUpdate(ctx, tx, 1)
		require.ErrorIs(t, err, wantErr)
		require.Nil(t, got)
	})
}

func TestRepository_DeleteUserCreationEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &Repository{}
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tx := newMockTx(t)
		tx.ExpectExec(`
		DELETE FROM user_creation_outbox
		WHERE id = $1`).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteUserCreationMessage(ctx, tx, id)
		require.NoError(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("delete failed")
		tx := newMockTx(t)
		tx.ExpectExec(`
		DELETE FROM user_creation_outbox
		WHERE id = $1`).
			WithArgs(id).
			WillReturnError(wantErr)

		err := repo.DeleteUserCreationMessage(ctx, tx, id)
		require.ErrorIs(t, err, wantErr)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		tx := newMockTx(t)
		tx.ExpectExec(`
		DELETE FROM user_creation_outbox
		WHERE id = $1`).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err := repo.DeleteUserCreationMessage(ctx, tx, id)
		require.ErrorIs(t, err, ErrUserCreationMessageNotFound)
	})
}

func newMockTx(t *testing.T) pgxmock.PgxConnIface {
	t.Helper()

	mock, err := pgxmock.NewConn(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, mock.ExpectationsWereMet())
	})

	return mock
}
