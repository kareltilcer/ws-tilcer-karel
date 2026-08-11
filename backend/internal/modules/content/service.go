package content

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"strings"

	appdb "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/db"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/slug"
)

// Service is the content module's business logic: validation, publish filtering,
// reference checks, and transactional reorder/create/update.
type Service struct {
	db    *sql.DB
	store *Store
}

// NewService builds the content service over db.
func NewService(db *sql.DB) *Service {
	return &Service{db: db, store: NewStore(db)}
}

// ==================================================================== public

func (s *Service) ListPublicProjects(ctx context.Context, categorySlug string, featured *bool) ([]ProjectSummary, error) {
	return s.store.ListPublicProjects(ctx, categorySlug, featured)
}

func (s *Service) GetPublicProject(ctx context.Context, projectSlug string) (*ProjectDetail, error) {
	p, err := s.store.GetPublicProjectBySlug(ctx, projectSlug)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, httpx.ErrNotFound("project not found")
	}
	return p, nil
}

func (s *Service) ListVisibleCategories(ctx context.Context) ([]Category, error) {
	return s.store.ListCategories(ctx, true)
}

func (s *Service) ListVisibleLinks(ctx context.Context) ([]Link, error) {
	return s.store.ListLinks(ctx, true)
}

func (s *Service) GetSection(ctx context.Context, key string) (*Section, error) {
	sec, err := s.store.GetSection(ctx, s.db, key)
	if err != nil {
		return nil, err
	}
	if sec == nil {
		return nil, httpx.ErrNotFound("section not found")
	}
	return sec, nil
}

func (s *Service) ListVisibleSkills(ctx context.Context) ([]Skill, error) {
	return s.store.ListSkills(ctx, true)
}

// ListAllSkills returns every skill (incl. hidden) for the admin editor.
func (s *Service) ListAllSkills(ctx context.Context) ([]Skill, error) {
	return s.store.ListSkills(ctx, false)
}

func (s *Service) GetBusinessInfo(ctx context.Context) (*BusinessInfo, error) {
	b, err := s.store.GetBusinessInfo(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if b == nil {
		// Singleton is seeded, but be defensive.
		return &BusinessInfo{LegalName: ""}, nil
	}
	return b, nil
}

// ClickLink increments a link's counter. Best-effort: an unknown id is a 404 the
// client ignores; success is 204.
func (s *Service) ClickLink(ctx context.Context, id int64) error {
	n, err := s.store.IncrementClick(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.ErrNotFound("link not found")
	}
	return nil
}

// ==================================================================== categories (admin)

func (s *Service) ListAllCategories(ctx context.Context) ([]Category, error) {
	return s.store.ListCategories(ctx, false)
}

func (s *Service) CreateCategory(ctx context.Context, in CategoryCreate) (*Category, error) {
	if !slug.ValidCategory(in.Slug) {
		return nil, httpx.ErrUnprocessable("invalid category slug")
	}
	if in.NameCS == "" || in.NameEN == "" {
		return nil, httpx.ErrUnprocessable("name_cs and name_en are required")
	}
	var created *Category
	err := appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		exists, err := s.store.CategorySlugExists(ctx, tx, in.Slug, 0)
		if err != nil {
			return err
		}
		if exists {
			return httpx.ErrConflict("category slug already exists")
		}
		sortOrder := int64(0)
		if in.SortOrder != nil {
			sortOrder = *in.SortOrder
		} else {
			maxs, err := s.store.MaxCategorySort(ctx, tx)
			if err != nil {
				return err
			}
			sortOrder = maxs + 1
		}
		visible := true
		if in.Visible != nil {
			visible = *in.Visible
		}
		id, err := s.store.InsertCategory(ctx, tx, in, sortOrder, visible)
		if err != nil {
			return err
		}
		created, err = s.store.GetCategory(ctx, tx, id)
		return err
	})
	return created, err
}

func (s *Service) UpdateCategory(ctx context.Context, id int64, in CategoryUpdate) (*Category, error) {
	if in.Slug != nil && !slug.ValidCategory(*in.Slug) {
		return nil, httpx.ErrUnprocessable("invalid category slug")
	}
	var updated *Category
	err := appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		cur, err := s.store.GetCategory(ctx, tx, id)
		if err != nil {
			return err
		}
		if cur == nil {
			return httpx.ErrNotFound("category not found")
		}
		if in.Slug != nil {
			exists, err := s.store.CategorySlugExists(ctx, tx, *in.Slug, id)
			if err != nil {
				return err
			}
			if exists {
				return httpx.ErrConflict("category slug already exists")
			}
		}
		if err := s.store.UpdateCategory(ctx, tx, id, in); err != nil {
			return err
		}
		updated, err = s.store.GetCategory(ctx, tx, id)
		return err
	})
	return updated, err
}

func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	return appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		cur, err := s.store.GetCategory(ctx, tx, id)
		if err != nil {
			return err
		}
		if cur == nil {
			return httpx.ErrNotFound("category not found")
		}
		inUse, err := s.store.CategoryInUse(ctx, tx, id)
		if err != nil {
			return err
		}
		if inUse {
			return httpx.ErrConflict("category is still referenced by one or more projects")
		}
		_, err = s.store.DeleteCategory(ctx, tx, id)
		return err
	})
}

func (s *Service) ReorderCategories(ctx context.Context, ids []int64) error {
	return appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		return s.store.Reorder(ctx, tx, "category", ids)
	})
}

// ==================================================================== projects (admin)

func (s *Service) ListAdminProjects(ctx context.Context, status, cursor string, limit int) (*AdminProjectsPage, error) {
	if status != "" && status != "draft" && status != "published" {
		return nil, httpx.ErrUnprocessable("invalid status filter")
	}
	items, next, err := s.store.ListAdminProjects(ctx, status, cursor, limit)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []ProjectDetail{}
	}
	return &AdminProjectsPage{Items: items, NextCursor: next}, nil
}

func (s *Service) GetAdminProject(ctx context.Context, id int64) (*ProjectDetail, error) {
	p, err := s.store.GetProjectByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, httpx.ErrNotFound("project not found")
	}
	return p, nil
}

func (s *Service) CreateProject(ctx context.Context, in ProjectCreate) (*ProjectDetail, error) {
	if !slug.ValidProject(in.Slug) {
		return nil, httpx.ErrUnprocessable("invalid project slug")
	}
	if in.TitleCS == "" || in.TitleEN == "" {
		return nil, httpx.ErrUnprocessable("title_cs and title_en are required")
	}
	status := "draft"
	if in.Status != nil {
		if *in.Status != "draft" && *in.Status != "published" {
			return nil, httpx.ErrUnprocessable("invalid status")
		}
		status = *in.Status
	}
	if err := validateProjectLinks(in.Links); err != nil {
		return nil, err
	}
	linksJSON := marshalLinks(in.Links)

	var created *ProjectDetail
	err := appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		exists, err := s.store.ProjectSlugExists(ctx, tx, in.Slug, 0)
		if err != nil {
			return err
		}
		if exists {
			return httpx.ErrConflict("project slug already exists")
		}
		// category must exist
		cat, err := s.store.GetCategory(ctx, tx, in.CategoryID)
		if err != nil {
			return err
		}
		if cat == nil {
			return httpx.ErrUnprocessable("unknown category")
		}
		// referenced media must exist
		if err := s.checkMediaExists(ctx, tx, in.CoverMediaID, in.Gallery); err != nil {
			return err
		}
		featured := in.Featured != nil && *in.Featured
		sortOrder := int64(0)
		if in.SortOrder != nil {
			sortOrder = *in.SortOrder
		} else {
			maxs, err := s.store.MaxProjectSort(ctx, tx)
			if err != nil {
				return err
			}
			sortOrder = maxs + 1
		}
		var coverID any
		if in.CoverMediaID != nil {
			coverID = *in.CoverMediaID
		}
		id, err := s.store.InsertProject(ctx, tx, in, sortOrder, featured, status, linksJSON, coverID)
		if err != nil {
			return err
		}
		if len(in.Gallery) > 0 {
			if err := s.store.SetGallery(ctx, tx, id, in.Gallery); err != nil {
				return err
			}
		}
		created, err = s.store.GetProjectByID(ctx, tx, id)
		return err
	})
	return created, err
}

func (s *Service) UpdateProject(ctx context.Context, id int64, in ProjectUpdate) (*ProjectDetail, error) {
	if in.Slug != nil && !slug.ValidProject(*in.Slug) {
		return nil, httpx.ErrUnprocessable("invalid project slug")
	}
	if in.Status != nil && *in.Status != "draft" && *in.Status != "published" {
		return nil, httpx.ErrUnprocessable("invalid status")
	}
	if in.Links != nil {
		if err := validateProjectLinks(*in.Links); err != nil {
			return nil, err
		}
	}
	var updated *ProjectDetail
	err := appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		cur, err := s.store.GetProjectByID(ctx, tx, id)
		if err != nil {
			return err
		}
		if cur == nil {
			return httpx.ErrNotFound("project not found")
		}
		if in.Slug != nil {
			exists, err := s.store.ProjectSlugExists(ctx, tx, *in.Slug, id)
			if err != nil {
				return err
			}
			if exists {
				return httpx.ErrConflict("project slug already exists")
			}
		}
		if in.CategoryID != nil {
			cat, err := s.store.GetCategory(ctx, tx, *in.CategoryID)
			if err != nil {
				return err
			}
			if cat == nil {
				return httpx.ErrUnprocessable("unknown category")
			}
		}
		// media existence checks
		var coverPtr *int64
		if in.CoverMediaID.Present {
			coverPtr = in.CoverMediaID.Value
		}
		var gallery []int64
		if in.Gallery != nil {
			gallery = *in.Gallery
		}
		if err := s.checkMediaExists(ctx, tx, coverPtr, gallery); err != nil {
			return err
		}

		sets := map[string]any{}
		if in.Slug != nil {
			sets["slug"] = *in.Slug
		}
		if in.CategoryID != nil {
			sets["category_id"] = *in.CategoryID
		}
		if in.TitleCS != nil {
			sets["title_cs"] = *in.TitleCS
		}
		if in.TitleEN != nil {
			sets["title_en"] = *in.TitleEN
		}
		if in.SummaryCS != nil {
			sets["summary_cs"] = nullable(*in.SummaryCS)
		}
		if in.SummaryEN != nil {
			sets["summary_en"] = nullable(*in.SummaryEN)
		}
		if in.BodyCS != nil {
			sets["body_cs"] = nullable(*in.BodyCS)
		}
		if in.BodyEN != nil {
			sets["body_en"] = nullable(*in.BodyEN)
		}
		if in.CoverMediaID.Present {
			if in.CoverMediaID.Value != nil {
				sets["cover_media_id"] = *in.CoverMediaID.Value
			} else {
				sets["cover_media_id"] = nil
			}
		}
		if in.ProjectDate != nil {
			sets["project_date"] = nullable(*in.ProjectDate)
		}
		if in.Featured != nil {
			sets["featured"] = boolInt(*in.Featured)
		}
		if in.Status != nil {
			sets["status"] = *in.Status
		}
		if in.SortOrder != nil {
			sets["sort_order"] = *in.SortOrder
		}
		if in.Links != nil {
			sets["links"] = nullable(marshalLinks(*in.Links))
		}
		if err := s.store.UpdateProject(ctx, tx, id, sets); err != nil {
			return err
		}
		if in.Gallery != nil {
			if err := s.store.SetGallery(ctx, tx, id, *in.Gallery); err != nil {
				return err
			}
		}
		updated, err = s.store.GetProjectByID(ctx, tx, id)
		return err
	})
	return updated, err
}

func (s *Service) DeleteProject(ctx context.Context, id int64) error {
	return appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		n, err := s.store.DeleteProject(ctx, tx, id) // project_media rows CASCADE
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.ErrNotFound("project not found")
		}
		return nil
	})
}

func (s *Service) ReorderProjects(ctx context.Context, ids []int64) error {
	return appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		return s.store.Reorder(ctx, tx, "project", ids)
	})
}

// checkMediaExists validates cover + gallery media ids exist (422 if not).
func (s *Service) checkMediaExists(ctx context.Context, tx DBTX, cover *int64, gallery []int64) error {
	var ids []int64
	if cover != nil {
		ids = append(ids, *cover)
	}
	ids = append(ids, gallery...)
	if len(ids) == 0 {
		return nil
	}
	ok, err := s.store.MediaIDsExist(ctx, tx, ids)
	if err != nil {
		return err
	}
	if !ok {
		return httpx.ErrUnprocessable("one or more referenced media ids do not exist")
	}
	return nil
}

// ==================================================================== links (admin)

func (s *Service) ListAllLinks(ctx context.Context) ([]Link, error) {
	return s.store.ListLinks(ctx, false)
}

func (s *Service) CreateLink(ctx context.Context, in LinkCreate) (*Link, error) {
	if in.LabelCS == "" || in.LabelEN == "" {
		return nil, httpx.ErrUnprocessable("label_cs and label_en are required")
	}
	if !validURL(in.URL) {
		return nil, httpx.ErrUnprocessable("url must be http or https")
	}
	var created *Link
	err := appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		sortOrder := int64(0)
		if in.SortOrder != nil {
			sortOrder = *in.SortOrder
		} else {
			maxs, err := s.store.MaxLinkSort(ctx, tx)
			if err != nil {
				return err
			}
			sortOrder = maxs + 1
		}
		visible := true
		if in.Visible != nil {
			visible = *in.Visible
		}
		id, err := s.store.InsertLink(ctx, tx, in, sortOrder, visible)
		if err != nil {
			return err
		}
		created, err = s.store.GetLink(ctx, tx, id)
		return err
	})
	return created, err
}

func (s *Service) UpdateLink(ctx context.Context, id int64, in LinkUpdate) (*Link, error) {
	if in.URL != nil && !validURL(*in.URL) {
		return nil, httpx.ErrUnprocessable("url must be http or https")
	}
	var updated *Link
	err := appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		cur, err := s.store.GetLink(ctx, tx, id)
		if err != nil {
			return err
		}
		if cur == nil {
			return httpx.ErrNotFound("link not found")
		}
		if err := s.store.UpdateLink(ctx, tx, id, in); err != nil {
			return err
		}
		updated, err = s.store.GetLink(ctx, tx, id)
		return err
	})
	return updated, err
}

func (s *Service) DeleteLink(ctx context.Context, id int64) error {
	return appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		n, err := s.store.DeleteLink(ctx, tx, id)
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.ErrNotFound("link not found")
		}
		return nil
	})
}

func (s *Service) ReorderLinks(ctx context.Context, ids []int64) error {
	return appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		return s.store.Reorder(ctx, tx, "link", ids)
	})
}

// ==================================================================== sections (admin)

func (s *Service) UpsertSection(ctx context.Context, key string, in SectionUpdate) (*Section, error) {
	if strings.TrimSpace(key) == "" {
		return nil, httpx.ErrUnprocessable("section key required")
	}
	var out *Section
	err := appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if in.MediaID.Present && in.MediaID.Value != nil {
			ok, err := s.store.MediaIDsExist(ctx, tx, []int64{*in.MediaID.Value})
			if err != nil {
				return err
			}
			if !ok {
				return httpx.ErrUnprocessable("unknown media id")
			}
		}
		sets := map[string]any{}
		if in.HeadingCS != nil {
			sets["heading_cs"] = nullable(*in.HeadingCS)
		}
		if in.HeadingEN != nil {
			sets["heading_en"] = nullable(*in.HeadingEN)
		}
		if in.BodyCS != nil {
			sets["body_cs"] = nullable(*in.BodyCS)
		}
		if in.BodyEN != nil {
			sets["body_en"] = nullable(*in.BodyEN)
		}
		if in.MediaID.Present {
			if in.MediaID.Value != nil {
				sets["media_id"] = *in.MediaID.Value
			} else {
				sets["media_id"] = nil
			}
		}
		if err := s.store.UpsertSection(ctx, tx, key, sets); err != nil {
			return err
		}
		var err error
		out, err = s.store.GetSection(ctx, tx, key)
		return err
	})
	return out, err
}

// ==================================================================== skills (admin)

func (s *Service) CreateSkill(ctx context.Context, in SkillCreate) (*Skill, error) {
	if in.NameCS == "" || in.NameEN == "" || in.CategoryCS == "" || in.CategoryEN == "" {
		return nil, httpx.ErrUnprocessable("name_cs, name_en, category_cs and category_en are required")
	}
	if in.Level != nil && (*in.Level < 1 || *in.Level > 5) {
		return nil, httpx.ErrUnprocessable("level must be 1..5")
	}
	var created *Skill
	err := appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		sortOrder := int64(0)
		if in.SortOrder != nil {
			sortOrder = *in.SortOrder
		} else {
			maxs, err := s.store.MaxSkillSort(ctx, tx)
			if err != nil {
				return err
			}
			sortOrder = maxs + 1
		}
		visible := true
		if in.Visible != nil {
			visible = *in.Visible
		}
		id, err := s.store.InsertSkill(ctx, tx, in, sortOrder, visible)
		if err != nil {
			return err
		}
		created, err = s.store.GetSkill(ctx, tx, id)
		return err
	})
	return created, err
}

func (s *Service) UpdateSkill(ctx context.Context, id int64, in SkillUpdate) (*Skill, error) {
	if in.Level.Present && in.Level.Value != nil && (*in.Level.Value < 1 || *in.Level.Value > 5) {
		return nil, httpx.ErrUnprocessable("level must be 1..5")
	}
	var updated *Skill
	err := appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		cur, err := s.store.GetSkill(ctx, tx, id)
		if err != nil {
			return err
		}
		if cur == nil {
			return httpx.ErrNotFound("skill not found")
		}
		if err := s.store.UpdateSkill(ctx, tx, id, in); err != nil {
			return err
		}
		updated, err = s.store.GetSkill(ctx, tx, id)
		return err
	})
	return updated, err
}

func (s *Service) DeleteSkill(ctx context.Context, id int64) error {
	return appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		n, err := s.store.DeleteSkill(ctx, tx, id)
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.ErrNotFound("skill not found")
		}
		return nil
	})
}

func (s *Service) ReorderSkills(ctx context.Context, ids []int64) error {
	return appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		return s.store.Reorder(ctx, tx, "skill", ids)
	})
}

// ==================================================================== business info (admin)

func (s *Service) UpdateBusinessInfo(ctx context.Context, in BusinessInfoUpdate) (*BusinessInfo, error) {
	var out *BusinessInfo
	err := appdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		sets := map[string]any{}
		if in.LegalName != nil {
			sets["legal_name"] = *in.LegalName // NOT NULL column
		}
		strFields := map[string]*string{
			"ico": in.ICO, "dic": in.DIC, "address_lines": in.AddressLines, "city": in.City,
			"zip": in.Zip, "country": in.Country, "register_note_cs": in.RegisterNoteCS,
			"register_note_en": in.RegisterNoteEN, "contact_email": in.ContactEmail,
			"contact_phone": in.ContactPhone, "data_note_cs": in.DataNoteCS, "data_note_en": in.DataNoteEN,
		}
		for col, p := range strFields {
			if p != nil {
				sets[col] = nullable(*p)
			}
		}
		if err := s.store.UpdateBusinessInfo(ctx, tx, sets); err != nil {
			return err
		}
		var err error
		out, err = s.store.GetBusinessInfo(ctx, tx)
		return err
	})
	return out, err
}

// ==================================================================== helpers

func validURL(u string) bool {
	parsed, err := url.Parse(strings.TrimSpace(u))
	if err != nil {
		return false
	}
	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func validateProjectLinks(links []ProjectLink) error {
	for _, l := range links {
		if l.Label == "" || !validURL(l.URL) {
			return httpx.ErrUnprocessable("each project link needs a label and a valid http(s) url")
		}
	}
	return nil
}

func marshalLinks(links []ProjectLink) string {
	if len(links) == 0 {
		return ""
	}
	b, err := json.Marshal(links)
	if err != nil {
		return ""
	}
	return string(b)
}
