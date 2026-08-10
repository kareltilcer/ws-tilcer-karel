package content

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/ratelimit"
)

// Handler serves the content module's public + admin HTTP endpoints.
type Handler struct {
	svc          *Service
	clickLimiter *ratelimit.Limiter
}

// NewHandler builds the content handler. clickLimiter rate-limits the public
// link-click counter per IP.
func NewHandler(svc *Service, clickLimiter *ratelimit.Limiter) *Handler {
	return &Handler{svc: svc, clickLimiter: clickLimiter}
}

// MountPublic registers the unauthenticated read + link-click routes.
func (h *Handler) MountPublic(r chi.Router) {
	r.Get("/projects", h.listPublicProjects)
	r.Get("/projects/{slug}", h.getPublicProject)
	r.Get("/categories", h.listVisibleCategories)
	r.Get("/links", h.listVisibleLinks)
	r.Get("/sections/{key}", h.getSection)
	r.Get("/skills", h.listVisibleSkills)
	r.Get("/business-info", h.getBusinessInfo)
	r.Post("/links/{id}/click", h.clickLink)
}

// MountAdmin registers the authenticated CMS routes (mounted under /api/admin).
func (h *Handler) MountAdmin(r chi.Router) {
	// projects
	r.Get("/projects", h.adminListProjects)
	r.Post("/projects", h.createProject)
	r.Post("/projects/reorder", h.reorderProjects)
	r.Get("/projects/{id}", h.adminGetProject)
	r.Patch("/projects/{id}", h.updateProject)
	r.Delete("/projects/{id}", h.deleteProject)
	// categories
	r.Get("/categories", h.adminListCategories)
	r.Post("/categories", h.createCategory)
	r.Post("/categories/reorder", h.reorderCategories)
	r.Patch("/categories/{id}", h.updateCategory)
	r.Delete("/categories/{id}", h.deleteCategory)
	// links
	r.Get("/links", h.adminListLinks)
	r.Post("/links", h.createLink)
	r.Post("/links/reorder", h.reorderLinks)
	r.Patch("/links/{id}", h.updateLink)
	r.Delete("/links/{id}", h.deleteLink)
	// sections
	r.Put("/sections/{key}", h.upsertSection)
	// skills
	r.Get("/skills", h.adminListSkills)
	r.Post("/skills", h.createSkill)
	r.Post("/skills/reorder", h.reorderSkills)
	r.Patch("/skills/{id}", h.updateSkill)
	r.Delete("/skills/{id}", h.deleteSkill)
	// business info
	r.Put("/business-info", h.updateBusinessInfo)
}

// ---- public handlers ----

func (h *Handler) listPublicProjects(w http.ResponseWriter, r *http.Request) {
	var featured *bool
	if v := r.URL.Query().Get("featured"); v == "true" || v == "1" {
		t := true
		featured = &t
	}
	items, err := h.svc.ListPublicProjects(r.Context(), r.URL.Query().Get("category"), featured)
	respond(w, http.StatusOK, items, err)
}

func (h *Handler) getPublicProject(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetPublicProject(r.Context(), chi.URLParam(r, "slug"))
	respond(w, http.StatusOK, p, err)
}

func (h *Handler) listVisibleCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListVisibleCategories(r.Context())
	respond(w, http.StatusOK, items, err)
}

func (h *Handler) listVisibleLinks(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListVisibleLinks(r.Context())
	respond(w, http.StatusOK, items, err)
}

func (h *Handler) getSection(w http.ResponseWriter, r *http.Request) {
	sec, err := h.svc.GetSection(r.Context(), chi.URLParam(r, "key"))
	respond(w, http.StatusOK, sec, err)
}

func (h *Handler) listVisibleSkills(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListVisibleSkills(r.Context())
	respond(w, http.StatusOK, items, err)
}

func (h *Handler) getBusinessInfo(w http.ResponseWriter, r *http.Request) {
	b, err := h.svc.GetBusinessInfo(r.Context())
	respond(w, http.StatusOK, b, err)
}

func (h *Handler) clickLink(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	ip := httpx.ClientIP(r)
	if h.clickLimiter != nil && !h.clickLimiter.Allow("click:"+ip) {
		httpx.WriteError(w, httpx.ErrTooManyRequests("slow down"))
		return
	}
	respondNoContent(w, h.svc.ClickLink(r.Context(), id))
}

// ---- admin: projects ----

func (h *Handler) adminListProjects(w http.ResponseWriter, r *http.Request) {
	page, err := h.svc.ListAdminProjects(r.Context(), r.URL.Query().Get("status"), r.URL.Query().Get("cursor"), pageLimit(r))
	respond(w, http.StatusOK, page, err)
}

func (h *Handler) adminGetProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	p, err := h.svc.GetAdminProject(r.Context(), id)
	respond(w, http.StatusOK, p, err)
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var in ProjectCreate
	if !decode(w, r, &in) {
		return
	}
	p, err := h.svc.CreateProject(r.Context(), in)
	respond(w, http.StatusCreated, p, err)
}

func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var in ProjectUpdate
	if !decode(w, r, &in) {
		return
	}
	p, err := h.svc.UpdateProject(r.Context(), id, in)
	respond(w, http.StatusOK, p, err)
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	respondNoContent(w, h.svc.DeleteProject(r.Context(), id))
}

func (h *Handler) reorderProjects(w http.ResponseWriter, r *http.Request) {
	var in ReorderRequest
	if !decode(w, r, &in) {
		return
	}
	respondNoContent(w, h.svc.ReorderProjects(r.Context(), in.IDs))
}

// ---- admin: categories ----

func (h *Handler) adminListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAllCategories(r.Context())
	respond(w, http.StatusOK, items, err)
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var in CategoryCreate
	if !decode(w, r, &in) {
		return
	}
	c, err := h.svc.CreateCategory(r.Context(), in)
	respond(w, http.StatusCreated, c, err)
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var in CategoryUpdate
	if !decode(w, r, &in) {
		return
	}
	c, err := h.svc.UpdateCategory(r.Context(), id, in)
	respond(w, http.StatusOK, c, err)
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	respondNoContent(w, h.svc.DeleteCategory(r.Context(), id))
}

func (h *Handler) reorderCategories(w http.ResponseWriter, r *http.Request) {
	var in ReorderRequest
	if !decode(w, r, &in) {
		return
	}
	respondNoContent(w, h.svc.ReorderCategories(r.Context(), in.IDs))
}

// ---- admin: links ----

func (h *Handler) adminListLinks(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAllLinks(r.Context())
	respond(w, http.StatusOK, items, err)
}

func (h *Handler) createLink(w http.ResponseWriter, r *http.Request) {
	var in LinkCreate
	if !decode(w, r, &in) {
		return
	}
	l, err := h.svc.CreateLink(r.Context(), in)
	respond(w, http.StatusCreated, l, err)
}

func (h *Handler) updateLink(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var in LinkUpdate
	if !decode(w, r, &in) {
		return
	}
	l, err := h.svc.UpdateLink(r.Context(), id, in)
	respond(w, http.StatusOK, l, err)
}

func (h *Handler) deleteLink(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	respondNoContent(w, h.svc.DeleteLink(r.Context(), id))
}

func (h *Handler) reorderLinks(w http.ResponseWriter, r *http.Request) {
	var in ReorderRequest
	if !decode(w, r, &in) {
		return
	}
	respondNoContent(w, h.svc.ReorderLinks(r.Context(), in.IDs))
}

// ---- admin: sections ----

func (h *Handler) upsertSection(w http.ResponseWriter, r *http.Request) {
	var in SectionUpdate
	if !decode(w, r, &in) {
		return
	}
	sec, err := h.svc.UpsertSection(r.Context(), chi.URLParam(r, "key"), in)
	respond(w, http.StatusOK, sec, err)
}

// ---- admin: skills ----

func (h *Handler) adminListSkills(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAllSkills(r.Context())
	respond(w, http.StatusOK, items, err)
}

func (h *Handler) createSkill(w http.ResponseWriter, r *http.Request) {
	var in SkillCreate
	if !decode(w, r, &in) {
		return
	}
	sk, err := h.svc.CreateSkill(r.Context(), in)
	respond(w, http.StatusCreated, sk, err)
}

func (h *Handler) updateSkill(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var in SkillUpdate
	if !decode(w, r, &in) {
		return
	}
	sk, err := h.svc.UpdateSkill(r.Context(), id, in)
	respond(w, http.StatusOK, sk, err)
}

func (h *Handler) deleteSkill(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	respondNoContent(w, h.svc.DeleteSkill(r.Context(), id))
}

func (h *Handler) reorderSkills(w http.ResponseWriter, r *http.Request) {
	var in ReorderRequest
	if !decode(w, r, &in) {
		return
	}
	respondNoContent(w, h.svc.ReorderSkills(r.Context(), in.IDs))
}

// ---- admin: business info ----

func (h *Handler) updateBusinessInfo(w http.ResponseWriter, r *http.Request) {
	var in BusinessInfoUpdate
	if !decode(w, r, &in) {
		return
	}
	b, err := h.svc.UpdateBusinessInfo(r.Context(), in)
	respond(w, http.StatusOK, b, err)
}

// ---- shared helpers ----

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := httpx.DecodeJSON(r, dst); err != nil {
		httpx.WriteError(w, httpx.ErrUnprocessable(err.Error()))
		return false
	}
	return true
}

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

func respond(w http.ResponseWriter, status int, v any, err error) {
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, status, v)
}

func respondNoContent(w http.ResponseWriter, err error) {
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
