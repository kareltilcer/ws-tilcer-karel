// Package testsupport provides a temp SQLite database wired with a given set of
// migration sources, for module integration tests. It deliberately does NOT
// import bootstrap (which imports the feature modules) so a module's white-box
// test can use it without an import cycle — each test passes the migration
// sources it needs (platform + the module under test).
package testsupport

import (
	"path/filepath"
	"testing"

	"database/sql"

	appdb "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/db"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/registry"
)

// NewDB opens a fresh temp SQLite database and applies the merged migrations
// from sources. The DB is closed automatically at test end.
func NewDB(t testing.TB, sources ...registry.MigrationSource) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := appdb.Open(path)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	fsys, err := registry.MergeMigrations(sources)
	if err != nil {
		t.Fatalf("merge migrations: %v", err)
	}
	if err := appdb.Migrate(db, fsys); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
