package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunDispatch(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantHandled bool
		wantCode    int
		wantStdout  []string // substrings that must appear on stdout
		wantStderr  []string // substrings that must appear on stderr
	}{
		{
			name:        "no args falls through to serve",
			args:        nil,
			wantHandled: false,
		},
		{
			name:        "folder arg falls through to serve",
			args:        []string{"./docs"},
			wantHandled: false,
		},
		{
			name:        "flags fall through to serve",
			args:        []string{"--port", "8080", "./docs"},
			wantHandled: false,
		},
		{
			name:        "help prints usage with subcommands and flags",
			args:        []string{"help"},
			wantHandled: true,
			wantCode:    0,
			wantStdout: []string{
				"Usage:", "Subcommands:", "Flags:", "Examples:",
				"style", "rules", "template <kind>",
				"prd, adr, architecture, overview",
				"-port", "-no-watch",
			},
		},
		{
			name:        "version prints version",
			args:        []string{"version"},
			wantHandled: true,
			wantCode:    0,
			wantStdout:  []string{"byakugan dev"},
		},
		{
			name:        "style prints the design system css",
			args:        []string{"style"},
			wantHandled: true,
			wantCode:    0,
			wantStdout:  []string{"blueprint editorial", "body.doc", ".d-flow"},
		},
		{
			name:        "rules prints the authoring guide",
			args:        []string{"rules"},
			wantHandled: true,
			wantCode:    0,
			wantStdout:  []string{"# Byakugan doc authoring rules", "byakugan style", "viewBox"},
		},
		{
			name:        "template prd",
			args:        []string{"template", "prd"},
			wantHandled: true,
			wantCode:    0,
			wantStdout:  []string{"PRD", "byakugan rules", "byakugan style > docs/_shared/doc.css"},
		},
		{
			name:        "template adr",
			args:        []string{"template", "adr"},
			wantHandled: true,
			wantCode:    0,
			wantStdout:  []string{"ADR", "byakugan rules"},
		},
		{
			name:        "template architecture",
			args:        []string{"template", "architecture"},
			wantHandled: true,
			wantCode:    0,
			wantStdout:  []string{"architecture", "byakugan rules"},
		},
		{
			name:        "template overview",
			args:        []string{"template", "overview"},
			wantHandled: true,
			wantCode:    0,
			wantStdout:  []string{"overview", "byakugan rules"},
		},
		{
			name:        "template with no kind lists kinds on stderr",
			args:        []string{"template"},
			wantHandled: true,
			wantCode:    2,
			wantStderr:  []string{"available kinds: prd, adr, architecture, overview"},
		},
		{
			name:        "template with unknown kind lists kinds on stderr",
			args:        []string{"template", "rfc"},
			wantHandled: true,
			wantCode:    2,
			wantStderr:  []string{`unknown template kind "rfc"`, "available kinds: prd, adr, architecture, overview"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			handled, code := run(tt.args, &stdout, &stderr)
			if handled != tt.wantHandled {
				t.Fatalf("run(%v) handled = %v, want %v", tt.args, handled, tt.wantHandled)
			}
			if !handled {
				if stdout.Len() != 0 || stderr.Len() != 0 {
					t.Errorf("run(%v) wrote output despite not handling: stdout=%q stderr=%q", tt.args, stdout.String(), stderr.String())
				}
				return
			}
			if code != tt.wantCode {
				t.Errorf("run(%v) code = %d, want %d", tt.args, code, tt.wantCode)
			}
			for _, want := range tt.wantStdout {
				if !strings.Contains(stdout.String(), want) {
					t.Errorf("run(%v) stdout missing %q\nstdout:\n%s", tt.args, want, stdout.String())
				}
			}
			for _, want := range tt.wantStderr {
				if !strings.Contains(stderr.String(), want) {
					t.Errorf("run(%v) stderr missing %q\nstderr:\n%s", tt.args, want, stderr.String())
				}
			}
			if len(tt.wantStderr) == 0 && stderr.Len() != 0 {
				t.Errorf("run(%v) wrote unexpected stderr: %q", tt.args, stderr.String())
			}
		})
	}
}
