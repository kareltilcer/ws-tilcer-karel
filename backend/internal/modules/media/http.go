package media

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
)

// Handler serves the media module's admin endpoints. Media has no public routes
// (objects are served directly from MEDIA_PUBLIC_BASE_URL).
type Handler struct {
	svc      *Service
	maxBytes int64
}

// NewHandler builds the media handler.
func NewHandler(svc *Service, maxBytes int64) *Handler {
	return &Handler{svc: svc, maxBytes: maxBytes}
}

// MountAdmin registers the media routes (mounted under /api/admin).
func (h *Handler) MountAdmin(r chi.Router) {
	r.Post("/media", h.upload)
	r.Get("/media", h.list)
	r.Delete("/media/{id}", h.delete)
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	// Cap the request body: the media size limit plus a small allowance for
	// multipart framing and the alt fields.
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes+(1<<16))
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			httpx.WriteError(w, httpx.ErrPayloadTooLarge("file exceeds size limit"))
			return
		}
		httpx.WriteError(w, httpx.ErrUnprocessable("invalid multipart form"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, httpx.ErrUnprocessable("missing file field"))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, h.maxBytes+1))
	if err != nil {
		httpx.WriteError(w, httpx.ErrUnprocessable("could not read file"))
		return
	}
	if int64(len(data)) > h.maxBytes {
		httpx.WriteError(w, httpx.ErrPayloadTooLarge("file exceeds size limit"))
		return
	}

	declaredMime := header.Header.Get("Content-Type")
	m, err := h.svc.Upload(r.Context(), declaredMime, data, formPtr(r, "alt_cs"), formPtr(r, "alt_en"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, m)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	page, err := h.svc.List(r.Context(), r.URL.Query().Get("cursor"), pageLimit(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, page)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		httpx.WriteError(w, httpx.ErrUnprocessable("invalid id"))
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func formPtr(r *http.Request, key string) *string {
	v := r.FormValue(key)
	if v == "" {
		return nil
	}
	return &v
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
