// Package registry is the core of the compile-time modular monolith. It defines
// the Module contract every feature module registers through (public routes,
// admin routes, migrations) and the composition helpers the core uses to build
// the router and assemble the one Goose migration sequence.
//
// registry lives in platform/ and is imported BY modules; it must never import a
// module.
package registry

import (
	"fmt"
	"io/fs"
	"path"
	"testing/fstest"

	"github.com/go-chi/chi/v5"
)

// Module is what every feature module (content, media, contact, arcade)
// implements to plug into the core. One binary, one deploy.
type Module interface {
	// Name is the code identifier: "content" | "media" | "contact" | "arcade".
	Name() string
	// RegisterPublic mounts the module's public routes (reads + anti-abuse-guarded
	// public writes) on the unauthenticated /api group.
	RegisterPublic(r chi.Router)
	// RegisterAdmin mounts the module's admin routes on the session-gated,
	// CSRF-protected /api/admin group.
	RegisterAdmin(r chi.Router)
	// Migrations returns this module's Goose *.sql files (embedded, under a
	// "migrations/" prefix). May be empty.
	Migrations() fs.FS
}

// MigrationSource is one contributor of Goose *.sql files: a module or the
// platform core. Files carry a globally-unique numeric prefix so the merged
// sequence is deterministic and stable across releases.
type MigrationSource struct {
	Name string
	FS   fs.FS
}

// MergeMigrations flattens every source's migrations/*.sql into a single FS that
// goose applies in one ascending sequence. Filenames must be globally unique; a
// collision is a bug and fails fast rather than silently dropping a migration.
func MergeMigrations(sources []MigrationSource) (fs.FS, error) {
	merged := fstest.MapFS{}
	for _, src := range sources {
		files, err := fs.Glob(src.FS, "migrations/*.sql")
		if err != nil {
			return nil, fmt.Errorf("registry: glob %s migrations: %w", src.Name, err)
		}
		for _, f := range files {
			name := path.Base(f)
			if _, dup := merged[name]; dup {
				return nil, fmt.Errorf("registry: duplicate migration filename %q (source %s)", name, src.Name)
			}
			b, err := fs.ReadFile(src.FS, f)
			if err != nil {
				return nil, fmt.Errorf("registry: read %s: %w", f, err)
			}
			merged[name] = &fstest.MapFile{Data: b}
		}
	}
	if len(merged) == 0 {
		return nil, fmt.Errorf("registry: no migrations found across %d sources", len(sources))
	}
	return merged, nil
}

// MountPublic registers every module's public routes on the /api group.
func MountPublic(api chi.Router, modules []Module) {
	for _, m := range modules {
		m.RegisterPublic(api)
	}
}

// MountAdmin registers every module's admin routes on the /api/admin group.
func MountAdmin(admin chi.Router, modules []Module) {
	for _, m := range modules {
		m.RegisterAdmin(admin)
	}
}
