package media

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
)

// Store is the media module's SQL access layer over the shared `media` table
// (defined in the content migration).
type Store struct{ db *sql.DB }

// NewStore returns a media store over db.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

const tsFormat = "2006-01-02T15:04:05.000000Z07:00"

func nowUTC() string { return time.Now().UTC().Format(tsFormat) }

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

func nullI64(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

func nullStrP(p *string) any {
	if p == nil || *p == "" {
		return nil
	}
	return *p
}

const mediaCols = `id, s3_key, public_url, mime, width, height, size_bytes, alt_cs, alt_en, created_at`

func scanMedia(r interface{ Scan(...any) error }) (Media, error) {
	var m Media
	var w, h sql.NullInt64
	var altCS, altEN sql.NullString
	if err := r.Scan(&m.ID, &m.S3Key, &m.PublicURL, &m.Mime, &w, &h, &m.SizeBytes, &altCS, &altEN, &m.CreatedAt); err != nil {
		return Media{}, err
	}
	m.Width, m.Height, m.AltCS, m.AltEN = niPtr(w), niPtr(h), nsPtr(altCS), nsPtr(altEN)
	return m, nil
}

func (s *Store) Insert(ctx context.Context, m Media) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO media (s3_key, public_url, mime, width, height, size_bytes, alt_cs, alt_en, created_at)
		 VALUES (?,?,?,?,?,?,?,?,?)`,
		m.S3Key, m.PublicURL, m.Mime, nullI64(m.Width), nullI64(m.Height), m.SizeBytes,
		nullStrP(m.AltCS), nullStrP(m.AltEN), nowUTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) Get(ctx context.Context, id int64) (*Media, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+mediaCols+` FROM media WHERE id = ?`, id)
	m, err := scanMedia(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// List returns media keyset-paginated by id desc (newest first). cursor is the
// last seen id.
func (s *Store) List(ctx context.Context, cursor string, limit int) ([]Media, *string, error) {
	args := []any{}
	where := ""
	if cursor != "" {
		if cid, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			where = " WHERE id < ?"
			args = append(args, cid)
		}
	}
	args = append(args, limit+1)
	rows, err := s.db.QueryContext(ctx, `SELECT `+mediaCols+` FROM media`+where+` ORDER BY id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	items := []Media{}
	for rows.Next() {
		m, err := scanMedia(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, m)
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
	return items, next, nil
}

func (s *Store) Delete(ctx context.Context, id int64) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM media WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// IsReferenced reports whether id is used as a project cover, in a project
// gallery, or as a page-section media.
func (s *Store) IsReferenced(ctx context.Context, id int64) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT
		   (SELECT COUNT(*) FROM project       WHERE cover_media_id = ?) +
		   (SELECT COUNT(*) FROM project_media WHERE media_id = ?) +
		   (SELECT COUNT(*) FROM page_section  WHERE media_id = ?)`,
		id, id, id).Scan(&n)
	return n > 0, err
}
