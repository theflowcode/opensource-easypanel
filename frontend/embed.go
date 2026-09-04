package frontend

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// DistFS contains the embedded single-binary production frontend assets.
//
//go:embed all:dist
var DistFS embed.FS

// SubFS returns a sub-filesystem anchored at "dist", allowing callers
// to access index.html, favicon.ico, and assets/ directly without prefixing "dist/".
func SubFS() (fs.FS, error) {
	return fs.Sub(DistFS, "dist")
}

// IndexHTML returns the raw bytes of the production index.html file.
func IndexHTML() ([]byte, error) {
	sub, err := SubFS()
	if err != nil {
		return nil, err
	}
	f, err := sub.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// HasAssets returns true if the embedded distribution contains actual build
// assets (e.g. index.html) beyond just the empty .gitkeep anchor.
func HasAssets() bool {
	sub, err := SubFS()
	if err != nil {
		return false
	}
	f, err := sub.Open("index.html")
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// Handler returns an http.Handler that serves the embedded SPA assets.
//
// It provides:
// 1. Direct file serving for existing assets (e.g., /assets/*, /favicon.ico).
// 2. Aggressive cache headers (max-age=31536000, immutable) for hashed assets in /assets/.
// 3. SPA fallback to index.html for client-side routing (e.g., /projects, /settings).
func Handler() http.Handler {
	sub, err := SubFS()
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "frontend assets unavailable", http.StatusInternalServerError)
		})
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle GET and HEAD requests for static SPA assets
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		cleanPath := path.Clean(r.URL.Path)
		cleanPath = strings.TrimPrefix(cleanPath, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		// Check if file exists directly in the embedded filesystem
		f, err := sub.Open(cleanPath)
		if err == nil {
			stat, statErr := f.Stat()
			_ = f.Close()

			if statErr == nil && !stat.IsDir() {
				// Apply caching policies
				if strings.HasPrefix(cleanPath, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else if cleanPath == "index.html" {
					w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
				}
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// If path has an extension (like .js, .css, .png) and wasn't found, return 404
		if path.Ext(cleanPath) != "" && !strings.HasSuffix(cleanPath, ".html") {
			http.NotFound(w, r)
			return
		}

		// Client-side SPA routing fallback: serve index.html
		indexBytes, err := IndexHTML()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			_, _ = w.Write(indexBytes)
		}
	})
}
