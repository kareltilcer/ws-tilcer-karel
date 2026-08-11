package db_test

import (
	"database/sql"
	"path/filepath"
	"testing"
	"testing/fstest"

	appdb "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/db"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/registry"
)

// gooseFile wraps a single CREATE TABLE in the minimal goose Up annotations so it
// can stand in for a real module migration.
func gooseFile(table string) []byte {
	return []byte("-- +goose Up\n-- +goose StatementBegin\nCREATE TABLE " +
		table + " (id INTEGER PRIMARY KEY);\n-- +goose StatementEnd\n")
}

func source(name string, files map[string][]byte) registry.MigrationSource {
	m := fstest.MapFS{}
	for path, data := range files {
		m[path] = &fstest.MapFile{Data: data}
	}
	return registry.MigrationSource{Name: name, FS: m}
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var found string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, name,
	).Scan(&found)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		t.Fatalf("query sqlite_master for %q: %v", name, err)
	}
	return found == name
}

// TestMigrateAppliesLowerBandAfterHigherBand reproduces the production upgrade
// failure fixed by goose.WithAllowMissing. The per-module band scheme means a new
// content migration (02002) can be authored after arcade's 04001 is already
// applied in production, so its version sorts *before* the current DB max version.
// Goose's default guard rejects that as a "missing migration before current
// version"; the fix must apply it anyway. A fresh DB never reproduces the bug (it
// applies every band ascending in one pass), so the test migrates in two phases.
func TestMigrateAppliesLowerBandAfterHigherBand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := appdb.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Phase 1: simulate a production DB already migrated through the arcade band
	// (04001) — content only carries 02001 so far.
	phase1, err := registry.MergeMigrations([]registry.MigrationSource{
		source("content", map[string][]byte{"migrations/02001_content.sql": gooseFile("content_a")}),
		source("arcade", map[string][]byte{"migrations/04001_arcade.sql": gooseFile("arcade_a")}),
	})
	if err != nil {
		t.Fatalf("merge phase 1: %v", err)
	}
	if err := appdb.Migrate(db, phase1); err != nil {
		t.Fatalf("migrate phase 1: %v", err)
	}

	// Phase 2: content ships a new migration (02002) whose version is below the
	// already-applied 04001. This is the exact shape that failed on boot.
	phase2, err := registry.MergeMigrations([]registry.MigrationSource{
		source("content", map[string][]byte{
			"migrations/02001_content.sql": gooseFile("content_a"),
			"migrations/02002_content.sql": gooseFile("content_b"),
		}),
		source("arcade", map[string][]byte{"migrations/04001_arcade.sql": gooseFile("arcade_a")}),
	})
	if err != nil {
		t.Fatalf("merge phase 2: %v", err)
	}
	if err := appdb.Migrate(db, phase2); err != nil {
		t.Fatalf("migrate phase 2 (out-of-order lower band): %v", err)
	}
	if !tableExists(t, db, "content_b") {
		t.Fatal("02002 migration did not apply: table content_b missing")
	}

	// Re-running is a no-op: goose has recorded 02002, so a third pass must not
	// re-apply it or error.
	if err := appdb.Migrate(db, phase2); err != nil {
		t.Fatalf("migrate phase 3 (idempotent re-run): %v", err)
	}
}
