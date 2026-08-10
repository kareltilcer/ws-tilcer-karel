package arcade

import (
	"embed"
	"io/fs"

	"github.com/go-chi/chi/v5"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/ratelimit"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// Module is the arcade feature module: code-defined game roster, public
// leaderboard + spam-guarded score submit, and admin moderation.
type Module struct {
	handler *Handler
}

// NewModule builds the arcade module over svc. limiter rate-limits score submits.
func NewModule(svc *Service, limiter *ratelimit.Limiter) *Module {
	return &Module{handler: NewHandler(svc, limiter)}
}

func (m *Module) Name() string { return "arcade" }

func (m *Module) RegisterPublic(r chi.Router) { m.handler.MountPublic(r) }

func (m *Module) RegisterAdmin(r chi.Router) { m.handler.MountAdmin(r) }

func (m *Module) Migrations() fs.FS { return MigrationsFS }
