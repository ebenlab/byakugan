// Package agentkit embeds the agent-facing assets the byakugan CLI exposes:
// the shared "blueprint editorial" doc stylesheet (`byakugan style`), the
// authoring rules guide (`byakugan rules`), and prompt templates for
// generating docs (`byakugan template <kind>`).
//
// The embedded assets/doc.css is the canonical copy of the design system;
// testdata/demo/_shared/doc.css must stay byte-identical to it (enforced by
// a unit test) so the demo never drifts from what `byakugan style` prints.
package agentkit

import (
	"embed"
	"fmt"
)

//go:embed assets
var assets embed.FS

// templateKinds lists the available template kinds in display order.
var templateKinds = []string{"prd", "adr", "architecture", "overview"}

// StyleCSS returns the design-system stylesheet that generated docs link as
// _shared/doc.css.
func StyleCSS() []byte { return mustAsset("assets/doc.css") }

// Rules returns the markdown authoring guide agents follow to produce docs
// that match the design system.
func Rules() []byte { return mustAsset("assets/rules.md") }

// Kinds returns the available template kinds in display order. The returned
// slice is a copy; callers may modify it freely.
func Kinds() []string { return append([]string(nil), templateKinds...) }

// Template returns the doc-generation prompt template for kind. Unknown
// kinds return an error; use Kinds for the valid set.
func Template(kind string) ([]byte, error) {
	for _, k := range templateKinds {
		if k == kind {
			return mustAsset("assets/templates/" + k + ".md"), nil
		}
	}
	return nil, fmt.Errorf("unknown template kind %q", kind)
}

// mustAsset reads an embedded asset. A missing asset is a build defect (the
// files are compiled in), so it panics rather than returning an error.
func mustAsset(name string) []byte {
	b, err := assets.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("agentkit: missing embedded asset %s: %v", name, err))
	}
	return b
}
