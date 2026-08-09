// Package bootstrap composes the compile-time modular monolith. It is the single
// place that assembles the one Goose migration sequence from the platform core
// and each feature module's per-module files. Feature modules are wired in
// cmd/karel/main.go (which owns their dependencies); bootstrap owns only the
// migration assembly so the server and any test harness agree on the schema.
//
// bootstrap sits above modules (it imports them for their embedded migrations);
// modules never import bootstrap, so there is no cycle.
package bootstrap

import (
	"io/fs"

	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/modules/arcade"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/modules/contact"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/modules/content"
	appdb "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/db"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/registry"
)

// MigrationSources returns every migration contributor in prefix order. Goose
// applies migrations globally by numeric filename prefix, so the effective order
// is platform(01) → content(02) → contact(03) → arcade(04). Feature modules are
// appended here as they land (M1–M4).
func MigrationSources() []registry.MigrationSource {
	return []registry.MigrationSource{
		{Name: "platform", FS: appdb.MigrationsFS},
		{Name: "content", FS: content.MigrationsFS},
		{Name: "contact", FS: contact.MigrationsFS},
		{Name: "arcade", FS: arcade.MigrationsFS},
	}
}

// MigrationFS assembles the merged, boot-time migration FS for the whole service.
func MigrationFS() (fs.FS, error) {
	return registry.MergeMigrations(MigrationSources())
}
