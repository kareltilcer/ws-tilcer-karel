package httpx

import (
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/idgen"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/reqctx"
)

const requestHeaderID = "X-Request-Id"

// statusRecorder captures the response status for the access log.
type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wrote {
		r.status = code
		r.wrote = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wrote {
		r.status = http.StatusOK
		r.wrote = true
	}
	return r.ResponseWriter.Write(b)
}

// RequestID mints (or accepts) a request id, stores request metadata in the
// context, and echoes the id on the response.
func RequestID(site string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(requestHeaderID)
			if id == "" {
				id = idgen.New()
			}
			w.Header().Set(requestHeaderID, id)
			info := reqctx.RequestInfo{
				RequestID: id,
				IP:        clientIP(r),
				UserAgent: r.UserAgent(),
				Site:      site,
			}
			next.ServeHTTP(w, r.WithContext(reqctx.WithRequest(r.Context(), info)))
		})
	}
}

// Logger emits one structured access log line per request carrying the request id.
func Logger(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			id := ""
			if info, ok := reqctx.RequestFrom(r.Context()); ok {
				id = info.RequestID
			}
			l.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"latency_ms", time.Since(start).Milliseconds(),
				"request_id", id,
			)
		})
	}
}

// Recover converts a panic into a 500 and logs it with the request id.
func Recover(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					id := ""
					if info, ok := reqctx.RequestFrom(r.Context()); ok {
						id = info.RequestID
					}
					l.Error("panic recovered", "panic", p, "path", r.URL.Path, "request_id", id)
					WriteError(w, ErrInternal(""))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts a best-effort client IP, honouring X-Forwarded-For (Coolify
// terminates TLS and proxies). Exactly one trusted reverse proxy sits in front
// and APPENDS the connecting peer to X-Forwarded-For, so the RIGHTMOST entry is
// the address our proxy actually saw. The leftmost entries are client-supplied
// and forgeable — using them would let an attacker rotate a fake header to evade
// every per-IP limiter (including the login brute-force lockout).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		last := strings.TrimSpace(parts[len(parts)-1])
		if last != "" {
			return last
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ClientIP returns the request's resolved client IP: the value computed by the
// RequestID middleware (X-Forwarded-For aware) when present, else the raw
// RemoteAddr. Handlers use this as the single source of truth for per-IP keys.
func ClientIP(r *http.Request) string {
	if info, ok := reqctx.RequestFrom(r.Context()); ok && info.IP != "" {
		return info.IP
	}
	return r.RemoteAddr
}

// SetRetryAfter sets the Retry-After header to ceil(d) seconds (minimum 1),
// for use alongside a 429 response.
func SetRetryAfter(w http.ResponseWriter, d time.Duration) {
	secs := int(math.Ceil(d.Seconds()))
	if secs < 1 {
		secs = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(secs))
}
