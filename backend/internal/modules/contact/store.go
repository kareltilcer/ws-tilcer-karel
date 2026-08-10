package contact

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
)

// Store is the contact module's SQL access layer.
type Store struct{ db *sql.DB }

// NewStore returns a contact store over db.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

const tsFormat = "2006-01-02T15:04:05.000000Z07:00"

func nowUTC() string { return time.Now().UTC().Format(tsFormat) }

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nsPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

const messageCols = `id, name, email, subject, message, locale, ip, user_agent, status, created_at`

func scanMessage(r interface{ Scan(...any) error }) (ContactMessage, error) {
	var m ContactMessage
	var subject, locale, ip, ua sql.NullString
	if err := r.Scan(&m.ID, &m.Name, &m.Email, &subject, &m.Message, &locale, &ip, &ua, &m.Status, &m.CreatedAt); err != nil {
		return ContactMessage{}, err
	}
	m.Subject, m.Locale, m.IP, m.UserAgent = nsPtr(subject), nsPtr(locale), nsPtr(ip), nsPtr(ua)
	return m, nil
}

// Insert stores a submission with status 'new' and returns its id.
func (s *Store) Insert(ctx context.Context, in ContactSubmit, ip, userAgent string) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO contact_message (name, email, subject, message, locale, ip, user_agent, status, created_at)
		 VALUES (?,?,?,?,?,?,?, 'new', ?)`,
		in.Name, in.Email, nullStr(in.Subject), in.Message, nullStr(in.Locale),
		nullStr(ip), nullStr(userAgent), nowUTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) Get(ctx context.Context, id int64) (*ContactMessage, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+messageCols+` FROM contact_message WHERE id = ?`, id)
	m, err := scanMessage(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// List returns messages newest-first, keyset-paginated by id desc, optionally
// filtered by status.
func (s *Store) List(ctx context.Context, status, cursor string, limit int) ([]ContactMessage, *string, error) {
	conds := []string{"1=1"}
	var args []any
	if status != "" {
		conds = append(conds, "status = ?")
		args = append(args, status)
	}
	if cursor != "" {
		if cid, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			conds = append(conds, "id < ?")
			args = append(args, cid)
		}
	}
	args = append(args, limit+1)
	q := `SELECT ` + messageCols + ` FROM contact_message WHERE ` + join(conds) + ` ORDER BY id DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	items := []ContactMessage{}
	for rows.Next() {
		m, err := scanMessage(rows)
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

func (s *Store) UpdateStatus(ctx context.Context, id int64, status string) (int64, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE contact_message SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *Store) Delete(ctx context.Context, id int64) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM contact_message WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func join(conds []string) string {
	out := conds[0]
	for _, c := range conds[1:] {
		out += " AND " + c
	}
	return out
}
