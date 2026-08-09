// Package auth implements Mode B (mirrors ws-tilcer-home): karel hosts its own
// login and owns its own SESSION for the CMS admin. The browser carries NO
// bearer token — every admin request is authorized from the session cookie, and
// roles are refreshed periodically by re-minting against the shared auth service
// BE→BE. This package owns the auth service client (/internal/login,
// /internal/token/mint), the session store, the session + CSRF middleware, and
// the /api/auth endpoints.
//
// Everything the network touches is behind the Authenticator interface, so tests
// run against a fake with no network.
package auth

import (
	"context"
	"errors"
)

// Identity is a user's site-scoped assertion as returned by the auth service.
type Identity struct {
	UserID      string   // auth subject id (sub)
	Email       string   //
	DisplayName string   // may be empty
	Roles       []string // e.g. ["admin"] or ["*"] (superuser)
}

// Authenticator verifies credentials and refreshes roles against the shared auth
// service, BE→BE, as the `karel` service client (Mode B).
type Authenticator interface {
	// Login verifies email+password, returning the site-scoped Identity or a typed error.
	Login(ctx context.Context, email, password string) (Identity, error)
	// Mint refreshes a user's identity/roles. Returns ErrUserClosed when the user
	// is disabled/deleted/unverified in auth (the fail-closed case).
	Mint(ctx context.Context, userID string) (Identity, error)
	// VerifyAccessToken cryptographically verifies a site-scoped access token (the
	// short-lived admin JWT sent as `Authorization: Bearer`) and returns the
	// identity carried in its claims, or an error if the token is invalid/expired.
	VerifyAccessToken(raw string) (Identity, error)
}

// Typed login/mint outcomes; the HTTP handler maps each to a status.
var (
	ErrBadCredentials = errors.New("auth: bad credentials")
	ErrDisabled       = errors.New("auth: account disabled or no access to site")
	ErrMFARequired    = errors.New("auth: mfa required")
	ErrUserClosed     = errors.New("auth: user closed")
	ErrUnreachable    = errors.New("auth: service unreachable")
)
