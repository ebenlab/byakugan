// Package index scans a documentation root and builds an in-memory,
// JSON-serializable index of projects and pages used for the landing page,
// navigation tree, and client-side search.
package index

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// maxTextLen caps how much visible text per page is shipped to the search
// index, keeping /api/index.json light even for large doc sets.
const maxTextLen = 2000

// Page is one HTML document under the root.
type Page struct {
	// Path is the URL path of the page relative to the server root,
	// always using forward slashes (e.g. "payments/adr-001.html").
	Path     string    `json:"path"`
	Project  string    `json:"project"`
	Title    string    `json:"title"`
	Headings []string  `json:"headings,omitempty"`
	Text     string    `json:"text,omitempty"`
	ModTime  time.Time `json:"modTime"`
}

// Project groups the pages of one first-level subdirectory of the root.
// Pages sitting directly in the root belong to the pseudo-project "".
type Project struct {
	Name      string    `json:"name"`
	Pages     []Page    `json:"pages"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Snapshot is the immutable result of one scan, served as /api/index.json.
type Snapshot struct {
	Root        string    `json:"root"`
	Projects    []Project `json:"projects"`
	PageCount   int       `json:"pageCount"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// Index scans a root directory on demand and holds the latest Snapshot.
// It is safe for concurrent use: Rebuild swaps the snapshot atomically
// under a mutex while readers call Current.
type Index struct {
	root string

	mu   sync.RWMutex
	snap *Snapshot
}

// New returns an Index for root. Call Rebuild before the first Current.
func New(root string) *Index {
	return &Index{root: root}
}

// Root returns the absolute directory this index scans.
func (ix *Index) Root() string { return ix.root }

// Current returns the latest snapshot.
func (ix *Index) Current() *Snapshot {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	return ix.snap
}

// Rebuild rescans the root and atomically replaces the snapshot.
func (ix *Index) Rebuild() error {
	byProject := map[string][]Page{}

	err := filepath.WalkDir(ix.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries rather than failing the scan
		}
		name := d.Name()
		if d.IsDir() {
			if path != ix.root && (strings.HasPrefix(name, ".") || name == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".html" && ext != ".htm" {
			return nil
		}
		rel, err := filepath.Rel(ix.root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		project := ""
		if i := strings.Index(rel, "/"); i >= 0 {
			project = rel[:i]
		}

		page := Page{Path: rel, Project: project}
		if info, err := d.Info(); err == nil {
			page.ModTime = info.ModTime()
		}
		if f, err := os.Open(path); err == nil {
			page.Title, page.Headings, page.Text = extract(f)
			f.Close()
		}
		if page.Title == "" {
			page.Title = strings.TrimSuffix(name, filepath.Ext(name))
		}
		byProject[project] = append(byProject[project], page)
		return nil
	})
	if err != nil {
		return err
	}

	snap := &Snapshot{Root: ix.root, GeneratedAt: time.Now()}
	names := make([]string, 0, len(byProject))
	for name := range byProject {
		names = append(names, name)
	}
	sort.Strings(names) // "" (root pages) naturally sorts first

	for _, name := range names {
		pages := byProject[name]
		sort.Slice(pages, func(i, j int) bool { return pages[i].Path < pages[j].Path })
		proj := Project{Name: name, Pages: pages}
		for _, p := range pages {
			if p.ModTime.After(proj.UpdatedAt) {
				proj.UpdatedAt = p.ModTime
			}
		}
		snap.Projects = append(snap.Projects, proj)
		snap.PageCount += len(pages)
	}

	ix.mu.Lock()
	ix.snap = snap
	ix.mu.Unlock()
	return nil
}

// extract pulls the title, headings, and a bounded amount of visible text
// out of an HTML document. It tolerates malformed HTML — x/net/html parses
// anything — and never returns an error; missing pieces come back empty.
func extract(r interface{ Read([]byte) (int, error) }) (title string, headings []string, text string) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", nil, ""
	}

	var sb strings.Builder
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "noscript", "template":
				return
			case "title":
				if title == "" {
					title = strings.TrimSpace(textContent(n))
				}
				return
			case "h1", "h2", "h3":
				if h := strings.TrimSpace(textContent(n)); h != "" && len(headings) < 50 {
					headings = append(headings, h)
				}
			}
		}
		if n.Type == html.TextNode && sb.Len() < maxTextLen {
			if t := strings.TrimSpace(n.Data); t != "" {
				sb.WriteString(t)
				sb.WriteByte(' ')
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	text = sb.String()
	if len(text) > maxTextLen {
		text = text[:maxTextLen]
	}
	return title, headings, strings.TrimSpace(text)
}

// textContent concatenates all text nodes under n.
func textContent(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}
