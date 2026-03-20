package outbox

import (
	"barter-port/internal/auth/domain"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestRepository_WriteUserCreationEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	event := domain.UserCreationEvent{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		CreatedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tx := &stubTx{}
		repo := &Repository{}

		err := repo.WriteUserCreationEvent(ctx, tx, event)
		require.NoError(t, err)

		require.Equal(t, normalizeSQL(`
			INSERT INTO user_creation_outbox (id, user_id, created_at)
			VALUES ($1, $2, $3)`), normalizeSQL(tx.execSQL))

		wantArgs := []any{event.ID, event.UserID, event.CreatedAt}
		require.Equal(t, wantArgs, tx.execArgs)
	})

	t.Run("exec error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("exec failed")
		tx := &stubTx{execErr: wantErr}
		repo := &Repository{}

		err := repo.WriteUserCreationEvent(ctx, tx, event)
		require.ErrorIs(t, err, wantErr)
	})
}

func TestRepository_ReadUserCreationEventsForUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &Repository{}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		first := domain.UserCreationEvent{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			CreatedAt: time.Date(2026, time.March, 20, 10, 0, 0, 0, time.UTC),
		}
		second := domain.UserCreationEvent{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			CreatedAt: time.Date(2026, time.March, 20, 11, 0, 0, 0, time.UTC),
		}

		rows := newStubRows(
			[]string{"id", "user_id", "created_at"},
			[][]any{
				{first.ID, first.UserID, first.CreatedAt},
				{second.ID, second.UserID, second.CreatedAt},
			},
		)
		tx := &stubTx{queryRows: rows}

		got, err := repo.ReadUserCreationEventsForUpdate(ctx, tx, 2)
		require.NoError(t, err)

		want := []domain.UserCreationEvent{first, second}
		require.Equal(t, want, got)

		require.Equal(t, normalizeSQL(`
			SELECT id, user_id, created_at FROM user_creation_outbox
			ORDER BY created_at, id LIMIT $1
			FOR UPDATE SKIP LOCKED`), normalizeSQL(tx.querySQL))

		wantArgs := []any{2}
		require.Equal(t, wantArgs, tx.queryArgs)

		require.True(t, rows.closed)
	})

	t.Run("query error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("query failed")
		tx := &stubTx{queryErr: wantErr}

		got, err := repo.ReadUserCreationEventsForUpdate(ctx, tx, 3)
		require.ErrorIs(t, err, wantErr)
		require.Nil(t, got)
	})

	t.Run("scan error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("scan failed")
		rows := newStubRows(
			[]string{"id", "user_id", "created_at"},
			[][]any{{uuid.New(), uuid.New(), time.Now().UTC()}},
		)
		rows.scanErr = wantErr
		tx := &stubTx{queryRows: rows}

		got, err := repo.ReadUserCreationEventsForUpdate(ctx, tx, 1)
		require.ErrorIs(t, err, wantErr)
		require.Nil(t, got)
		require.True(t, rows.closed)
	})
}

func TestRepository_DeleteUserCreationEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &Repository{}
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tx := &stubTx{execTag: pgconn.NewCommandTag("DELETE 1")}

		err := repo.DeleteUserCreationEvent(ctx, tx, id)
		require.NoError(t, err)

		require.Equal(t, normalizeSQL(`
			DELETE FROM user_creation_outbox
			WHERE id = $1`), normalizeSQL(tx.execSQL))

		wantArgs := []any{id}
		require.Equal(t, wantArgs, tx.execArgs)
	})

	t.Run("exec error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("delete failed")
		tx := &stubTx{execErr: wantErr}

		err := repo.DeleteUserCreationEvent(ctx, tx, id)
		require.ErrorIs(t, err, wantErr)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		tx := &stubTx{execTag: pgconn.NewCommandTag("DELETE 0")}

		err := repo.DeleteUserCreationEvent(ctx, tx, id)
		require.ErrorIs(t, err, ErrUserCreationEventNotFound)
	})
}

type stubTx struct {
	execTag  pgconn.CommandTag
	execErr  error
	execSQL  string
	execArgs []any

	queryRows pgx.Rows
	queryErr  error
	querySQL  string
	queryArgs []any
}

func (s *stubTx) Begin(context.Context) (pgx.Tx, error) {
	return nil, errors.New("unexpected Begin call")
}
func (s *stubTx) Commit(context.Context) error   { return errors.New("unexpected Commit call") }
func (s *stubTx) Rollback(context.Context) error { return errors.New("unexpected Rollback call") }
func (s *stubTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("unexpected CopyFrom call")
}
func (s *stubTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	return nil
}
func (s *stubTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (s *stubTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("unexpected Prepare call")
}
func (s *stubTx) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	s.execSQL = sql
	s.execArgs = args
	return s.execTag, s.execErr
}
func (s *stubTx) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	s.querySQL = sql
	s.queryArgs = args
	return s.queryRows, s.queryErr
}
func (s *stubTx) QueryRow(context.Context, string, ...any) pgx.Row { return nil }
func (s *stubTx) Conn() *pgx.Conn                                  { return nil }

type stubRows struct {
	descs   []pgconn.FieldDescription
	values  [][]any
	index   int
	closed  bool
	scanErr error
	rowsErr error
}

func newStubRows(columns []string, values [][]any) *stubRows {
	descs := make([]pgconn.FieldDescription, len(columns))
	for i, column := range columns {
		descs[i] = pgconn.FieldDescription{Name: column}
	}

	return &stubRows{
		descs:  descs,
		values: values,
	}
}

func (s *stubRows) Close() { s.closed = true }

func (s *stubRows) Err() error { return s.rowsErr }

func (s *stubRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

func (s *stubRows) FieldDescriptions() []pgconn.FieldDescription { return s.descs }

func (s *stubRows) Next() bool {
	if s.index >= len(s.values) {
		s.closed = true
		return false
	}

	s.index++
	return true
}

func (s *stubRows) Scan(dest ...any) error {
	if s.scanErr != nil {
		s.closed = true
		return s.scanErr
	}

	row := s.values[s.index-1]
	for i := range dest {
		target := reflect.ValueOf(dest[i])
		if target.Kind() != reflect.Pointer || target.IsNil() {
			return errors.New("destination must be a non-nil pointer")
		}

		value := reflect.ValueOf(row[i])
		target.Elem().Set(value)
	}

	return nil
}

func (s *stubRows) Values() ([]any, error) {
	if s.index == 0 || s.index > len(s.values) {
		return nil, errors.New("no current row")
	}
	return s.values[s.index-1], nil
}

func (s *stubRows) RawValues() [][]byte { return nil }

func (s *stubRows) Conn() *pgx.Conn { return nil }

func normalizeSQL(sql string) string {
	var out []rune
	space := false

	for _, r := range sql {
		if r == ' ' || r == '\n' || r == '\t' {
			if !space {
				out = append(out, ' ')
				space = true
			}
			continue
		}

		out = append(out, r)
		space = false
	}

	return string(out)
}
