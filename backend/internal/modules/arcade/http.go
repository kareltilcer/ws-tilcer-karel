package arcade

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/ratelimit"
)

// maxScoreBody caps the score-submit body (it is tiny: name + score + token).
const maxScoreBody = 4 << 10

// Handler serves the arcade module's public leaderboard/submit + admin moderation.
type Handler struct {
	svc     *Service
	limiter *ratelimit.Limiter
}

// NewHandler builds the arcade handler. limiter rate-limits score submits per IP.
func NewHandler(svc *Service, limiter *ratelimit.Limiter) *Handler {
	return &Handler{svc: svc, limiter: limiter}
}

// MountPublic registers the leaderboard read + score submit.
func (h *Handler) MountPublic(r chi.Router) {
	r.Get("/arcade/{gameKey}/scores", h.leaderboard)
	r.Post("/arcade/{gameKey}/scores", h.submit)
}

// MountAdmin registers the moderation routes (mounted under /api/admin).
func (h *Handler) MountAdmin(r chi.Router) {
	r.Get("/arcade/{gameKey}/scores", h.adminListScores)
	r.Delete("/arcade/scores/{id}", h.deleteScore)
	r.Post("/arcade/{gameKey}/reset", h.resetBoard)
}

func (h *Handler) adminListScores(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	scores, err := h.svc.ListScoresAdmin(r.Context(), chi.URLParam(r, "gameKey"), limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, scores)
}

func (h *Handler) leaderboard(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	top, err := h.svc.Leaderboard(r.Context(), chi.URLParam(r, "gameKey"), limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, top)
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxScoreBody)
	var in ScoreSubmit
	if err := httpx.DecodeJSONLimit(r, &in, maxScoreBody); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			httpx.WriteError(w, httpx.ErrPayloadTooLarge("body too large"))
			return
		}
		httpx.WriteError(w, httpx.ErrUnprocessable(err.Error()))
		return
	}
	ip := httpx.ClientIP(r)
	if h.limiter != nil && !h.limiter.Allow("score:"+ip) {
		httpx.SetRetryAfter(w, h.limiter.RetryAfter("score:"+ip))
		httpx.WriteError(w, httpx.ErrTooManyRequests("too many submissions"))
		return
	}
	res, err := h.svc.SubmitScore(r.Context(), chi.URLParam(r, "gameKey"), in, ip, r.UserAgent())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, res)
}

func (h *Handler) deleteScore(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		httpx.WriteError(w, httpx.ErrUnprocessable("invalid id"))
		return
	}
	if err := h.svc.DeleteScore(r.Context(), id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) resetBoard(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.ResetBoard(r.Context(), chi.URLParam(r, "gameKey")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
