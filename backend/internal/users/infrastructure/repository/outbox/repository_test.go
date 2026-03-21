package outbox

import (
	usersauth "barter-port/internal/contracts/kafka/messages/users-auth"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestRepository_WriteUCResultMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewRepository()
	message := usersauth.UCResultMessage{
		ID:        uuid.New(),
		EventID:   uuid.New(),
		UserID:    uuid.New(),
		Status:    "success",
		CreatedAt: time.Date(2026, time.March, 21, 12, 0, 0, 0, time.UTC),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tx := newMockTx(t)
		tx.ExpectExec(`
		INSERT INTO user_creation_result_outbox (id, event_id, user_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)`).
			WithArgs(message.ID, message.EventID, message.UserID, message.Status, message.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.WriteUCResultMessage(ctx, tx, message)
		require.NoError(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		t.Parallel()

		wantErr := &pgconn.PgError{
			Code:    pgerrcode.ConnectionException,
			Message: "exec failed",
		}
		tx := newMockTx(t)
		tx.ExpectExec(`
		INSERT INTO user_creation_result_outbox (id, event_id, user_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)`).
			WithArgs(message.ID, message.EventID, message.UserID, message.Status, message.CreatedAt).
			WillReturnError(wantErr)

		err := repo.WriteUCResultMessage(ctx, tx, message)
		var gotErr *pgconn.PgError
		require.ErrorAs(t, err, &gotErr)
		require.Equal(t, wantErr.Code, gotErr.Code)
		require.Equal(t, wantErr.Message, gotErr.Message)
	})
}

func TestRepository_ReadUCResultMessagesForUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewRepository()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		first := usersauth.UCResultMessage{
			ID:        uuid.New(),
			EventID:   uuid.New(),
			UserID:    uuid.New(),
			Status:    "success",
			CreatedAt: time.Date(2026, time.March, 21, 10, 0, 0, 0, time.UTC),
		}
		second := usersauth.UCResultMessage{
			ID:        uuid.New(),
			EventID:   uuid.New(),
			UserID:    uuid.New(),
			Status:    "failed",
			CreatedAt: time.Date(2026, time.March, 21, 11, 0, 0, 0, time.UTC),
		}

		rows := pgxmock.NewRows([]string{"id", "event_id", "user_id", "status", "created_at"}).
			AddRow(first.ID, first.EventID, first.UserID, first.Status, first.CreatedAt).
			AddRow(second.ID, second.EventID, second.UserID, second.Status, second.CreatedAt)
		tx := newMockTx(t)
		tx.ExpectQuery(`
		SELECT id, event_id, user_id, status, created_at FROM user_creation_result_outbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`).
			WithArgs(2).
			WillReturnRows(rows).
			RowsWillBeClosed()

		got, err := repo.ReadUCResultMessagesForUpdate(ctx, tx, 2)
		require.NoError(t, err)
		require.Equal(t, []usersauth.UCResultMessage{first, second}, got)
	})

	t.Run("query error", func(t *testing.T) {
		t.Parallel()

		wantErr := &pgconn.PgError{
			Code:    pgerrcode.ConnectionException,
			Message: "query failed",
		}
		tx := newMockTx(t)
		tx.ExpectQuery(`
		SELECT id, event_id, user_id, status, created_at FROM user_creation_result_outbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`).
			WithArgs(3).
			WillReturnError(wantErr)

		got, err := repo.ReadUCResultMessagesForUpdate(ctx, tx, 3)
		var gotErr *pgconn.PgError
		require.ErrorAs(t, err, &gotErr)
		require.Equal(t, wantErr.Code, gotErr.Code)
		require.Equal(t, wantErr.Message, gotErr.Message)
		require.Nil(t, got)
	})

	t.Run("scan error", func(t *testing.T) {
		t.Parallel()

		wantErr := &pgconn.PgError{
			Code:    pgerrcode.InvalidTextRepresentation,
			Message: "scan failed",
		}
		rows := pgxmock.NewRows([]string{"id", "event_id", "user_id", "status", "created_at"}).
			AddRow(uuid.New(), uuid.New(), uuid.New(), "success", time.Now().UTC()).
			RowError(0, wantErr)
		tx := newMockTx(t)
		tx.ExpectQuery(`
		SELECT id, event_id, user_id, status, created_at FROM user_creation_result_outbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`).
			WithArgs(1).
			WillReturnRows(rows).
			RowsWillBeClosed()

		got, err := repo.ReadUCResultMessagesForUpdate(ctx, tx, 1)
		var gotErr *pgconn.PgError
		require.ErrorAs(t, err, &gotErr)
		require.Equal(t, wantErr.Code, gotErr.Code)
		require.Equal(t, wantErr.Message, gotErr.Message)
		require.Nil(t, got)
	})
}

func TestRepository_DeleteUCResultMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewRepository()
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tx := newMockTx(t)
		tx.ExpectExec(`
		DELETE FROM user_creation_result_outbox
       	WHERE id = $1`).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteUCResultMessage(ctx, tx, id)
		require.NoError(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		t.Parallel()

		wantErr := &pgconn.PgError{
			Code:    pgerrcode.ConnectionException,
			Message: "delete failed",
		}
		tx := newMockTx(t)
		tx.ExpectExec(`
		DELETE FROM user_creation_result_outbox
       	WHERE id = $1`).
			WithArgs(id).
			WillReturnError(wantErr)

		err := repo.DeleteUCResultMessage(ctx, tx, id)
		var gotErr *pgconn.PgError
		require.ErrorAs(t, err, &gotErr)
		require.Equal(t, wantErr.Code, gotErr.Code)
		require.Equal(t, wantErr.Message, gotErr.Message)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		tx := newMockTx(t)
		tx.ExpectExec(`
		DELETE FROM user_creation_result_outbox
       	WHERE id = $1`).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err := repo.DeleteUCResultMessage(ctx, tx, id)
		require.ErrorIs(t, err, ErrUCResulMessageNotFound)
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
