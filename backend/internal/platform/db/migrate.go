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
//
// WithAllowMissing is required by the per-module band scheme. Each module owns a
// numeric prefix band (platform 01, content 02, contact 03, arcade 04), so a new
// migration added to an earlier band — e.g. content's 02002 — has a version below
// a later band already applied in production (arcade's 04001). Goose's default
// guard rejects any un-applied migration ordered before the current DB max
// version ("found N missing migrations before current version"); allowing missing
// migrations applies each un-applied file exactly once regardless of global order,
// which is the intended behaviour here. Fresh DBs (tests) apply everything
// ascending in one pass and never hit the guard, so this only affects upgrades.
func Migrate(sqldb *sql.DB, fsys fs.FS) error {
	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	goose.SetLogger(goose.NopLogger())
	if err := goose.Up(sqldb, ".", goose.WithAllowMissing()); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
