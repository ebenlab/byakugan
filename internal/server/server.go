// Package server implements the Byakugan HTTP server: it serves the docs
// tree, generates the landing page, exposes the search index as JSON, and
// pushes live-reload events over SSE. Every served HTML page gets a small
// script injected that renders the navigation overlay.
package server

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ebenlab/byakugan/internal/agentkit"
	"github.com/ebenlab/byakugan/internal/index"
	"github.com/ebenlab/byakugan/internal/markdown"
	"github.com/ebenlab/byakugan/internal/scanrules"
)

//go:embed assets
var assets embed.FS

// injectSnippet is appended to every served HTML page. It loads the overlay
// UI (navigation drawer, search, prev/next) and the live-reload client.
const injectSnippet = `<script defer src="/__byakugan/inject.js"></script>`

// Server routes requests for one Index. Create it with New.
type Server struct {
	idx     *index.Index
	version string
	mux     *http.ServeMux

	mu   sync.Mutex
	subs map[chan string]struct{}
}

// New wires up all routes for idx.
func New(idx *index.Index, version string) *Server {
	s := &Server{
		idx:     idx,
		version: version,
		mux:     http.NewServeMux(),
		subs:    map[chan string]struct{}{},
	}
	s.mux.HandleFunc("GET /api/index.json", s.handleIndex)
	// The doc design system ships in the binary so rendered Markdown pages
	// (and any HTML page that wants it) work without a _shared/doc.css.
	s.mux.HandleFunc("GET /__byakugan/doc.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write(agentkit.StyleCSS())
	})
	s.mux.HandleFunc("GET /events", s.handleEvents)
	s.mux.Handle("GET /__byakugan/", http.StripPrefix("/__byakugan/", s.assetHandler()))
	s.mux.HandleFunc("GET /", s.handlePage)
	return s
}

// ListenAndServe blocks serving HTTP on addr. The header timeout guards the
// listener against stalled connections (relevant once bound to 0.0.0.0).
func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{Addr: addr, Handler: s.mux, ReadHeaderTimeout: 5 * time.Second}
	return srv.ListenAndServe()
}

// ServeHTTP makes Server usable directly in tests via httptest.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Broadcast pushes an SSE event (e.g. "reload") to every connected browser.
func (s *Server) Broadcast(event string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.subs {
		select {
		case ch <- event:
		default: // slow subscriber; drop rather than block the watcher
		}
	}
}

func (s *Server) assetHandler() http.Handler {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(fmt.Sprintf("embedded assets missing: %v", err))
	}
	return http.FileServerFS(sub)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(s.idx.Current())
}

// handleEvents keeps an SSE connection open and forwards Broadcast events.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan string, 4)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.subs, ch)
		s.mu.Unlock()
	}()

	fmt.Fprint(w, "event: hello\ndata: byakugan\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-ch:
			fmt.Fprintf(w, "event: %s\ndata: 1\n\n", ev)
			flusher.Flush()
		}
	}
}

// handlePage serves the landing page at "/", static files elsewhere, and
// injects the overlay snippet into every HTML response.
func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		s.serveAsset(w, "landing.html")
		return
	}

	clean := path.Clean(r.URL.Path)
	if strings.Contains(clean, "..") {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	full := filepath.Join(s.idx.Root(), filepath.FromSlash(clean))

	info, err := os.Stat(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if info.IsDir() {
		// Directories with an index.html serve it; others get the landing
		// page, which reads the URL and scopes itself to that project.
		idxPath := filepath.Join(full, "index.html")
		if _, err := os.Stat(idxPath); err == nil {
			s.serveHTMLFile(w, idxPath)
			return
		}
		s.serveAsset(w, "landing.html")
		return
	}

	ext := strings.ToLower(filepath.Ext(full))
	if ext == ".html" || ext == ".htm" {
		s.serveHTMLFile(w, full)
		return
	}
	if ext == ".md" {
		s.serveMarkdownFile(w, full, clean)
		return
	}
	http.ServeFile(w, r, full)
}

// serveHTMLFile writes an HTML file with the overlay snippet injected
// before </body> (or appended when no closing tag exists).
func (s *Server) serveHTMLFile(w http.ResponseWriter, fullPath string) {
	b, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "unreadable file", http.StatusInternalServerError)
		return
	}
	b = injectHTML(b)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(b)
}

// serveMarkdownFile renders a .md file into the doc.css skeleton and
// injects the overlay, making it an ordinary Byakugan page.
func (s *Server) serveMarkdownFile(w http.ResponseWriter, fullPath, urlPath string) {
	src, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "unreadable file", http.StatusInternalServerError)
		return
	}
	project := scanrules.ProjectOf(strings.TrimPrefix(urlPath, "/"))
	name := strings.TrimSuffix(filepath.Base(fullPath), filepath.Ext(fullPath))
	page, err := markdown.Page(src, name, project)
	if err != nil {
		http.Error(w, "markdown rendering failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(injectHTML(page))
}

func (s *Server) serveAsset(w http.ResponseWriter, name string) {
	b, err := assets.ReadFile("assets/" + name)
	if err != nil {
		http.Error(w, "asset missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(b)
}

// injectHTML inserts the snippet before the last </body>, falling back to
// appending so pages without a closing tag still get the overlay.
func injectHTML(b []byte) []byte {
	lower := bytes.ToLower(b)
	if i := bytes.LastIndex(lower, []byte("</body>")); i >= 0 {
		var out bytes.Buffer
		out.Grow(len(b) + len(injectSnippet))
		out.Write(b[:i])
		out.WriteString(injectSnippet)
		out.Write(b[i:])
		return out.Bytes()
	}
	return append(b, []byte(injectSnippet)...)
}
