package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/reqctx"
)

// bearerAuthCtxKey marks a request as authenticated via Authorization: Bearer,
// so the CSRF middleware (which runs after session auth) can skip it.
type bearerAuthCtxKey struct{}

func withBearerAuth(ctx context.Context) context.Context {
	return context.WithValue(ctx, bearerAuthCtxKey{}, true)
}

func isBearerAuth(ctx context.Context) bool {
	v, _ := ctx.Value(bearerAuthCtxKey{}).(bool)
	return v
}

// bearerToken returns the token from an `Authorization: Bearer <token>` header,
// or "" when absent/malformed.
func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return ""
}

// Cookie names (host-only on karel.tilcer.cz).
const (
	cookieSession = "session"
	cookieCSRF    = "csrf"
)

// touchGranularity bounds how often the middleware slides a session's expiry.
const touchGranularity = time.Hour

// Config configures the session middleware (and is reused by the /api/auth
// handler). Under the dev bypass, BypassActor is set and Sessions/Authr are nil:
// every admin request is authenticated as that actor (the app runs offline).
type Config struct {
	Sessions    *SessionStore
	Authr       Authenticator
	RoleRefresh time.Duration
	SessionTTL  time.Duration
	Secure      bool
	Origins     []string
	BypassActor *reqctx.Actor
	Now         func() time.Time
	Logger      *slog.Logger
}

func (c Config) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func (c Config) logger() *slog.Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return slog.Default()
}

// NewSessionAuth returns middleware that authorizes every request in the admin
// group from the karel session cookie, refreshing roles past the threshold and
// failing closed when the user is disabled/deleted in auth. Under the dev bypass
// it injects a fixed actor and skips the session entirely.
func NewSessionAuth(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.BypassActor != nil {
				next.ServeHTTP(w, r.WithContext(reqctx.WithActor(r.Context(), *cfg.BypassActor)))
				return
			}

			// Bearer path: an admin client following the OpenAPI contract may send a
			// short-lived site-scoped JWT (Authorization: Bearer) instead of the
			// session cookie. Verify it directly — no session lookup, no role re-mint
			// (the token itself is fresh and short-lived). A present-but-invalid token
			// fails closed rather than silently falling back to the cookie.
			if raw := bearerToken(r); raw != "" {
				if cfg.Authr == nil {
					httpx.WriteError(w, httpx.ErrUnauthorized("bearer auth unavailable"))
					return
				}
				id, err := cfg.Authr.VerifyAccessToken(raw)
				if err != nil {
					httpx.WriteError(w, httpx.ErrUnauthorized("invalid or expired token"))
					return
				}
				actor := reqctx.Actor{
					UserID: id.UserID,
					Type:   "user",
					Label:  labelForIdentity(id),
					Roles:  id.Roles,
				}
				ctx := withBearerAuth(reqctx.WithActor(r.Context(), actor))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			c, err := r.Cookie(cookieSession)
			if err != nil || c.Value == "" {
				httpx.WriteError(w, httpx.ErrUnauthorized("no session"))
				return
			}
			now := cfg.now()
			sess, ok, err := cfg.Sessions.Lookup(r.Context(), c.Value, now)
			if err != nil {
				httpx.WriteError(w, httpx.ErrInternal(""))
				return
			}
			if !ok {
				clearAuthCookies(w, cfg.Secure)
				httpx.WriteError(w, httpx.ErrUnauthorized("invalid or expired session"))
				return
			}

			roles := sess.Roles
			if cfg.Authr != nil && now.Sub(sess.RolesRefreshedAt) > cfg.RoleRefresh {
				id, mintErr := cfg.Authr.Mint(r.Context(), sess.UserID)
				switch {
				case errors.Is(mintErr, ErrUserClosed):
					_ = cfg.Sessions.RevokeByID(r.Context(), sess.ID)
					clearAuthCookies(w, cfg.Secure)
					httpx.WriteError(w, httpx.ErrUnauthorized("session revoked"))
					return
				case mintErr == nil:
					roles = id.Roles
					_ = cfg.Sessions.RefreshRoles(r.Context(), sess.ID, roles, now)
				default:
					cfg.logger().Warn("role re-mint failed (transient)", "user", sess.UserID, "err", mintErr)
				}
			}

			// Slide the expiry lazily and re-issue the cookies with a fresh lifetime.
			if now.Sub(sess.LastSeenAt) > touchGranularity {
				if err := cfg.Sessions.Touch(r.Context(), sess.ID, now, now.Add(cfg.SessionTTL)); err == nil {
					setSessionCookie(w, c.Value, cfg.SessionTTL, cfg.Secure)
					if csrf, err := r.Cookie(cookieCSRF); err == nil && csrf.Value != "" {
						setCSRFCookie(w, csrf.Value, cfg.SessionTTL, cfg.Secure)
					}
				}
			}

			actor := reqctx.Actor{
				UserID: sess.UserID,
				Type:   "user",
				Label:  labelFor(sess),
				Roles:  roles,
			}
			next.ServeHTTP(w, r.WithContext(reqctx.WithActor(r.Context(), actor)))
		})
	}
}

// NewCSRF returns double-submit CSRF middleware: every cookie-authenticated
// state-changing request must carry an X-CSRF-Token header equal to the csrf
// cookie AND an Origin/Referer within the allowlist. Safe methods pass through.
// Under the dev bypass it is a no-op.
func NewCSRF(origins []string, bypass bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bearer-authenticated requests carry a non-ambient credential the browser
			// never attaches automatically, so CSRF (double-submit) does not apply.
			if bypass || safeMethod(r.Method) || isBearerAuth(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}
			if !originAllowed(r, origins) {
				httpx.WriteError(w, httpx.ErrForbidden("origin not allowed"))
				return
			}
			header := r.Header.Get("X-CSRF-Token")
			cookie, err := r.Cookie(cookieCSRF)
			if header == "" || err != nil || cookie.Value == "" || header != cookie.Value {
				httpx.WriteError(w, httpx.ErrForbidden("csrf token mismatch"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func safeMethod(m string) bool {
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

func originAllowed(r *http.Request, allowed []string) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		if ref := r.Header.Get("Referer"); ref != "" {
			if u, err := url.Parse(ref); err == nil {
				origin = u.Scheme + "://" + u.Host
			}
		}
	}
	if origin == "" {
		return false
	}
	for _, a := range allowed {
		if originMatches(origin, a) {
			return true
		}
	}
	return false
}

func originMatches(origin, pattern string) bool {
	if origin == pattern {
		return true
	}
	scheme, host, ok := splitOrigin(origin)
	pScheme, pHost, ok2 := splitOrigin(pattern)
	if !ok || !ok2 || scheme != pScheme {
		return false
	}
	if strings.HasPrefix(pHost, "*.") {
		suffix := pHost[1:] // ".tilcer.cz"
		return strings.HasSuffix(host, suffix) && host != suffix[1:]
	}
	return false
}

func splitOrigin(o string) (scheme, host string, ok bool) {
	i := strings.Index(o, "://")
	if i < 0 {
		return "", "", false
	}
	return o[:i], o[i+3:], true
}

func labelFor(s Session) string {
	if s.DisplayName != "" {
		return s.DisplayName
	}
	return s.Email
}

func labelForIdentity(id Identity) string {
	if id.DisplayName != "" {
		return id.DisplayName
	}
	return id.Email
}

func setAuthCookies(w http.ResponseWriter, sessionToken, csrfToken string, ttl time.Duration, secure bool) {
	setSessionCookie(w, sessionToken, ttl, secure)
	setCSRFCookie(w, csrfToken, ttl, secure)
}

func setSessionCookie(w http.ResponseWriter, token string, ttl time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name: cookieSession, Value: token, Path: "/",
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: int(ttl.Seconds()),
	})
}

func setCSRFCookie(w http.ResponseWriter, token string, ttl time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name: cookieCSRF, Value: token, Path: "/",
		HttpOnly: false, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: int(ttl.Seconds()),
	})
}

func clearAuthCookies(w http.ResponseWriter, secure bool) {
	for _, name := range []string{cookieSession, cookieCSRF} {
		http.SetCookie(w, &http.Cookie{
			Name: name, Value: "", Path: "/",
			HttpOnly: name == cookieSession, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: -1,
		})
	}
}

func newCSRFToken() (string, error) {
	raw, _, err := newToken()
	return raw, err
}
