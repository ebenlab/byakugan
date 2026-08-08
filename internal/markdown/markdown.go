// Package markdown renders .md documents into the shared doc.css page
// skeleton so agents' Markdown files become first-class Byakugan pages.
//
// Scope is deliberately narrow (see EBE-129): CommonMark + GFM tables,
// strikethrough, and autolinks; optional front matter capped at title,
// status, and date; no configuration, templates, syntax highlighting, or
// TOC generation. Byakugan is a docs server, not a static site generator.
package markdown

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var engine = goldmark.New(goldmark.WithExtensions(extension.GFM))

// Meta is the capped front matter set. Anything else in a front matter
// block is deliberately ignored.
type Meta struct {
	Title  string
	Status string
	Date   string
}

// Parse splits an optional leading front matter block ("---" fences) off
// src and returns the recognized keys plus the remaining Markdown body.
// The parse is intentionally naive — three flat "key: value" lines, no YAML
// dependency; a malformed block is treated as body text.
func Parse(src []byte) (Meta, []byte) {
	var meta Meta
	rest, ok := bytes.CutPrefix(src, []byte("---\n"))
	if !ok {
		if rest, ok = bytes.CutPrefix(src, []byte("---\r\n")); !ok {
			return meta, src
		}
	}
	head, body, ok := bytes.Cut(rest, []byte("\n---"))
	if !ok {
		return meta, src
	}
	// The closing fence must end its line.
	if trimmed := bytes.TrimLeft(body, "\r"); len(trimmed) > 0 {
		if trimmed[0] != '\n' {
			return meta, src
		}
		body = trimmed[1:]
	}
	for _, line := range strings.Split(string(head), "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		val = strings.TrimSpace(val)
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "title":
			meta.Title = val
		case "status":
			meta.Status = val
		case "date":
			meta.Date = val
		}
	}
	return meta, body
}

// FirstHeading returns the text of the first ATX "# " heading, or "".
func FirstHeading(body []byte) string {
	for _, line := range strings.Split(string(body), "\n") {
		trimmed := strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(trimmed, "# "); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// ToHTML converts a Markdown body to an HTML fragment.
func ToHTML(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := engine.Convert(body, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Title resolves the page title: front matter first, then the first "# "
// heading, then the file's base name without extension.
func Title(meta Meta, body []byte, name string) string {
	if meta.Title != "" {
		return meta.Title
	}
	if h := FirstHeading(body); h != "" {
		return h
	}
	return name
}

// Page renders src into a complete doc.css-styled HTML page. name is the
// file base name without extension (title fallback); project labels the
// kicker line. The stylesheet is served by the byakugan binary itself, so
// rendered pages work in folders with no _shared/doc.css.
func Page(src []byte, name, project string) ([]byte, error) {
	meta, body := Parse(src)
	frag, err := ToHTML(body)
	if err != nil {
		return nil, err
	}
	title := Title(meta, body, name)

	kicker := kickerHTML(meta, project)
	// The Markdown's own leading "# " heading renders inside <main>; only
	// synthesize an <h1> when the document doesn't bring one.
	var h1 string
	if FirstHeading(body) == "" {
		h1 = "<h1>" + html.EscapeString(title) + "</h1>\n"
	}

	var page bytes.Buffer
	fmt.Fprintf(&page, `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<link rel="stylesheet" href="/__byakugan/doc.css">
</head>
<body class="doc">
<header class="doc-head">
%s%s</header>
<main>
`, html.EscapeString(title), kicker, h1)
	page.Write(frag)
	page.WriteString(`</main>
<footer class="doc-foot">rendered from ` + html.EscapeString(name) + `.md · served by byakugan</footer>
</body>
</html>
`)
	return page.Bytes(), nil
}

// kickerHTML builds the mono meta line above the title.
func kickerHTML(meta Meta, project string) string {
	parts := []string{}
	if project != "" {
		parts = append(parts, html.EscapeString(project))
	}
	parts = append(parts, "markdown")
	if meta.Status != "" {
		parts = append(parts, `<span class="badge badge--`+badgeClass(meta.Status)+`">`+html.EscapeString(meta.Status)+`</span>`)
	}
	if meta.Date != "" {
		parts = append(parts, html.EscapeString(meta.Date))
	}
	return `<p class="kicker">` + strings.Join(parts, " · ") + "</p>\n"
}

// badgeClass maps a free-form status to one of doc.css's badge variants.
func badgeClass(status string) string {
	switch s := strings.ToLower(status); {
	case strings.Contains(s, "accept"), strings.Contains(s, "current"), strings.Contains(s, "living"), strings.Contains(s, "final"):
		return "accepted"
	case strings.Contains(s, "supersed"), strings.Contains(s, "deprecat"), strings.Contains(s, "reject"):
		return "superseded"
	default:
		return "draft"
	}
}
