package httpx

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Deps carries everything the router needs. It grows as modules land without
// churning call sites.
type Deps struct {
	Logger       *slog.Logger
	DB           Pinger
	Site         string
	InsecureAuth bool

	// MountAuth mounts the public /api/auth/* endpoints (login/logout/session).
	// These run OUTSIDE the admin gate — login must work before a session exists.
	MountAuth func(api chi.Router)
	// MountPublicAPI mounts the public read + public write endpoints (content
	// reads, contact submit, arcade leaderboard+submit, link-click). No session;
	// the public writes carry their own honeypot/Turnstile/rate-limit guards.
	MountPublicAPI func(api chi.Router)

	// SessionMW authorizes every request in the /api/admin group from the karel
	// session cookie. May be nil (dev bypass injects a fixed actor instead).
	SessionMW func(http.Handler) http.Handler
	// CSRFMW enforces the double-submit CSRF check on admin mutations. May be nil.
	CSRFMW func(http.Handler) http.Handler
	// MountAdminAPI mounts the authenticated CMS endpoints onto the gated
	// /api/admin group.
	MountAdminAPI func(admin chi.Router)

	// StaticDir, when non-empty, serves the built SPA on non-API routes with an
	// index.html fallback. Unset in the two-app production deploy (the SPA is its
	// own Nginx image).
	StaticDir string
}

// NewRouter assembles the HTTP handler: baseline middleware, public health
// probes, the public auth + public API surface, the session-gated + CSRF
// protected /api/admin surface, and the SPA fallback. Health probes are mounted
// OUTSIDE auth so they stay public and cheap.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(RequestID(d.Site))
	r.Use(Logger(d.Logger))
	r.Use(Recover(d.Logger))

	// Health (public).
	r.Get("/healthz", Healthz())
	r.Get("/readyz", Readyz(d.DB, d.InsecureAuth))

	r.Route("/api", func(api chi.Router) {
		// Public auth endpoints (login is pre-session, Mode B).
		if d.MountAuth != nil {
			d.MountAuth(api)
		}
		// Public read + public write surface (no session; own anti-abuse guards).
		if d.MountPublicAPI != nil {
			api.Group(func(pub chi.Router) {
				d.MountPublicAPI(pub)
			})
		}
		// Admin CMS: session-authorized and, for mutations, CSRF-protected.
		api.Route("/admin", func(admin chi.Router) {
			if d.SessionMW != nil {
				admin.Use(d.SessionMW)
			}
			if d.CSRFMW != nil {
				admin.Use(d.CSRFMW)
			}
			if d.MountAdminAPI != nil {
				d.MountAdminAPI(admin)
			}
		})
	})

	// Catch-all: the built SPA is served on non-API routes with an index.html
	// fallback. An unmatched /api/** path must NOT fall through to the SPA shell —
	// it gets a JSON 404 so a mistyped endpoint fails loudly.
	var spa http.Handler
	if d.StaticDir != "" {
		spa = SPAHandler(d.StaticDir)
	}
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		if spa != nil && !strings.HasPrefix(req.URL.Path, "/api/") {
			spa.ServeHTTP(w, req)
			return
		}
		WriteError(w, ErrNotFound("no such endpoint"))
	})

	return r
}
