package inbox

import (
	authusers "barter-port/internal/contracts/kafka/auth-users"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestRepository_WriteUserCreationMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewRepository()
	message := authusers.UserCreationMessage{
		ID:        uuid.New(),
		EventID:   uuid.New(),
		UserID:    uuid.New(),
		CreatedAt: time.Date(2026, time.March, 21, 12, 0, 0, 0, time.UTC),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db := newMockDB(t)
		db.ExpectExec(`
		INSERT INTO user_creation_inbox (id, event_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)`).
			WithArgs(message.ID, message.EventID, message.UserID, message.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.WriteUserCreationMessage(ctx, db, message)
		require.NoError(t, err)
	})

	t.Run("unique violation", func(t *testing.T) {
		t.Parallel()

		db := newMockDB(t)
		db.ExpectExec(`
		INSERT INTO user_creation_inbox (id, event_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)`).
			WithArgs(message.ID, message.EventID, message.UserID, message.CreatedAt).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		err := repo.WriteUserCreationMessage(ctx, db, message)
		require.ErrorIs(t, err, ErrUCEventAlreadyExists)
	})

	t.Run("exec error", func(t *testing.T) {
		t.Parallel()

		wantErr := &pgconn.PgError{
			Code:    pgerrcode.ConnectionException,
			Message: "exec failed",
		}
		db := newMockDB(t)
		db.ExpectExec(`
		INSERT INTO user_creation_inbox (id, event_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)`).
			WithArgs(message.ID, message.EventID, message.UserID, message.CreatedAt).
			WillReturnError(wantErr)

		err := repo.WriteUserCreationMessage(ctx, db, message)
		var gotErr *pgconn.PgError
		require.ErrorAs(t, err, &gotErr)
		require.Equal(t, wantErr.Code, gotErr.Code)
		require.Equal(t, wantErr.Message, gotErr.Message)
	})
}

func TestRepository_ReadUserCreationMessagesForUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewRepository()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		first := authusers.UserCreationMessage{
			ID:        uuid.New(),
			EventID:   uuid.New(),
			UserID:    uuid.New(),
			CreatedAt: time.Date(2026, time.March, 21, 10, 0, 0, 0, time.UTC),
		}
		second := authusers.UserCreationMessage{
			ID:        uuid.New(),
			EventID:   uuid.New(),
			UserID:    uuid.New(),
			CreatedAt: time.Date(2026, time.March, 21, 11, 0, 0, 0, time.UTC),
		}

		rows := pgxmock.NewRows([]string{"id", "event_id", "user_id", "created_at"}).
			AddRow(first.ID, first.EventID, first.UserID, first.CreatedAt).
			AddRow(second.ID, second.EventID, second.UserID, second.CreatedAt)
		db := newMockDB(t)
		db.ExpectQuery(`
		SELECT id, event_id, user_id, created_at FROM user_creation_inbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`).
			WithArgs(2).
			WillReturnRows(rows).
			RowsWillBeClosed()

		got, err := repo.ReadUserCreationMessagesForUpdate(ctx, db, 2)
		require.NoError(t, err)
		require.Equal(t, []authusers.UserCreationMessage{first, second}, got)
	})

	t.Run("query error", func(t *testing.T) {
		t.Parallel()

		wantErr := &pgconn.PgError{
			Code:    pgerrcode.ConnectionException,
			Message: "query failed",
		}
		db := newMockDB(t)
		db.ExpectQuery(`
		SELECT id, event_id, user_id, created_at FROM user_creation_inbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`).
			WithArgs(3).
			WillReturnError(wantErr)

		got, err := repo.ReadUserCreationMessagesForUpdate(ctx, db, 3)
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
		rows := pgxmock.NewRows([]string{"id", "event_id", "user_id", "created_at"}).
			AddRow(uuid.New(), uuid.New(), uuid.New(), time.Now().UTC()).
			RowError(0, wantErr)
		db := newMockDB(t)
		db.ExpectQuery(`
		SELECT id, event_id, user_id, created_at FROM user_creation_inbox
		ORDER BY created_at, id LIMIT $1
		FOR UPDATE SKIP LOCKED`).
			WithArgs(1).
			WillReturnRows(rows).
			RowsWillBeClosed()

		got, err := repo.ReadUserCreationMessagesForUpdate(ctx, db, 1)
		var gotErr *pgconn.PgError
		require.ErrorAs(t, err, &gotErr)
		require.Equal(t, wantErr.Code, gotErr.Code)
		require.Equal(t, wantErr.Message, gotErr.Message)
		require.Nil(t, got)
	})
}

func TestRepository_DeleteUserCreationMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewRepository()
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db := newMockDB(t)
		db.ExpectExec(`
		DELETE FROM user_creation_inbox
		WHERE id = $1`).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteUserCreationMessage(ctx, db, id)
		require.NoError(t, err)
	})

	t.Run("exec error", func(t *testing.T) {
		t.Parallel()

		wantErr := &pgconn.PgError{
			Code:    pgerrcode.ConnectionException,
			Message: "delete failed",
		}
		db := newMockDB(t)
		db.ExpectExec(`
		DELETE FROM user_creation_inbox
		WHERE id = $1`).
			WithArgs(id).
			WillReturnError(wantErr)

		err := repo.DeleteUserCreationMessage(ctx, db, id)
		var gotErr *pgconn.PgError
		require.ErrorAs(t, err, &gotErr)
		require.Equal(t, wantErr.Code, gotErr.Code)
		require.Equal(t, wantErr.Message, gotErr.Message)
	})
}

func newMockDB(t *testing.T) pgxmock.PgxConnIface {
	t.Helper()

	mock, err := pgxmock.NewConn(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, mock.ExpectationsWereMet())
	})

	return mock
}
