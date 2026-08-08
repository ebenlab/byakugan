package markdown

import (
	"strings"
	"testing"
)

func TestParseFrontMatter(t *testing.T) {
	src := []byte("---\ntitle: Decision log\nstatus: living\ndate: 2026-08-08\nowner: ignored\n---\n# Heading\nBody.")
	meta, body := Parse(src)
	if meta.Title != "Decision log" || meta.Status != "living" || meta.Date != "2026-08-08" {
		t.Fatalf("meta = %+v", meta)
	}
	if !strings.HasPrefix(string(body), "# Heading") {
		t.Errorf("body = %q", body)
	}
}

func TestParseWithoutFrontMatter(t *testing.T) {
	src := []byte("# Just a doc\n\nText.")
	meta, body := Parse(src)
	if meta != (Meta{}) {
		t.Errorf("meta = %+v, want zero", meta)
	}
	if string(body) != string(src) {
		t.Errorf("body altered: %q", body)
	}
}

func TestParseUnclosedFenceIsBody(t *testing.T) {
	src := []byte("---\ntitle: nope\nno closing fence")
	meta, body := Parse(src)
	if meta.Title != "" || string(body) != string(src) {
		t.Errorf("malformed front matter should be body text; meta=%+v", meta)
	}
}

func TestTitleFallbackChain(t *testing.T) {
	if got := Title(Meta{Title: "From meta"}, []byte("# From heading"), "file"); got != "From meta" {
		t.Errorf("meta title: %q", got)
	}
	if got := Title(Meta{}, []byte("intro\n# From heading\n"), "file"); got != "From heading" {
		t.Errorf("heading title: %q", got)
	}
	if got := Title(Meta{}, []byte("no headings"), "file"); got != "file" {
		t.Errorf("name title: %q", got)
	}
}

func TestPageRendersGFM(t *testing.T) {
	src := []byte("---\ntitle: T\nstatus: accepted\ndate: 2026-08-08\n---\n# T\n\n| a | b |\n| - | - |\n| 1 | 2 |\n\n~~gone~~ https://example.com\n")
	page, err := Page(src, "t", "proj")
	if err != nil {
		t.Fatal(err)
	}
	s := string(page)
	for _, want := range []string{
		"<title>T</title>",
		`href="/__byakugan/doc.css"`,
		`<body class="doc">`,
		"<table>", "<del>gone</del>",
		`<a href="https://example.com"`,
		`badge--accepted`, "proj · markdown",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("page missing %q", want)
		}
	}
	if strings.Count(s, "<h1") != 1 {
		t.Errorf("want exactly one h1, got %d", strings.Count(s, "<h1"))
	}
}

func TestPageSynthesizesH1WhenAbsent(t *testing.T) {
	page, err := Page([]byte("just text"), "notes", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "<h1>notes</h1>") {
		t.Errorf("missing synthesized h1: %s", page)
	}
}

func TestPageEscapesRawHTMLTitle(t *testing.T) {
	page, err := Page([]byte("---\ntitle: <script>x</script>\n---\nbody"), "n", "<p>")
	if err != nil {
		t.Fatal(err)
	}
	s := string(page)
	if strings.Contains(s, "<script>x</script>") {
		t.Error("title not escaped")
	}
	if strings.Contains(s, "<p> · markdown") {
		t.Error("project not escaped")
	}
}
