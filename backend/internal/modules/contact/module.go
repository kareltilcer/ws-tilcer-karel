package contact

import (
	"embed"
	"io/fs"

	"github.com/go-chi/chi/v5"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/ratelimit"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// Module is the contact feature module: a public, anti-abuse-guarded submit that
// stores + emails the message, plus the admin inbox.
type Module struct {
	handler *Handler
}

// NewModule builds the contact module over svc. limiter rate-limits public
// submits per IP.
func NewModule(svc *Service, limiter *ratelimit.Limiter, maxBytes int64) *Module {
	return &Module{handler: NewHandler(svc, limiter, maxBytes)}
}

func (m *Module) Name() string { return "contact" }

func (m *Module) RegisterPublic(r chi.Router) { m.handler.MountPublic(r) }

func (m *Module) RegisterAdmin(r chi.Router) { m.handler.MountAdmin(r) }

func (m *Module) Migrations() fs.FS { return MigrationsFS }
