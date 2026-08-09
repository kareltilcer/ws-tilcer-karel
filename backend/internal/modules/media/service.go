package media

import (
	"context"
	"database/sql"
	"image"
	_ "image/gif"  // register decoders for DecodeConfig (dimensions)
	_ "image/jpeg" //
	_ "image/png"  //
	"net/http"
	"strings"

	"bytes"

	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/idgen"
	objstore "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/media"
)

// allowedMimes maps a canonical mime to its file extension.
var allowedMimes = map[string]string{
	"image/png":     ".png",
	"image/jpeg":    ".jpg",
	"image/gif":     ".gif",
	"image/webp":    ".webp",
	"image/svg+xml": ".svg",
}

// Service is the media module's business logic: validate → (sanitize SVG) →
// store in S3 → persist metadata; and reference-guarded delete.
type Service struct {
	db         *sql.DB
	store      *Store
	s3         objstore.Store // nil when media storage is not configured
	prefix     string         // object key prefix, e.g. "media/"
	publicBase string         // MEDIA_PUBLIC_BASE_URL
	maxBytes   int64
}

// NewService builds the media service. s3 may be nil (uploads then return 503).
func NewService(db *sql.DB, s3 objstore.Store, prefix, publicBase string, maxBytes int64) *Service {
	return &Service{db: db, store: NewStore(db), s3: s3, prefix: prefix, publicBase: publicBase, maxBytes: maxBytes}
}

func errUnavailable(detail string) *httpx.APIError {
	return &httpx.APIError{Status: http.StatusServiceUnavailable, Code: "media_unavailable", Detail: detail}
}

// Upload validates and stores an uploaded image, returning its metadata.
func (s *Service) Upload(ctx context.Context, declaredMime string, data []byte, altCS, altEN *string) (*Media, error) {
	if len(data) == 0 {
		return nil, httpx.ErrUnprocessable("empty file")
	}
	if int64(len(data)) > s.maxBytes {
		return nil, httpx.ErrPayloadTooLarge("file exceeds size limit")
	}
	mime := detectMime(data, declaredMime)
	if mime == "" {
		return nil, httpx.ErrUnsupportedMedia("unsupported media type")
	}

	body := data
	var width, height *int64
	if mime == "image/svg+xml" {
		clean, err := objstore.SanitizeSVG(data)
		if err != nil {
			return nil, httpx.ErrUnprocessable("failed to sanitize SVG: " + err.Error())
		}
		body = clean
	} else if w, h, ok := imageDims(data); ok {
		width, height = &w, &h
	}

	if s.s3 == nil {
		return nil, errUnavailable("media storage is not configured")
	}
	key := s.prefix + idgen.New() + allowedMimes[mime]
	if err := s.s3.Put(ctx, key, mime, body); err != nil {
		return nil, err
	}

	m := Media{
		S3Key:     key,
		PublicURL: objstore.PublicURL(s.publicBase, key),
		Mime:      mime,
		Width:     width,
		Height:    height,
		SizeBytes: int64(len(body)),
		AltCS:     altCS,
		AltEN:     altEN,
	}
	id, err := s.store.Insert(ctx, m)
	if err != nil {
		// Best-effort cleanup of the just-stored object so we don't orphan it.
		_ = s.s3.Delete(ctx, key)
		return nil, err
	}
	return s.store.Get(ctx, id)
}

// List returns a paginated media listing.
func (s *Service) List(ctx context.Context, cursor string, limit int) (*MediaPage, error) {
	items, next, err := s.store.List(ctx, cursor, limit)
	if err != nil {
		return nil, err
	}
	return &MediaPage{Items: items, NextCursor: next}, nil
}

// Delete removes a media object + row, blocked (409) while referenced.
func (s *Service) Delete(ctx context.Context, id int64) error {
	m, err := s.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return httpx.ErrNotFound("media not found")
	}
	referenced, err := s.store.IsReferenced(ctx, id)
	if err != nil {
		return err
	}
	if referenced {
		return httpx.ErrConflict("media is still referenced by a project or section")
	}
	if s.s3 != nil {
		if err := s.s3.Delete(ctx, m.S3Key); err != nil {
			return err // keep the row if the object could not be removed
		}
	}
	_, err = s.store.Delete(ctx, id)
	return err
}

// detectMime returns the canonical allowed mime for data, or "" if unsupported.
// SVG is detected from the declared type or an <svg> root; raster types are
// sniffed from the bytes so a mislabelled upload is caught.
func detectMime(data []byte, declared string) string {
	declared = strings.ToLower(strings.TrimSpace(declared))
	if declared == "image/svg+xml" || looksLikeSVG(data) {
		return "image/svg+xml"
	}
	// Raster types are decided by the sniffed bytes only — never by the declared
	// type — so a mislabelled (or disguised non-image) upload cannot be stored and
	// served under an image content-type. All allowed rasters carry strong magic
	// signatures that http.DetectContentType recognises.
	sniff := http.DetectContentType(data)
	switch {
	case strings.HasPrefix(sniff, "image/png"):
		return "image/png"
	case strings.HasPrefix(sniff, "image/jpeg"):
		return "image/jpeg"
	case strings.HasPrefix(sniff, "image/gif"):
		return "image/gif"
	case strings.HasPrefix(sniff, "image/webp"):
		return "image/webp"
	}
	return ""
}

func looksLikeSVG(data []byte) bool {
	head := data
	if len(head) > 1024 {
		head = head[:1024]
	}
	low := strings.ToLower(string(bytes.TrimSpace(head)))
	return strings.Contains(low, "<svg")
}

func imageDims(data []byte) (int64, int64, bool) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, false
	}
	return int64(cfg.Width), int64(cfg.Height), true
}
