package contact

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/email"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/spam"
)

var validStatuses = map[string]bool{"new": true, "read": true, "archived": true, "spam": true}

// Service is the contact module's business logic: validate → spam-check → store
// → email (non-fatal), plus the admin inbox operations.
type Service struct {
	db          *sql.DB
	store       *Store
	guard       *spam.Guard
	sender      email.Sender // nil when Resend is not configured
	contactTo   string
	contactFrom string
	logger      *slog.Logger
}

// NewService builds the contact service. sender may be nil (submissions are then
// stored but not emailed — logged).
func NewService(db *sql.DB, guard *spam.Guard, sender email.Sender, contactTo, contactFrom string, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{db: db, store: NewStore(db), guard: guard, sender: sender, contactTo: contactTo, contactFrom: contactFrom, logger: logger}
}

// Submit validates, spam-checks, stores, and emails a contact submission. The
// stored row survives an email failure (the caller still returns 202). Rate
// limiting is enforced by the handler (which sets Retry-After).
func (s *Service) Submit(ctx context.Context, in ContactSubmit, ip, userAgent string) error {
	// 1. validate (422)
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)
	in.Message = strings.TrimSpace(in.Message)
	if in.Name == "" || in.Message == "" {
		return httpx.ErrUnprocessable("name and message are required")
	}
	if _, err := mail.ParseAddress(in.Email); err != nil {
		return httpx.ErrUnprocessable("a valid email is required")
	}
	if in.Locale != "" && in.Locale != "cs" && in.Locale != "en" {
		return httpx.ErrUnprocessable("locale must be cs or en")
	}

	// 2. spam guard: honeypot + Turnstile (400)
	if err := s.guard.Check(ctx, in.Website, in.TurnstileToken, ip); err != nil {
		return httpx.ErrSpamRejected("failed spam check")
	}

	// 3. store first — never lose the message.
	if _, err := s.store.Insert(ctx, in, ip, userAgent); err != nil {
		return err
	}

	// 4. email second — a failure is logged, not returned (still 202).
	s.sendEmail(ctx, in, ip)
	return nil
}

func (s *Service) sendEmail(ctx context.Context, in ContactSubmit, ip string) {
	if s.sender == nil || s.contactTo == "" || s.contactFrom == "" {
		s.logger.Warn("contact email not sent: Resend not configured", "from_email", in.Email)
		return
	}
	subject := in.Subject
	if subject == "" {
		subject = "Nová zpráva z karel.tilcer.cz"
	}
	body := fmt.Sprintf("Od: %s <%s>\nLocale: %s\nIP: %s\n\n%s\n", in.Name, in.Email, in.Locale, ip, in.Message)
	err := s.sender.Send(ctx, email.Message{
		From:    s.contactFrom,
		To:      s.contactTo,
		Subject: subject,
		Text:    body,
		ReplyTo: in.Email,
	})
	if err != nil {
		// The message is already stored; surface the failure in logs for follow-up.
		s.logger.Error("contact email send failed (message stored, not lost)", "err", err, "from_email", in.Email)
	}
}

// ---- admin inbox ----

func (s *Service) List(ctx context.Context, status, cursor string, limit int) (*ContactPage, error) {
	if status != "" && !validStatuses[status] {
		return nil, httpx.ErrUnprocessable("invalid status filter")
	}
	items, next, err := s.store.List(ctx, status, cursor, limit)
	if err != nil {
		return nil, err
	}
	return &ContactPage{Items: items, NextCursor: next}, nil
}

func (s *Service) SetStatus(ctx context.Context, id int64, status string) (*ContactMessage, error) {
	if !validStatuses[status] {
		return nil, httpx.ErrUnprocessable("invalid status")
	}
	n, err := s.store.UpdateStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, httpx.ErrNotFound("message not found")
	}
	return s.store.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	n, err := s.store.Delete(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.ErrNotFound("message not found")
	}
	return nil
}
