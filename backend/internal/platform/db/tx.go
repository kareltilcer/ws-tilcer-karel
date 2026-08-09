package db

import (
	"context"
	"database/sql"
)

// WithTx runs fn inside a single transaction, committing on success and rolling
// back on any error or panic. Every multi-statement mutation (reorder, delete
// with reference guards, create-with-joins) flows through here so it is atomic.
func WithTx(ctx context.Context, sqldb *sql.DB, fn func(*sql.Tx) error) (err error) {
	tx, err := sqldb.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback() // no-op if already committed
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	err = tx.Commit()
	return err
}
