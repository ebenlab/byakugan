package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ebenlab/byakugan/internal/index"
)

func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	root := t.TempDir()
	page := `<html><head><title>ADR 1</title></head><body><h1>Decision</h1></body></html>`
	if err := os.MkdirAll(filepath.Join(root, "proj"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "proj", "adr.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	ix := index.New(root)
	if err := ix.Rebuild(); err != nil {
		t.Fatal(err)
	}
	return New(ix, "test"), root
}

func get(t *testing.T, s *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
	return rec
}

func TestLandingPage(t *testing.T) {
	s, _ := newTestServer(t)
	rec := get(t, s, "/")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Byakugan") {
		t.Fatalf("landing: code=%d", rec.Code)
	}
}

func TestHTMLInjection(t *testing.T) {
	s, _ := newTestServer(t)
	rec := get(t, s, "/proj/adr.html")
	body := rec.Body.String()
	if rec.Code != 200 {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(body, "/__byakugan/inject.js") {
		t.Error("overlay script not injected")
	}
	if !strings.Contains(body, "<h1>Decision</h1>") {
		t.Error("original content mangled")
	}
	// Snippet must land before </body>, not after.
	if strings.Index(body, "inject.js") > strings.Index(body, "</body>") {
		t.Error("snippet injected after </body>")
	}
}

func TestIndexJSON(t *testing.T) {
	s, _ := newTestServer(t)
	rec := get(t, s, "/api/index.json")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "ADR 1") {
		t.Fatalf("index.json: code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPathTraversalBlocked(t *testing.T) {
	s, _ := newTestServer(t)
	for _, p := range []string{"/../etc/passwd", "/proj/../../etc/passwd"} {
		rec := get(t, s, p)
		if rec.Code == 200 {
			t.Errorf("%s served, want rejection", p)
		}
	}
}

func TestDirectoryWithoutIndexServesLanding(t *testing.T) {
	s, _ := newTestServer(t)
	rec := get(t, s, "/proj/")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "bk-projects") {
		t.Fatalf("dir listing: code=%d", rec.Code)
	}
}

func TestInjectWithoutBodyTag(t *testing.T) {
	out := string(injectHTML([]byte("<p>no body tag</p>")))
	if !strings.Contains(out, "inject.js") {
		t.Error("snippet not appended")
	}
}
