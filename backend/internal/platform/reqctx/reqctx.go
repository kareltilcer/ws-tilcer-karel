// Package reqctx carries request-scoped identity and metadata through the
// context — the authenticated actor (for admin routes) and operational request
// info (request id, client IP, site) used by logging and the anti-abuse guards.
package reqctx

import "context"

// Actor is the authenticated principal behind an admin request. karel is a
// single-admin CMS, so any valid session is the admin; roles are still carried
// (from the shared auth service) for parity with the fleet.
type Actor struct {
	UserID string   // auth subject id ("" for system)
	Type   string   // "user" | "system"
	Label  string   // human-readable label
	Roles  []string // e.g. ["admin"] or ["*"] (superuser)
}

// RequestInfo is the per-request operational metadata (client IP feeds the
// per-IP rate limiters; RequestID ties log lines together).
type RequestInfo struct {
	RequestID string
	IP        string
	UserAgent string
	Site      string
}

type ctxKey int

const (
	actorKey ctxKey = iota
	requestKey
)

// WithActor returns a context carrying a.
func WithActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, actorKey, a)
}

// ActorFrom returns the actor stored in ctx, if any.
func ActorFrom(ctx context.Context) (Actor, bool) {
	a, ok := ctx.Value(actorKey).(Actor)
	return a, ok
}

// WithRequest returns a context carrying r.
func WithRequest(ctx context.Context, r RequestInfo) context.Context {
	return context.WithValue(ctx, requestKey, r)
}

// RequestFrom returns the request info stored in ctx, if any.
func RequestFrom(ctx context.Context) (RequestInfo, bool) {
	r, ok := ctx.Value(requestKey).(RequestInfo)
	return r, ok
}

// HasRole reports whether roles grants any of allowed. The superuser token "*"
// always grants access.
func HasRole(roles []string, allowed ...string) bool {
	for _, r := range roles {
		if r == "*" {
			return true
		}
		for _, a := range allowed {
			if r == a {
				return true
			}
		}
	}
	return false
}
