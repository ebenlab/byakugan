package agentkit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStyleMatchesDemoStylesheet pins the embedded canonical stylesheet to
// the demo copy at testdata/demo/_shared/doc.css: `byakugan style` and the
// demo docs must never drift apart.
func TestStyleMatchesDemoStylesheet(t *testing.T) {
	demoPath := filepath.Join("..", "..", "testdata", "demo", "_shared", "doc.css")
	demo, err := os.ReadFile(demoPath)
	if err != nil {
		t.Fatalf("reading demo stylesheet: %v", err)
	}
	if !bytes.Equal(StyleCSS(), demo) {
		t.Fatalf("internal/agentkit/assets/doc.css and %s differ; they must stay byte-identical — update both together", demoPath)
	}
}

func TestStyleIsSelfContained(t *testing.T) {
	css := string(StyleCSS())
	if !strings.Contains(css, "body.doc") {
		t.Error("stylesheet missing body.doc base rule")
	}
	for _, banned := range []string{"@import", "url(http", "url(//"} {
		if strings.Contains(css, banned) {
			t.Errorf("stylesheet references external resource via %q", banned)
		}
	}
}

func TestRulesCoverCoreConventions(t *testing.T) {
	rules := string(Rules())
	for _, want := range []string{
		`body class="doc"`,
		".doc-head",
		".doc-foot",
		"kicker",
		"lede",
		"callout",
		"stats",
		"viewBox",
		`marker id="arr"`,
		"d-box",
		"d-flow",
		"d-async",
		"d-num",
		"byakugan style",
		"<title>",
		"no CDNs",
	} {
		if !strings.Contains(rules, want) {
			t.Errorf("rules.md missing %q", want)
		}
	}
}

func TestKindsAreStable(t *testing.T) {
	want := []string{"prd", "adr", "architecture", "overview"}
	got := Kinds()
	if len(got) != len(want) {
		t.Fatalf("Kinds() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Kinds() = %v, want %v", got, want)
		}
	}
	// Mutating the returned slice must not affect the package copy.
	got[0] = "mutated"
	if Kinds()[0] != "prd" {
		t.Error("Kinds() returns the internal slice; callers can corrupt it")
	}
}

func TestTemplates(t *testing.T) {
	for _, kind := range Kinds() {
		t.Run(kind, func(t *testing.T) {
			tpl, err := Template(kind)
			if err != nil {
				t.Fatalf("Template(%q): %v", kind, err)
			}
			text := string(tpl)
			for _, want := range []string{
				"byakugan rules",
				"byakugan style > docs/_shared/doc.css",
				"doc-head",
				"doc-foot",
				"<!doctype html>", // every template carries a miniature example
			} {
				if !strings.Contains(text, want) {
					t.Errorf("template %q missing %q", kind, want)
				}
			}
		})
	}
}

func TestTemplateUnknownKind(t *testing.T) {
	for _, kind := range []string{"", "rfc", "PRD", "adr.md", "../doc"} {
		if _, err := Template(kind); err == nil {
			t.Errorf("Template(%q) succeeded, want error", kind)
		}
	}
}
