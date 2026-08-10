package db

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

// Migrate applies all pending migrations from fsys (the merged per-module
// sequence assembled by the registry). Safe to run on every boot: Goose records
// applied versions, and a database restored from Litestream already carries the
// version table, so nothing re-runs.
func Migrate(sqldb *sql.DB, fsys fs.FS) error {
	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	goose.SetLogger(goose.NopLogger())
	if err := goose.Up(sqldb, "."); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
