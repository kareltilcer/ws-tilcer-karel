package content

import (
	"embed"
	"io/fs"

	"github.com/go-chi/chi/v5"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/ratelimit"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// Module is the content feature module: projects, categories, links (+click),
// page sections, skills, and the business-info disclosure. It owns the content +
// media tables (the media module reuses the media table for uploads).
type Module struct {
	handler *Handler
}

// NewModule builds the content module over svc. clickLimiter rate-limits the
// public link-click counter.
func NewModule(svc *Service, clickLimiter *ratelimit.Limiter) *Module {
	return &Module{handler: NewHandler(svc, clickLimiter)}
}

func (m *Module) Name() string { return "content" }

func (m *Module) RegisterPublic(r chi.Router) { m.handler.MountPublic(r) }

func (m *Module) RegisterAdmin(r chi.Router) { m.handler.MountAdmin(r) }

func (m *Module) Migrations() fs.FS { return MigrationsFS }
