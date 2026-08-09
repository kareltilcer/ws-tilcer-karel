package contact

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/ratelimit"
)

// Handler serves the contact module's public submit + admin inbox endpoints.
type Handler struct {
	svc      *Service
	limiter  *ratelimit.Limiter
	maxBytes int64
}

// NewHandler builds the contact handler.
func NewHandler(svc *Service, limiter *ratelimit.Limiter, maxBytes int64) *Handler {
	return &Handler{svc: svc, limiter: limiter, maxBytes: maxBytes}
}

// MountPublic registers the public contact-submit route.
func (h *Handler) MountPublic(r chi.Router) {
	r.Post("/contact", h.submit)
}

// MountAdmin registers the admin inbox routes (mounted under /api/admin).
func (h *Handler) MountAdmin(r chi.Router) {
	r.Get("/contact", h.list)
	r.Patch("/contact/{id}", h.patch)
	r.Delete("/contact/{id}", h.delete)
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	// Enforce the body cap before decoding.
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)
	var in ContactSubmit
	if err := httpx.DecodeJSONLimit(r, &in, h.maxBytes); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			httpx.WriteError(w, httpx.ErrPayloadTooLarge("message too large"))
			return
		}
		httpx.WriteError(w, httpx.ErrUnprocessable(err.Error()))
		return
	}
	ip := httpx.ClientIP(r)
	if h.limiter != nil && !h.limiter.Allow("contact:"+ip) {
		httpx.SetRetryAfter(w, h.limiter.RetryAfter("contact:"+ip))
		httpx.WriteError(w, httpx.ErrTooManyRequests("too many submissions"))
		return
	}
	if err := h.svc.Submit(r.Context(), in, ip, r.UserAgent()); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusAccepted, OkResult{OK: true})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	page, err := h.svc.List(r.Context(), r.URL.Query().Get("status"), r.URL.Query().Get("cursor"), pageLimit(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, page)
}

func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var in StatusUpdate
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.WriteError(w, httpx.ErrUnprocessable(err.Error()))
		return
	}
	m, err := h.svc.SetStatus(r.Context(), id, in.Status)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, m)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ----

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		httpx.WriteError(w, httpx.ErrUnprocessable("invalid id"))
		return 0, false
	}
	return id, true
}

func pageLimit(r *http.Request) int {
	const def, max = 50, 200
	v := r.URL.Query().Get("limit")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	if n > max {
		return max
	}
	return n
}
