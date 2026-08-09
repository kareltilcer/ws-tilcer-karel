package media

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"

	appcontent "github.com/kareltilcer/ws-tilcer-karel/backend/internal/modules/content"
	appdb "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/db"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/registry"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/testsupport"
)

// mockS3 records Put/Delete calls without a real bucket.
type mockS3 struct {
	put map[string][]byte
	del []string
}

func newMockS3() *mockS3 { return &mockS3{put: map[string][]byte{}} }

func (m *mockS3) Put(_ context.Context, key, _ string, body []byte) error {
	m.put[key] = body
	return nil
}
func (m *mockS3) Delete(_ context.Context, key string) error {
	m.del = append(m.del, key)
	return nil
}

func newTestService(t *testing.T, s3 *mockS3) (*Service, *sql.DB) {
	t.Helper()
	db := testsupport.NewDB(t,
		registry.MigrationSource{Name: "platform", FS: appdb.MigrationsFS},
		registry.MigrationSource{Name: "content", FS: appcontent.MigrationsFS},
	)
	svc := NewService(db, s3, "media/", "https://cdn.example/karel-media", 1<<20)
	return svc, db
}

func apiStatus(err error) int {
	var ae *httpx.APIError
	if errors.As(err, &ae) {
		return ae.Status
	}
	return 0
}

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func TestUploadPNGHappyPath(t *testing.T) {
	s3 := newMockS3()
	svc, _ := newTestService(t, s3)
	m, err := svc.Upload(context.Background(), "image/png", tinyPNG(t), strPtr("alt cs"), nil)
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if m.Mime != "image/png" {
		t.Errorf("mime = %q, want image/png", m.Mime)
	}
	if m.Width == nil || *m.Width != 3 || m.Height == nil || *m.Height != 2 {
		t.Errorf("dims = %v x %v, want 3x2", m.Width, m.Height)
	}
	if m.PublicURL != "https://cdn.example/karel-media/"+m.S3Key {
		t.Errorf("public url = %q (key %q)", m.PublicURL, m.S3Key)
	}
	if _, ok := s3.put[m.S3Key]; !ok {
		t.Errorf("object %q not stored in S3", m.S3Key)
	}
}

func TestUploadSVGSanitized(t *testing.T) {
	s3 := newMockS3()
	svc, _ := newTestService(t, s3)
	dirty := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script><rect width="1" height="1"/></svg>`)
	m, err := svc.Upload(context.Background(), "image/svg+xml", dirty, nil, nil)
	if err != nil {
		t.Fatalf("Upload svg: %v", err)
	}
	stored := s3.put[m.S3Key]
	if bytes.Contains(bytes.ToLower(stored), []byte("<script")) {
		t.Errorf("stored SVG still contains a script tag: %s", stored)
	}
}

func TestUploadUnsupportedType(t *testing.T) {
	s3 := newMockS3()
	svc, _ := newTestService(t, s3)
	if _, err := svc.Upload(context.Background(), "application/pdf", []byte("%PDF-1.4 not an image"), nil, nil); apiStatus(err) != 415 {
		t.Fatalf("unsupported type: want 415, got err=%v", err)
	}
}

func TestUploadOversize(t *testing.T) {
	s3 := newMockS3()
	svc, _ := newTestService(t, s3)
	svc.maxBytes = 10
	if _, err := svc.Upload(context.Background(), "image/png", tinyPNG(t), nil, nil); apiStatus(err) != 413 {
		t.Fatalf("oversize: want 413, got err=%v", err)
	}
}

func TestDeleteReferencedBlocked(t *testing.T) {
	s3 := newMockS3()
	db := testsupport.NewDB(t,
		registry.MigrationSource{Name: "platform", FS: appdb.MigrationsFS},
		registry.MigrationSource{Name: "content", FS: appcontent.MigrationsFS},
	)
	svc := NewService(db, s3, "media/", "https://cdn.example/karel-media", 1<<20)
	ctx := context.Background()

	m, err := svc.Upload(ctx, "image/png", tinyPNG(t), nil, nil)
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	// Reference the media from a project (raw insert: category + project cover).
	if _, err := db.ExecContext(ctx, `INSERT INTO category (slug, name_cs, name_en, created_at) VALUES ('c','C','C','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("insert category: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO project (slug, category_id, title_cs, title_en, cover_media_id, created_at, updated_at)
		 VALUES ('p', 1, 'T', 'T', ?, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`, m.ID); err != nil {
		t.Fatalf("insert project: %v", err)
	}

	if err := svc.Delete(ctx, m.ID); apiStatus(err) != 409 {
		t.Fatalf("delete referenced: want 409, got err=%v", err)
	}
	// Unknown id → 404.
	if err := svc.Delete(ctx, 9999); apiStatus(err) != 404 {
		t.Fatalf("delete unknown: want 404, got err=%v", err)
	}
}

func strPtr(s string) *string { return &s }
