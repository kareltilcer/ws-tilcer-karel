package media

import (
	"io/fs"

	"github.com/go-chi/chi/v5"
)

// Module is the media feature module: S3-backed image upload/list/delete with
// mime+size validation, SVG sanitization, and a reference-guarded delete. It has
// no public routes — objects are served directly from MEDIA_PUBLIC_BASE_URL.
type Module struct {
	handler *Handler
}

// NewModule builds the media module over svc.
func NewModule(svc *Service, maxBytes int64) *Module {
	return &Module{handler: NewHandler(svc, maxBytes)}
}

func (m *Module) Name() string { return "media" }

// RegisterPublic is a no-op: media has no public API surface.
func (m *Module) RegisterPublic(r chi.Router) {}

func (m *Module) RegisterAdmin(r chi.Router) { m.handler.MountAdmin(r) }

// Migrations returns nil: the `media` table is created by the content migration
// (the content tables carry FKs into it), so the media module owns no migration.
func (m *Module) Migrations() fs.FS { return nil }
