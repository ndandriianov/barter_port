package db

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

func RunInTx(ctx context.Context, db *pgxpool.Pool, fn func(ctx context.Context, tx pgx.Tx) error) (err error) {
	return RunInTxWithOpts(ctx, db, pgx.TxOptions{}, fn)
}

func RunInTxWithOpts(
	ctx context.Context,
	db *pgxpool.Pool,
	opts pgx.TxOptions,
	fn func(ctx context.Context, tx pgx.Tx) error,
) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-throw panic after Rollback
		}
	}()

	err = fn(ctx, tx)
	if err == nil {
		return tx.Commit(ctx)
	}

	return errors.Join(err, tx.Rollback(ctx))
}

type DB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
