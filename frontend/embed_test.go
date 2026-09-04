package frontend_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opensource-easypanel/openpanel/frontend"
)

func TestDistFS(t *testing.T) {
	sub, err := frontend.SubFS()
	if err != nil {
		t.Fatalf("SubFS() failed: %v", err)
	}

	// Verify .gitkeep or index.html exists
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		t.Fatalf("ReadDir(.) failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected embedded entries in dist, got none")
	}

	if frontend.HasAssets() {
		// Assets are extracted: verify index.html and essential icons
		indexBytes, err := frontend.IndexHTML()
		if err != nil {
			t.Fatalf("IndexHTML() error: %v", err)
		}
		if !strings.Contains(string(indexBytes), "<div id=\"root\"></div>") {
			t.Errorf("index.html missing root div, content: %s", string(indexBytes))
		}

		// Verify favicon.ico and logos
		for _, name := range []string{"favicon.ico", "logo_dark.svg", "logo_light.svg", "logomark.svg"} {
			f, err := sub.Open(name)
			if err != nil {
				t.Errorf("failed to open embedded %s: %v", name, err)
			} else {
				_ = f.Close()
			}
		}
	}
}

func TestHandler_SPAFallbackAndCaching(t *testing.T) {
	if !frontend.HasAssets() {
		t.Skip("skipping handler tests because frontend dist assets are not populated")
	}

	h := frontend.Handler()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedCode   int
		expectedHeader string
		expectedBody   string
	}{
		{
			name:           "Root path serves index.html",
			method:         http.MethodGet,
			path:           "/",
			expectedCode:   http.StatusOK,
			expectedHeader: "text/html",
			expectedBody:   "<div id=\"root\"></div>",
		},
		{
			name:           "SPA client route serves index.html",
			method:         http.MethodGet,
			path:           "/projects/my-project/services",
			expectedCode:   http.StatusOK,
			expectedHeader: "text/html",
			expectedBody:   "<div id=\"root\"></div>",
		},
		{
			name:           "Direct asset serves with immutable cache",
			method:         http.MethodGet,
			path:           "/assets/index-EFRsbNOO.css",
			expectedCode:   http.StatusOK,
			expectedHeader: "text/css",
		},
		{
			name:         "Non-existent asset file returns 404",
			method:       http.MethodGet,
			path:         "/assets/non-existent-chunk.js",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Disallowed method returns 405",
			method:       http.MethodPost,
			path:         "/",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)

			if w.Code != tc.expectedCode {
				t.Errorf("path %s: expected status %d, got %d", tc.path, tc.expectedCode, w.Code)
			}

			if tc.expectedHeader != "" {
				contentType := w.Header().Get("Content-Type")
				if !strings.Contains(contentType, tc.expectedHeader) {
					t.Errorf("path %s: expected content-type containing %q, got %q", tc.path, tc.expectedHeader, contentType)
				}
			}

			if tc.expectedBody != "" && !strings.Contains(w.Body.String(), tc.expectedBody) {
				t.Errorf("path %s: body missing expected substring %q", tc.path, tc.expectedBody)
			}
		})
	}
}
