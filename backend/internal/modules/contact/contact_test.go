package contact

import (
	"context"
	"errors"
	"testing"

	appdb "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/db"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/email"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/registry"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/spam"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/testsupport"
)

type mockSender struct {
	sent []email.Message
	err  error
}

func (m *mockSender) Send(_ context.Context, msg email.Message) error {
	m.sent = append(m.sent, msg)
	return m.err
}

func apiStatus(err error) int {
	var ae *httpx.APIError
	if errors.As(err, &ae) {
		return ae.Status
	}
	return 0
}

func newSvc(t *testing.T, sender email.Sender) *Service {
	t.Helper()
	db := testsupport.NewDB(t,
		registry.MigrationSource{Name: "platform", FS: appdb.MigrationsFS},
		registry.MigrationSource{Name: "contact", FS: MigrationsFS},
	)
	// Turnstile disabled (empty secret) — honeypot still enforced.
	return NewService(db, spam.NewGuard(""), sender, "to@example.com", "from@example.com", nil)
}

func validSubmit() ContactSubmit {
	return ContactSubmit{Name: "Karel", Email: "visitor@example.com", Message: "Ahoj!", Locale: "cs"}
}

func TestSubmitStoresAndEmails(t *testing.T) {
	sender := &mockSender{}
	s := newSvc(t, sender)
	ctx := context.Background()
	if err := s.Submit(ctx, validSubmit(), "1.2.3.4", "UA/1"); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if len(sender.sent) != 1 {
		t.Fatalf("emails sent = %d, want 1", len(sender.sent))
	}
	if sender.sent[0].ReplyTo != "visitor@example.com" {
		t.Errorf("reply-to = %q, want visitor@example.com", sender.sent[0].ReplyTo)
	}
	page, err := s.List(ctx, "", "", 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].Status != "new" {
		t.Fatalf("inbox = %+v, want 1 message with status new", page.Items)
	}
}

func TestSubmitHoneypotRejected(t *testing.T) {
	s := newSvc(t, &mockSender{})
	in := validSubmit()
	in.Website = "i-am-a-bot"
	if err := s.Submit(context.Background(), in, "1.2.3.4", ""); apiStatus(err) != 400 {
		t.Fatalf("honeypot: want 400, got err=%v", err)
	}
}

func TestSubmitInvalidEmailRejected(t *testing.T) {
	s := newSvc(t, &mockSender{})
	in := validSubmit()
	in.Email = "not-an-email"
	if err := s.Submit(context.Background(), in, "1.2.3.4", ""); apiStatus(err) != 422 {
		t.Fatalf("invalid email: want 422, got err=%v", err)
	}
}

func TestSubmitEmailFailureIsNonFatal(t *testing.T) {
	sender := &mockSender{err: errors.New("resend down")}
	s := newSvc(t, sender)
	ctx := context.Background()
	if err := s.Submit(ctx, validSubmit(), "1.2.3.4", ""); err != nil {
		t.Fatalf("email failure must not fail the submit: %v", err)
	}
	page, _ := s.List(ctx, "", "", 50)
	if len(page.Items) != 1 {
		t.Fatalf("message must still be stored on email failure, inbox = %d", len(page.Items))
	}
}

func TestInboxStatusFilterAndDelete(t *testing.T) {
	s := newSvc(t, &mockSender{})
	ctx := context.Background()
	if err := s.Submit(ctx, validSubmit(), "1.2.3.4", ""); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	page, _ := s.List(ctx, "", "", 50)
	id := page.Items[0].ID

	// invalid status → 422
	if _, err := s.SetStatus(ctx, id, "bogus"); apiStatus(err) != 422 {
		t.Fatalf("invalid status: want 422, got err=%v", err)
	}
	// set read
	m, err := s.SetStatus(ctx, id, "read")
	if err != nil || m.Status != "read" {
		t.Fatalf("SetStatus read: %v (m=%+v)", err, m)
	}
	// status filter
	readPage, _ := s.List(ctx, "read", "", 50)
	if len(readPage.Items) != 1 {
		t.Fatalf("read filter = %d, want 1", len(readPage.Items))
	}
	newPage, _ := s.List(ctx, "new", "", 50)
	if len(newPage.Items) != 0 {
		t.Fatalf("new filter = %d, want 0", len(newPage.Items))
	}
	// delete
	if err := s.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := s.Delete(ctx, id); apiStatus(err) != 404 {
		t.Fatalf("delete again: want 404, got err=%v", err)
	}
}
