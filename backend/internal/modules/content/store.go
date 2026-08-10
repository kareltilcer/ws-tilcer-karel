package content

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

// DBTX is satisfied by *sql.DB and *sql.Tx. Reads use the store's *sql.DB;
// multi-statement mutations take the explicit tx so they commit atomically.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Store is the content module's SQL access layer.
type Store struct{ db *sql.DB }

// NewStore returns a content store over db.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

const tsFormat = "2006-01-02T15:04:05.000000Z07:00"

func nowUTC() string { return time.Now().UTC().Format(tsFormat) }

// ---- small helpers ----

func nsPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func niPtr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	v := ni.Int64
	return &v
}

// nullable stores "" as SQL NULL (empty optional text == "missing translation").
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullablePtr(p *string) any {
	if p == nil || *p == "" {
		return nil
	}
	return *p
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// ==========================================================================
// media (shared table; content reads MediaRefs for embedding)
// ==========================================================================

func (s *Store) mediaRef(ctx context.Context, q DBTX, id int64) (*MediaRef, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, public_url, alt_cs, alt_en, width, height FROM media WHERE id = ?`, id)
	var m MediaRef
	var alt1, alt2 sql.NullString
	var w, h sql.NullInt64
	if err := row.Scan(&m.ID, &m.PublicURL, &alt1, &alt2, &w, &h); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	m.AltCS, m.AltEN, m.Width, m.Height = nsPtr(alt1), nsPtr(alt2), niPtr(w), niPtr(h)
	return &m, nil
}

// ==========================================================================
// category
// ==========================================================================

const categoryCols = `id, slug, name_cs, name_en, sort_order, visible`

func scanCategory(r interface{ Scan(...any) error }) (Category, error) {
	var c Category
	var visible int
	if err := r.Scan(&c.ID, &c.Slug, &c.NameCS, &c.NameEN, &c.SortOrder, &visible); err != nil {
		return Category{}, err
	}
	c.Visible = visible != 0
	return c, nil
}

func (s *Store) ListCategories(ctx context.Context, visibleOnly bool) ([]Category, error) {
	where := ""
	if visibleOnly {
		where = " WHERE visible = 1"
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+categoryCols+` FROM category`+where+` ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Category{}
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetCategory(ctx context.Context, q DBTX, id int64) (*Category, error) {
	row := q.QueryRowContext(ctx, `SELECT `+categoryCols+` FROM category WHERE id = ?`, id)
	c, err := scanCategory(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) CategorySlugExists(ctx context.Context, q DBTX, slug string, excludeID int64) (bool, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM category WHERE slug = ? AND id != ?`, slug, excludeID).Scan(&n)
	return n > 0, err
}

func (s *Store) InsertCategory(ctx context.Context, q DBTX, in CategoryCreate, sortOrder int64, visible bool) (int64, error) {
	res, err := q.ExecContext(ctx,
		`INSERT INTO category (slug, name_cs, name_en, sort_order, visible, created_at)
		 VALUES (?,?,?,?,?,?)`,
		in.Slug, in.NameCS, in.NameEN, sortOrder, boolInt(visible), nowUTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateCategory(ctx context.Context, q DBTX, id int64, in CategoryUpdate) error {
	var sets []string
	var args []any
	if in.Slug != nil {
		sets, args = append(sets, "slug = ?"), append(args, *in.Slug)
	}
	if in.NameCS != nil {
		sets, args = append(sets, "name_cs = ?"), append(args, *in.NameCS)
	}
	if in.NameEN != nil {
		sets, args = append(sets, "name_en = ?"), append(args, *in.NameEN)
	}
	if in.SortOrder != nil {
		sets, args = append(sets, "sort_order = ?"), append(args, *in.SortOrder)
	}
	if in.Visible != nil {
		sets, args = append(sets, "visible = ?"), append(args, boolInt(*in.Visible))
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	_, err := q.ExecContext(ctx, `UPDATE category SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...)
	return err
}

func (s *Store) CategoryInUse(ctx context.Context, q DBTX, id int64) (bool, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM project WHERE category_id = ?`, id).Scan(&n)
	return n > 0, err
}

func (s *Store) DeleteCategory(ctx context.Context, q DBTX, id int64) (int64, error) {
	res, err := q.ExecContext(ctx, `DELETE FROM category WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *Store) MaxCategorySort(ctx context.Context, q DBTX) (int64, error) {
	var m sql.NullInt64
	err := q.QueryRowContext(ctx, `SELECT MAX(sort_order) FROM category`).Scan(&m)
	return m.Int64, err
}

// ==========================================================================
// project
// ==========================================================================

// projectSummaryCols / projectDetailCols are the SELECT column lists; projectFrom
// is the shared FROM+JOIN clause. They are kept separate so the detail query can
// extend the column list WITHOUT appending columns after the FROM (a SQL error).
const projectSummaryCols = `p.id, p.slug, p.title_cs, p.title_en, p.summary_cs, p.summary_en,
       p.featured, p.status, p.project_date, p.sort_order,
       c.id, c.slug, c.name_cs, c.name_en, c.sort_order, c.visible,
       m.id, m.public_url, m.alt_cs, m.alt_en, m.width, m.height`

const projectDetailCols = projectSummaryCols + `, p.body_cs, p.body_en, p.links, p.created_at, p.updated_at`

const projectFrom = `
FROM project p
JOIN category c ON c.id = p.category_id
LEFT JOIN media m ON m.id = p.cover_media_id`

func scanProjectSummary(r interface{ Scan(...any) error }) (ProjectSummary, error) {
	var p ProjectSummary
	var sumCS, sumEN, projDate sql.NullString
	var featured int
	var cat Category
	var catVisible int
	var covID sql.NullInt64
	var covURL, covAltCS, covAltEN sql.NullString
	var covW, covH sql.NullInt64
	if err := r.Scan(
		&p.ID, &p.Slug, &p.TitleCS, &p.TitleEN, &sumCS, &sumEN,
		&featured, &p.Status, &projDate, &p.SortOrder,
		&cat.ID, &cat.Slug, &cat.NameCS, &cat.NameEN, &cat.SortOrder, &catVisible,
		&covID, &covURL, &covAltCS, &covAltEN, &covW, &covH,
	); err != nil {
		return ProjectSummary{}, err
	}
	p.SummaryCS, p.SummaryEN, p.ProjectDate = nsPtr(sumCS), nsPtr(sumEN), nsPtr(projDate)
	p.Featured = featured != 0
	cat.Visible = catVisible != 0
	p.Category = cat
	if covID.Valid {
		p.Cover = &MediaRef{ID: covID.Int64, PublicURL: covURL.String,
			AltCS: nsPtr(covAltCS), AltEN: nsPtr(covAltEN), Width: niPtr(covW), Height: niPtr(covH)}
	}
	return p, nil
}

// ListPublicProjects returns published projects, optionally filtered by category
// slug / featured, ordered by sort_order then project_date desc (PRD FR-1).
func (s *Store) ListPublicProjects(ctx context.Context, categorySlug string, featured *bool) ([]ProjectSummary, error) {
	conds := []string{"p.status = 'published'"}
	var args []any
	if categorySlug != "" {
		conds = append(conds, "c.slug = ?")
		args = append(args, categorySlug)
	}
	if featured != nil && *featured {
		conds = append(conds, "p.featured = 1")
	}
	q := `SELECT ` + projectSummaryCols + projectFrom + " WHERE " + strings.Join(conds, " AND ") +
		" ORDER BY p.sort_order, p.project_date DESC, p.id"
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ProjectSummary{}
	for rows.Next() {
		p, err := scanProjectSummary(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// getProjectDetail loads a full ProjectDetail by a WHERE clause (slug or id).
func (s *Store) getProjectDetail(ctx context.Context, q DBTX, where string, arg any, publishedOnly bool) (*ProjectDetail, error) {
	cond := where
	if publishedOnly {
		cond += " AND p.status = 'published'"
	}
	query := `SELECT ` + projectDetailCols + projectFrom + " WHERE " + cond
	row := q.QueryRowContext(ctx, query, arg)

	var d ProjectDetail
	var sumCS, sumEN, projDate sql.NullString
	var featured int
	var cat Category
	var catVisible int
	var covID sql.NullInt64
	var covURL, covAltCS, covAltEN sql.NullString
	var covW, covH sql.NullInt64
	var bodyCS, bodyEN, linksJSON sql.NullString
	if err := row.Scan(
		&d.ID, &d.Slug, &d.TitleCS, &d.TitleEN, &sumCS, &sumEN,
		&featured, &d.Status, &projDate, &d.SortOrder,
		&cat.ID, &cat.Slug, &cat.NameCS, &cat.NameEN, &cat.SortOrder, &catVisible,
		&covID, &covURL, &covAltCS, &covAltEN, &covW, &covH,
		&bodyCS, &bodyEN, &linksJSON, &d.CreatedAt, &d.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	d.SummaryCS, d.SummaryEN, d.ProjectDate = nsPtr(sumCS), nsPtr(sumEN), nsPtr(projDate)
	d.Featured = featured != 0
	cat.Visible = catVisible != 0
	d.Category = cat
	if covID.Valid {
		d.Cover = &MediaRef{ID: covID.Int64, PublicURL: covURL.String,
			AltCS: nsPtr(covAltCS), AltEN: nsPtr(covAltEN), Width: niPtr(covW), Height: niPtr(covH)}
	}
	d.BodyCS, d.BodyEN = nsPtr(bodyCS), nsPtr(bodyEN)
	d.Links = parseProjectLinks(linksJSON)
	gallery, err := s.projectGallery(ctx, q, d.ID)
	if err != nil {
		return nil, err
	}
	d.Gallery = gallery
	return &d, nil
}

func (s *Store) GetPublicProjectBySlug(ctx context.Context, slug string) (*ProjectDetail, error) {
	return s.getProjectDetail(ctx, s.db, "p.slug = ?", slug, true)
}

func (s *Store) GetProjectByID(ctx context.Context, q DBTX, id int64) (*ProjectDetail, error) {
	return s.getProjectDetail(ctx, q, "p.id = ?", id, false)
}

func (s *Store) projectGallery(ctx context.Context, q DBTX, projectID int64) ([]MediaRef, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT m.id, m.public_url, m.alt_cs, m.alt_en, m.width, m.height
		 FROM project_media pm JOIN media m ON m.id = pm.media_id
		 WHERE pm.project_id = ? ORDER BY pm.sort_order, pm.media_id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MediaRef{}
	for rows.Next() {
		var m MediaRef
		var alt1, alt2 sql.NullString
		var w, h sql.NullInt64
		if err := rows.Scan(&m.ID, &m.PublicURL, &alt1, &alt2, &w, &h); err != nil {
			return nil, err
		}
		m.AltCS, m.AltEN, m.Width, m.Height = nsPtr(alt1), nsPtr(alt2), niPtr(w), niPtr(h)
		out = append(out, m)
	}
	return out, rows.Err()
}

// galleriesFor loads the gallery MediaRefs for every project id in one query,
// grouped by project id (avoids an N+1 per-project fetch in list endpoints).
func (s *Store) galleriesFor(ctx context.Context, q DBTX, projectIDs []int64) (map[int64][]MediaRef, error) {
	out := map[int64][]MediaRef{}
	if len(projectIDs) == 0 {
		return out, nil
	}
	args := make([]any, len(projectIDs))
	for i, id := range projectIDs {
		args[i] = id
	}
	rows, err := q.QueryContext(ctx,
		`SELECT pm.project_id, m.id, m.public_url, m.alt_cs, m.alt_en, m.width, m.height
		 FROM project_media pm JOIN media m ON m.id = pm.media_id
		 WHERE pm.project_id IN (`+placeholders(len(projectIDs))+`)
		 ORDER BY pm.project_id, pm.sort_order, pm.media_id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pid int64
		var m MediaRef
		var alt1, alt2 sql.NullString
		var w, h sql.NullInt64
		if err := rows.Scan(&pid, &m.ID, &m.PublicURL, &alt1, &alt2, &w, &h); err != nil {
			return nil, err
		}
		m.AltCS, m.AltEN, m.Width, m.Height = nsPtr(alt1), nsPtr(alt2), niPtr(w), niPtr(h)
		out[pid] = append(out[pid], m)
	}
	return out, rows.Err()
}

func parseProjectLinks(ns sql.NullString) []ProjectLink {
	out := []ProjectLink{}
	if !ns.Valid || ns.String == "" {
		return out
	}
	_ = json.Unmarshal([]byte(ns.String), &out)
	if out == nil {
		out = []ProjectLink{}
	}
	return out
}

// ListAdminProjects returns all projects (incl. drafts) as details, keyset
// paginated by id desc. status filters by publish status when non-empty.
func (s *Store) ListAdminProjects(ctx context.Context, status string, cursor string, limit int) ([]ProjectDetail, *string, error) {
	conds := []string{"1=1"}
	var args []any
	if status != "" {
		conds = append(conds, "p.status = ?")
		args = append(args, status)
	}
	if cursor != "" {
		if cid, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			conds = append(conds, "p.id < ?")
			args = append(args, cid)
		}
	}
	query := `SELECT ` + projectDetailCols + projectFrom +
		" WHERE " + strings.Join(conds, " AND ") + " ORDER BY p.id DESC LIMIT ?"
	args = append(args, limit+1)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []ProjectDetail
	for rows.Next() {
		var d ProjectDetail
		var sumCS, sumEN, projDate sql.NullString
		var featured int
		var cat Category
		var catVisible int
		var covID sql.NullInt64
		var covURL, covAltCS, covAltEN sql.NullString
		var covW, covH sql.NullInt64
		var bodyCS, bodyEN, linksJSON sql.NullString
		if err := rows.Scan(
			&d.ID, &d.Slug, &d.TitleCS, &d.TitleEN, &sumCS, &sumEN,
			&featured, &d.Status, &projDate, &d.SortOrder,
			&cat.ID, &cat.Slug, &cat.NameCS, &cat.NameEN, &cat.SortOrder, &catVisible,
			&covID, &covURL, &covAltCS, &covAltEN, &covW, &covH,
			&bodyCS, &bodyEN, &linksJSON, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		d.SummaryCS, d.SummaryEN, d.ProjectDate = nsPtr(sumCS), nsPtr(sumEN), nsPtr(projDate)
		d.Featured = featured != 0
		cat.Visible = catVisible != 0
		d.Category = cat
		if covID.Valid {
			d.Cover = &MediaRef{ID: covID.Int64, PublicURL: covURL.String,
				AltCS: nsPtr(covAltCS), AltEN: nsPtr(covAltEN), Width: niPtr(covW), Height: niPtr(covH)}
		}
		d.BodyCS, d.BodyEN = nsPtr(bodyCS), nsPtr(bodyEN)
		d.Links = parseProjectLinks(linksJSON)
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var next *string
	if len(items) > limit {
		items = items[:limit]
		c := strconv.FormatInt(items[len(items)-1].ID, 10)
		next = &c
	}
	// Fill galleries for the whole page in a single batched query.
	ids := make([]int64, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	galleries, err := s.galleriesFor(ctx, s.db, ids)
	if err != nil {
		return nil, nil, err
	}
	for i := range items {
		g := galleries[items[i].ID]
		if g == nil {
			g = []MediaRef{}
		}
		items[i].Gallery = g
	}
	return items, next, nil
}

func (s *Store) ProjectSlugExists(ctx context.Context, q DBTX, slug string, excludeID int64) (bool, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM project WHERE slug = ? AND id != ?`, slug, excludeID).Scan(&n)
	return n > 0, err
}

func (s *Store) MaxProjectSort(ctx context.Context, q DBTX) (int64, error) {
	var m sql.NullInt64
	err := q.QueryRowContext(ctx, `SELECT MAX(sort_order) FROM project`).Scan(&m)
	return m.Int64, err
}

func (s *Store) InsertProject(ctx context.Context, tx DBTX, in ProjectCreate, sortOrder int64, featured bool, status string, linksJSON string, coverID any) (int64, error) {
	now := nowUTC()
	res, err := tx.ExecContext(ctx,
		`INSERT INTO project
		   (slug, category_id, title_cs, title_en, summary_cs, summary_en, body_cs, body_en,
		    cover_media_id, links, project_date, featured, status, sort_order, created_at, updated_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		in.Slug, in.CategoryID, in.TitleCS, in.TitleEN, nullable(in.SummaryCS), nullable(in.SummaryEN),
		nullable(in.BodyCS), nullable(in.BodyEN), coverID, nullable(linksJSON), nullablePtr(in.ProjectDate),
		boolInt(featured), status, sortOrder, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// projectPatch is the resolved SET map for an UpdateProject.
func (s *Store) UpdateProject(ctx context.Context, tx DBTX, id int64, sets map[string]any) error {
	if len(sets) == 0 {
		return nil
	}
	cols := make([]string, 0, len(sets)+1)
	args := make([]any, 0, len(sets)+2)
	// Deterministic order not required by SQLite; iterate the map.
	for c, v := range sets {
		cols = append(cols, c+" = ?")
		args = append(args, v)
	}
	cols = append(cols, "updated_at = ?")
	args = append(args, nowUTC(), id)
	_, err := tx.ExecContext(ctx, `UPDATE project SET `+strings.Join(cols, ", ")+` WHERE id = ?`, args...)
	return err
}

func (s *Store) DeleteProject(ctx context.Context, q DBTX, id int64) (int64, error) {
	res, err := q.ExecContext(ctx, `DELETE FROM project WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// SetGallery replaces a project's gallery join rows with mediaIDs in order.
func (s *Store) SetGallery(ctx context.Context, tx DBTX, projectID int64, mediaIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM project_media WHERE project_id = ?`, projectID); err != nil {
		return err
	}
	for i, mid := range mediaIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO project_media (project_id, media_id, sort_order) VALUES (?,?,?)`,
			projectID, mid, i); err != nil {
			return err
		}
	}
	return nil
}

// MediaIDsExist reports whether every id in ids exists in media. Duplicate ids
// are collapsed first: the same media may legitimately be referenced twice (e.g.
// used as both the cover and a gallery image), and counting duplicates would make
// a fully-present set look partially missing.
func (s *Store) MediaIDsExist(ctx context.Context, q DBTX, ids []int64) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	seen := make(map[int64]struct{}, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		args = append(args, id)
	}
	var n int
	if err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM media WHERE id IN (`+placeholders(len(args))+`)`, args...).Scan(&n); err != nil {
		return false, err
	}
	return n == len(args), nil
}

// ==========================================================================
// generic transactional reorder (projects, categories, links, skills)
// ==========================================================================

// Reorder rewrites sort_order on table to match the given id order (0-based).
func (s *Store) Reorder(ctx context.Context, tx DBTX, table string, ids []int64) error {
	for i, id := range ids {
		if _, err := tx.ExecContext(ctx,
			`UPDATE `+table+` SET sort_order = ? WHERE id = ?`, i, id); err != nil {
			return err
		}
	}
	return nil
}

// ==========================================================================
// link
// ==========================================================================

const linkCols = `id, label_cs, label_en, url, icon, visible, sort_order, click_count`

func scanLink(r interface{ Scan(...any) error }) (Link, error) {
	var l Link
	var icon sql.NullString
	var visible int
	if err := r.Scan(&l.ID, &l.LabelCS, &l.LabelEN, &l.URL, &icon, &visible, &l.SortOrder, &l.ClickCount); err != nil {
		return Link{}, err
	}
	l.Icon = nsPtr(icon)
	l.Visible = visible != 0
	return l, nil
}

func (s *Store) ListLinks(ctx context.Context, visibleOnly bool) ([]Link, error) {
	where := ""
	if visibleOnly {
		where = " WHERE visible = 1"
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+linkCols+` FROM link`+where+` ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Link{}
	for rows.Next() {
		l, err := scanLink(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) GetLink(ctx context.Context, q DBTX, id int64) (*Link, error) {
	row := q.QueryRowContext(ctx, `SELECT `+linkCols+` FROM link WHERE id = ?`, id)
	l, err := scanLink(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *Store) MaxLinkSort(ctx context.Context, q DBTX) (int64, error) {
	var m sql.NullInt64
	err := q.QueryRowContext(ctx, `SELECT MAX(sort_order) FROM link`).Scan(&m)
	return m.Int64, err
}

func (s *Store) InsertLink(ctx context.Context, q DBTX, in LinkCreate, sortOrder int64, visible bool) (int64, error) {
	res, err := q.ExecContext(ctx,
		`INSERT INTO link (label_cs, label_en, url, icon, visible, sort_order, click_count, created_at)
		 VALUES (?,?,?,?,?,?,0,?)`,
		in.LabelCS, in.LabelEN, in.URL, nullablePtr(in.Icon), boolInt(visible), sortOrder, nowUTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateLink(ctx context.Context, q DBTX, id int64, in LinkUpdate) error {
	var sets []string
	var args []any
	if in.LabelCS != nil {
		sets, args = append(sets, "label_cs = ?"), append(args, *in.LabelCS)
	}
	if in.LabelEN != nil {
		sets, args = append(sets, "label_en = ?"), append(args, *in.LabelEN)
	}
	if in.URL != nil {
		sets, args = append(sets, "url = ?"), append(args, *in.URL)
	}
	if in.Icon != nil {
		sets, args = append(sets, "icon = ?"), append(args, nullablePtr(in.Icon))
	}
	if in.Visible != nil {
		sets, args = append(sets, "visible = ?"), append(args, boolInt(*in.Visible))
	}
	if in.SortOrder != nil {
		sets, args = append(sets, "sort_order = ?"), append(args, *in.SortOrder)
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	_, err := q.ExecContext(ctx, `UPDATE link SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...)
	return err
}

func (s *Store) DeleteLink(ctx context.Context, q DBTX, id int64) (int64, error) {
	res, err := q.ExecContext(ctx, `DELETE FROM link WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// IncrementClick bumps click_count. Returns rows affected (0 = unknown id).
func (s *Store) IncrementClick(ctx context.Context, id int64) (int64, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE link SET click_count = click_count + 1 WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ==========================================================================
// page section
// ==========================================================================

func (s *Store) GetSection(ctx context.Context, q DBTX, key string) (*Section, error) {
	row := q.QueryRowContext(ctx,
		`SELECT s.key, s.heading_cs, s.heading_en, s.body_cs, s.body_en, s.media_id, s.updated_at,
		        m.public_url, m.alt_cs, m.alt_en, m.width, m.height
		 FROM page_section s LEFT JOIN media m ON m.id = s.media_id WHERE s.key = ?`, key)
	var sec Section
	var hCS, hEN, bCS, bEN sql.NullString
	var mediaID sql.NullInt64
	var mURL, mAltCS, mAltEN sql.NullString
	var mW, mH sql.NullInt64
	if err := row.Scan(&sec.Key, &hCS, &hEN, &bCS, &bEN, &mediaID, &sec.UpdatedAt,
		&mURL, &mAltCS, &mAltEN, &mW, &mH); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	sec.HeadingCS, sec.HeadingEN, sec.BodyCS, sec.BodyEN = nsPtr(hCS), nsPtr(hEN), nsPtr(bCS), nsPtr(bEN)
	if mediaID.Valid {
		sec.Media = &MediaRef{ID: mediaID.Int64, PublicURL: mURL.String,
			AltCS: nsPtr(mAltCS), AltEN: nsPtr(mAltEN), Width: niPtr(mW), Height: niPtr(mH)}
	}
	return &sec, nil
}

// UpsertSection updates the given columns for key (creating the row if absent).
func (s *Store) UpsertSection(ctx context.Context, q DBTX, key string, sets map[string]any) error {
	// Ensure the row exists (about/hero are seeded; other keys are created here).
	if _, err := q.ExecContext(ctx,
		`INSERT OR IGNORE INTO page_section (key, updated_at) VALUES (?, ?)`, key, nowUTC()); err != nil {
		return err
	}
	cols := make([]string, 0, len(sets)+1)
	args := make([]any, 0, len(sets)+2)
	for c, v := range sets {
		cols = append(cols, c+" = ?")
		args = append(args, v)
	}
	cols = append(cols, "updated_at = ?")
	args = append(args, nowUTC(), key)
	_, err := q.ExecContext(ctx, `UPDATE page_section SET `+strings.Join(cols, ", ")+` WHERE key = ?`, args...)
	return err
}

// ==========================================================================
// skill
// ==========================================================================

const skillCols = `id, name_cs, name_en, category, icon, level, visible, sort_order`

func scanSkill(r interface{ Scan(...any) error }) (Skill, error) {
	var sk Skill
	var icon sql.NullString
	var level sql.NullInt64
	var visible int
	if err := r.Scan(&sk.ID, &sk.NameCS, &sk.NameEN, &sk.Category, &icon, &level, &visible, &sk.SortOrder); err != nil {
		return Skill{}, err
	}
	sk.Icon = nsPtr(icon)
	sk.Level = niPtr(level)
	sk.Visible = visible != 0
	return sk, nil
}

func (s *Store) ListSkills(ctx context.Context, visibleOnly bool) ([]Skill, error) {
	where := ""
	if visibleOnly {
		where = " WHERE visible = 1"
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+skillCols+` FROM skill`+where+` ORDER BY category, sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Skill{}
	for rows.Next() {
		sk, err := scanSkill(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sk)
	}
	return out, rows.Err()
}

func (s *Store) GetSkill(ctx context.Context, q DBTX, id int64) (*Skill, error) {
	row := q.QueryRowContext(ctx, `SELECT `+skillCols+` FROM skill WHERE id = ?`, id)
	sk, err := scanSkill(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sk, nil
}

func (s *Store) MaxSkillSort(ctx context.Context, q DBTX) (int64, error) {
	var m sql.NullInt64
	err := q.QueryRowContext(ctx, `SELECT MAX(sort_order) FROM skill`).Scan(&m)
	return m.Int64, err
}

func (s *Store) InsertSkill(ctx context.Context, q DBTX, in SkillCreate, sortOrder int64, visible bool) (int64, error) {
	var level any
	if in.Level != nil {
		level = *in.Level
	}
	res, err := q.ExecContext(ctx,
		`INSERT INTO skill (name_cs, name_en, category, icon, level, visible, sort_order) VALUES (?,?,?,?,?,?,?)`,
		in.NameCS, in.NameEN, in.Category, nullablePtr(in.Icon), level, boolInt(visible), sortOrder)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateSkill(ctx context.Context, q DBTX, id int64, in SkillUpdate) error {
	var sets []string
	var args []any
	if in.NameCS != nil {
		sets, args = append(sets, "name_cs = ?"), append(args, *in.NameCS)
	}
	if in.NameEN != nil {
		sets, args = append(sets, "name_en = ?"), append(args, *in.NameEN)
	}
	if in.Category != nil {
		sets, args = append(sets, "category = ?"), append(args, *in.Category)
	}
	if in.Icon != nil {
		sets, args = append(sets, "icon = ?"), append(args, nullablePtr(in.Icon))
	}
	if in.Level.Present {
		var v any
		if in.Level.Value != nil {
			v = *in.Level.Value
		}
		sets, args = append(sets, "level = ?"), append(args, v)
	}
	if in.Visible != nil {
		sets, args = append(sets, "visible = ?"), append(args, boolInt(*in.Visible))
	}
	if in.SortOrder != nil {
		sets, args = append(sets, "sort_order = ?"), append(args, *in.SortOrder)
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	_, err := q.ExecContext(ctx, `UPDATE skill SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...)
	return err
}

func (s *Store) DeleteSkill(ctx context.Context, q DBTX, id int64) (int64, error) {
	res, err := q.ExecContext(ctx, `DELETE FROM skill WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ==========================================================================
// business info (singleton id=1)
// ==========================================================================

func (s *Store) GetBusinessInfo(ctx context.Context, q DBTX) (*BusinessInfo, error) {
	row := q.QueryRowContext(ctx,
		`SELECT legal_name, ico, dic, address_lines, city, zip, country,
		        register_note_cs, register_note_en, contact_email, contact_phone,
		        data_note_cs, data_note_en, updated_at
		 FROM business_info WHERE id = 1`)
	var b BusinessInfo
	var ico, dic, addr, city, zip, country, rnCS, rnEN, email, phone, dnCS, dnEN sql.NullString
	if err := row.Scan(&b.LegalName, &ico, &dic, &addr, &city, &zip, &country,
		&rnCS, &rnEN, &email, &phone, &dnCS, &dnEN, &b.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	b.ICO, b.DIC, b.AddressLines, b.City, b.Zip, b.Country = nsPtr(ico), nsPtr(dic), nsPtr(addr), nsPtr(city), nsPtr(zip), nsPtr(country)
	b.RegisterNoteCS, b.RegisterNoteEN = nsPtr(rnCS), nsPtr(rnEN)
	b.ContactEmail, b.ContactPhone = nsPtr(email), nsPtr(phone)
	b.DataNoteCS, b.DataNoteEN = nsPtr(dnCS), nsPtr(dnEN)
	return &b, nil
}

func (s *Store) UpdateBusinessInfo(ctx context.Context, q DBTX, sets map[string]any) error {
	if _, err := q.ExecContext(ctx,
		`INSERT OR IGNORE INTO business_info (id, legal_name, updated_at) VALUES (1, '', ?)`, nowUTC()); err != nil {
		return err
	}
	if len(sets) == 0 {
		return nil
	}
	cols := make([]string, 0, len(sets)+1)
	args := make([]any, 0, len(sets)+1)
	for c, v := range sets {
		cols = append(cols, c+" = ?")
		args = append(args, v)
	}
	cols = append(cols, "updated_at = ?")
	args = append(args, nowUTC())
	_, err := q.ExecContext(ctx, `UPDATE business_info SET `+strings.Join(cols, ", ")+` WHERE id = 1`, args...)
	return err
}
