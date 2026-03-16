package repox

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

func FetchStructs[T any](
	ctx context.Context,
	db *pgxpool.Pool,
	query string,
	args ...interface{},
) ([]T, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	structs, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []T{}, nil
		}
		return nil, err
	}

	return structs, nil
}
