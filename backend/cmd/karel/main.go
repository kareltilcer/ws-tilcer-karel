// Command karel is the entrypoint for the `karel` personal-site + CMS backend: it
// loads configuration, opens the embedded SQLite database, runs migrations, and
// serves the public JSON API and the authenticated CMS admin API (Mode B). The
// public SPA ships as its own Nginx image; in production this backend serves no
// static assets (KAREL_STATIC_DIR unset).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/bootstrap"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/modules/arcade"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/modules/contact"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/modules/content"
	mediamod "github.com/kareltilcer/ws-tilcer-karel/backend/internal/modules/media"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/auth"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/config"
	appdb "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/db"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/email"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	objstore "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/media"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/ratelimit"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/registry"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/reqctx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/spam"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// 1. Configuration — fail fast and loud.
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return err
	}
	logger.Info("config loaded", "config", cfg.Redacted())
	if cfg.DevAuthBypass {
		logger.Warn("AUTH BYPASS ACTIVE — ALL ADMIN REQUESTS ARE FAKE-AUTHENTICATED — DO NOT DEPLOY",
			"dev_actor", cfg.DevActorID, "dev_roles", cfg.DevActorRoles)
	}

	// 2. Database: open → migrate.
	sqldb, err := appdb.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer func() { _ = sqldb.Close() }()

	migFS, err := bootstrap.MigrationFS()
	if err != nil {
		return err
	}
	if err := appdb.Migrate(sqldb, migFS); err != nil {
		return err
	}
	logger.Info("database ready", "path", cfg.DBPath)

	// 3. Mode B auth + session. Under the dev bypass a fixed actor is injected and
	// no auth service / session store is used, so the app runs offline.
	authConf := auth.Config{
		RoleRefresh: time.Duration(cfg.RoleRefreshMinutes) * time.Minute,
		SessionTTL:  time.Duration(cfg.SessionTTLDays) * 24 * time.Hour,
		Secure:      cfg.IsProduction(),
		Origins:     cfg.AllowedOrigins,
		Logger:      logger,
	}
	if cfg.DevAuthBypass {
		authConf.BypassActor = &reqctx.Actor{
			UserID: cfg.DevActorID,
			Type:   "user",
			Label:  cfg.DevActorID,
			Roles:  cfg.DevActorRoles,
		}
	} else {
		authConf.Sessions = auth.NewSessionStore(sqldb)
		authConf.Authr = auth.NewHTTPAuthenticator(cfg.AuthBaseURL, cfg.AuthServiceSecret, cfg.AuthJWTSecret, cfg.AuthJWTIssuer, cfg.SiteKey, logger)
	}
	authHandler := auth.NewHandler(authConf, sqldb)
	sessionMW := auth.NewSessionAuth(authConf)
	csrfMW := auth.NewCSRF(cfg.AllowedOrigins, cfg.DevAuthBypass)

	// 4. Feature modules (content + media now; contact, arcade in M3–M4).
	clickLimiter := ratelimit.New(cfg.ClickRate.Max, cfg.ClickRate.Window)
	contentMod := content.NewModule(content.NewService(sqldb), clickLimiter)

	var s3 objstore.Store
	if cfg.MediaConfigured() {
		s3 = objstore.NewS3Store(objstore.S3Config{
			Endpoint:       cfg.S3Endpoint,
			Region:         cfg.S3Region,
			Bucket:         cfg.S3Bucket,
			AccessKey:      cfg.S3AccessKey,
			SecretKey:      cfg.S3SecretKey,
			ForcePathStyle: cfg.S3ForcePathStyle,
		})
	} else {
		logger.Warn("media storage not configured — uploads will return 503 (set S3_* + MEDIA_PUBLIC_BASE_URL)")
	}
	mediaSvc := mediamod.NewService(sqldb, s3, cfg.S3MediaPrefix, cfg.MediaPublicBaseURL, cfg.MaxMediaBytes)
	mediaMod := mediamod.NewModule(mediaSvc, cfg.MaxMediaBytes)

	// Shared spam guard (honeypot + Turnstile) for the public write endpoints.
	guard := spam.NewGuard(cfg.TurnstileSecret)
	if !guard.Enabled() {
		logger.Warn("Turnstile verification disabled (TURNSTILE_SECRET unset) — honeypot still enforced")
	}
	var mailer email.Sender
	if cfg.EmailConfigured() {
		mailer = email.NewResendClient(cfg.ResendAPIKey)
	} else {
		logger.Warn("contact email not configured — submissions are stored but not emailed (set RESEND_API_KEY/CONTACT_TO/CONTACT_FROM)")
	}
	contactLimiter := ratelimit.New(cfg.ContactRate.Max, cfg.ContactRate.Window)
	contactSvc := contact.NewService(sqldb, guard, mailer, cfg.ContactTo, cfg.ContactFrom, logger)
	contactMod := contact.NewModule(contactSvc, contactLimiter, cfg.MaxContactBytes)

	scoreLimiter := ratelimit.New(cfg.ScoreRate.Max, cfg.ScoreRate.Window)
	arcadeMod := arcade.NewModule(arcade.NewService(sqldb, guard), scoreLimiter)

	modules := []registry.Module{contentMod, mediaMod, contactMod, arcadeMod}

	// 5. HTTP server.
	handler := httpx.NewRouter(httpx.Deps{
		Logger:       logger,
		DB:           sqldb,
		Site:         cfg.SiteKey,
		InsecureAuth: cfg.DevAuthBypass,
		MountAuth:    func(api chi.Router) { authHandler.Mount(api, csrfMW) },
		MountPublicAPI: func(api chi.Router) {
			registry.MountPublic(api, modules)
		},
		SessionMW: sessionMW,
		CSRFMW:    csrfMW,
		MountAdminAPI: func(admin chi.Router) {
			admin.Get("/me", adminMe)
			registry.MountAdmin(admin, modules)
		},
		StaticDir: cfg.StaticDir,
	})
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 6. Serve until interrupted, then shut down gracefully.
	serveErr := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-serveErr:
		return err
	case <-stop:
		logger.Info("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}

// adminMe returns the current authenticated admin (proves the admin gate and
// backs the CMS "who am I" check). Mounted inside the session-gated group.
func adminMe(w http.ResponseWriter, r *http.Request) {
	actor, ok := reqctx.ActorFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.ErrUnauthorized("authentication required"))
		return
	}
	roles := actor.Roles
	if roles == nil {
		roles = []string{}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"user_id": actor.UserID,
		"label":   actor.Label,
		"roles":   roles,
	})
}
