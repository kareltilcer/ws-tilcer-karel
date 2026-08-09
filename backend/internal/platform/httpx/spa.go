package httpx

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// SPAHandler serves the built single-page app from dir. A request that resolves
// to an existing file is served directly; any other path falls back to
// index.html so the client-side router can resolve deep links.
//
// It is only reached for paths that matched no server route. The router keeps
// /api/** out of this fallback (they get a JSON 404 instead) so a mistyped
// endpoint fails loudly rather than silently returning the SPA shell.
//
// In production karel ships the SPA as its own Nginx image, so StaticDir is
// unset and this handler is not installed; it exists for the docker-compose /
// single-image path and parity with the fleet.
func SPAHandler(dir string) http.HandlerFunc {
	fileServer := http.FileServer(http.Dir(dir))
	indexPath := filepath.Join(dir, "index.html")

	return func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		full := filepath.Join(dir, filepath.FromSlash(clean))

		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			if strings.HasPrefix(clean, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, indexPath)
	}
}
