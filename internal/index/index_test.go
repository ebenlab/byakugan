package index

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRebuild(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "payments", "arch.html"),
		`<html><head><title>Payments Architecture</title></head>
		 <body><h1>Ledger design</h1><p>Double-entry bookkeeping core.</p></body></html>`)
	writeFile(t, filepath.Join(root, "payments", "adr-001.html"),
		`<html><body><h2>Choose Postgres</h2></body></html>`)
	writeFile(t, filepath.Join(root, "overview.html"),
		`<html><head><title>Overview</title></head><body>hello</body></html>`)
	writeFile(t, filepath.Join(root, "payments", "notes.txt"), "not html")
	writeFile(t, filepath.Join(root, ".hidden", "secret.html"), "<html></html>")
	writeFile(t, filepath.Join(root, "node_modules", "dep.html"), "<html></html>")

	ix := New(root)
	if err := ix.Rebuild(); err != nil {
		t.Fatal(err)
	}
	snap := ix.Current()

	if snap.PageCount != 3 {
		t.Fatalf("PageCount = %d, want 3", snap.PageCount)
	}
	if len(snap.Projects) != 2 {
		t.Fatalf("Projects = %d, want 2 (root + payments)", len(snap.Projects))
	}

	// Root pseudo-project sorts first.
	if snap.Projects[0].Name != "" || snap.Projects[0].Pages[0].Title != "Overview" {
		t.Errorf("root project wrong: %+v", snap.Projects[0])
	}

	pay := snap.Projects[1]
	if pay.Name != "payments" || len(pay.Pages) != 2 {
		t.Fatalf("payments project wrong: %+v", pay)
	}
	// adr-001 has no <title>; falls back to the file name.
	if pay.Pages[0].Title != "adr-001" {
		t.Errorf("fallback title = %q, want adr-001", pay.Pages[0].Title)
	}
	arch := pay.Pages[1]
	if arch.Title != "Payments Architecture" {
		t.Errorf("title = %q", arch.Title)
	}
	if len(arch.Headings) != 1 || arch.Headings[0] != "Ledger design" {
		t.Errorf("headings = %v", arch.Headings)
	}
	if !strings.Contains(arch.Text, "Double-entry") {
		t.Errorf("text missing body content: %q", arch.Text)
	}
	if strings.Contains(arch.Text, "Payments Architecture") {
		t.Errorf("text should not repeat the <title>: %q", arch.Text)
	}
}

func TestExtractMalformedHTML(t *testing.T) {
	title, _, text := extract(strings.NewReader("<h1>Broken<p>but parsed"))
	if title != "" {
		t.Errorf("title = %q, want empty", title)
	}
	if !strings.Contains(text, "Broken") {
		t.Errorf("text = %q", text)
	}
}
