package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	appdb "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/db"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/reqctx"
)

// Login rate-limit: per IP and per (IP,email).
const (
	loginMaxAttempts = 10
	loginWindow      = 15 * time.Minute
)

// Handler serves the karel-hosted /api/auth endpoints (Mode B). karel has no
// audit spine (unlike home), so session create/revoke run in a plain tx.
type Handler struct {
	cfg     Config
	db      *sql.DB
	limiter *loginLimiter
}

// NewHandler builds the auth handler.
func NewHandler(cfg Config, db *sql.DB) *Handler {
	return &Handler{cfg: cfg, db: db, limiter: newLoginLimiter(loginMaxAttempts, loginWindow, cfg.Now)}
}

// Mount registers the auth routes on the /api router (OUTSIDE the admin gate —
// login must work before there is a session). csrf is applied to logout only.
func (h *Handler) Mount(api chi.Router, csrf func(http.Handler) http.Handler) {
	api.Post("/auth/login", h.login)
	api.Get("/auth/session", h.session)
	api.With(csrf).Post("/auth/logout", h.logout)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userPublic struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	DisplayName *string  `json:"display_name"`
	Roles       []string `json:"roles"`
}

type sessionUser struct {
	User userPublic `json:"user"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.WriteError(w, httpx.ErrUnprocessable(err.Error()))
		return
	}

	// Dev bypass: no session store or auth service — accept and echo the actor.
	if h.cfg.BypassActor != nil {
		a := h.cfg.BypassActor
		id := Identity{UserID: a.UserID, Email: firstNonEmpty(in.Email, a.Label, a.UserID), DisplayName: a.Label, Roles: a.Roles}
		httpx.JSON(w, http.StatusOK, sessionUser{User: publicUser(id)})
		return
	}

	ip := httpx.ClientIP(r)
	ipKey := "ip:" + ip
	acctKey := ipKey + "|email:" + normalizeEmail(in.Email)
	if !h.limiter.allowed(ipKey) || !h.limiter.allowed(acctKey) {
		httpx.WriteError(w, httpx.ErrTooManyRequests("příliš mnoho pokusů"))
		return
	}

	id, err := h.cfg.Authr.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		if errors.Is(err, ErrBadCredentials) {
			h.limiter.fail(ipKey)
			h.limiter.fail(acctKey)
		}
		writeLoginError(w, err)
		return
	}
	h.limiter.reset(ipKey)
	h.limiter.reset(acctKey)

	now := h.cfg.now()
	req, _ := reqctx.RequestFrom(r.Context())

	var rawToken string
	if err := appdb.WithTx(r.Context(), h.db, func(tx *sql.Tx) error {
		raw, _, err := h.cfg.Sessions.Create(r.Context(), tx, id, req.UserAgent, ip, h.cfg.SessionTTL, now)
		if err != nil {
			return err
		}
		rawToken = raw
		return nil
	}); err != nil {
		httpx.WriteError(w, httpx.ErrInternal(""))
		return
	}

	csrfToken, err := newCSRFToken()
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal(""))
		return
	}
	setAuthCookies(w, rawToken, csrfToken, h.cfg.SessionTTL, h.cfg.Secure)
	httpx.JSON(w, http.StatusOK, sessionUser{User: publicUser(id)})
}

func (h *Handler) session(w http.ResponseWriter, r *http.Request) {
	if h.cfg.BypassActor != nil {
		a := h.cfg.BypassActor
		httpx.JSON(w, http.StatusOK, sessionUser{User: publicUser(Identity{
			UserID: a.UserID, Email: a.Label, DisplayName: a.Label, Roles: a.Roles,
		})})
		return
	}
	c, err := r.Cookie(cookieSession)
	if err != nil || c.Value == "" {
		httpx.WriteError(w, httpx.ErrUnauthorized("no session"))
		return
	}
	sess, ok, err := h.cfg.Sessions.Lookup(r.Context(), c.Value, h.cfg.now())
	if err != nil {
		httpx.WriteError(w, httpx.ErrInternal(""))
		return
	}
	if !ok {
		clearAuthCookies(w, h.cfg.Secure)
		httpx.WriteError(w, httpx.ErrUnauthorized("invalid or expired session"))
		return
	}
	httpx.JSON(w, http.StatusOK, sessionUser{User: publicUser(Identity{
		UserID: sess.UserID, Email: sess.Email, DisplayName: sess.DisplayName, Roles: sess.Roles,
	})})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if h.cfg.BypassActor != nil {
		clearAuthCookies(w, h.cfg.Secure)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	c, err := r.Cookie(cookieSession)
	if err != nil || c.Value == "" {
		httpx.WriteError(w, httpx.ErrUnauthorized("no session"))
		return
	}
	now := h.cfg.now()
	if err := appdb.WithTx(r.Context(), h.db, func(tx *sql.Tx) error {
		_, _, _, err := h.cfg.Sessions.RevokeByToken(r.Context(), tx, c.Value, now)
		return err
	}); err != nil {
		httpx.WriteError(w, httpx.ErrInternal(""))
		return
	}
	clearAuthCookies(w, h.cfg.Secure)
	w.WriteHeader(http.StatusNoContent)
}

func writeLoginError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrBadCredentials):
		httpx.WriteError(w, httpx.ErrUnauthorized("neplatné přihlašovací údaje"))
	case errors.Is(err, ErrDisabled):
		httpx.WriteError(w, httpx.ErrForbidden("účet je zablokován nebo nemá přístup"))
	case errors.Is(err, ErrMFARequired):
		httpx.WriteError(w, &httpx.APIError{Status: http.StatusConflict, Code: "mfa_required", Detail: "dokončete přihlášení na auth.tilcer.cz"})
	default:
		httpx.WriteError(w, &httpx.APIError{Status: http.StatusBadGateway, Code: "auth_unreachable", Detail: "ověřovací služba je nedostupná"})
	}
}

func publicUser(id Identity) userPublic {
	var dn *string
	if id.DisplayName != "" {
		d := id.DisplayName
		dn = &d
	}
	roles := id.Roles
	if roles == nil {
		roles = []string{}
	}
	return userPublic{ID: id.UserID, Email: id.Email, DisplayName: dn, Roles: roles}
}

func normalizeEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}
