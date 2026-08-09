package scanrules

import "testing"

func TestSkipName(t *testing.T) {
	for name, want := range map[string]bool{
		".git": true, ".hidden": true, "node_modules": true,
		"docs": false, "payments": false, "node_modules2": false,
	} {
		if got := SkipName(name); got != want {
			t.Errorf("SkipName(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestTooDeep(t *testing.T) {
	cases := []struct {
		rel   string
		limit int
		want  bool
	}{
		{".", 1, false},
		{"", 1, false},
		{"a", 1, false},
		{"a/b", 1, true},
		{"a/b", 2, false},
		{"a/b/c", 2, true},
	}
	for _, c := range cases {
		if got := TooDeep(c.rel, c.limit); got != c.want {
			t.Errorf("TooDeep(%q, %d) = %v, want %v", c.rel, c.limit, got, c.want)
		}
	}
}

func TestMaxDepth(t *testing.T) {
	t.Setenv(EnvMaxDepth, "")
	if got := MaxDepth(); got != DefaultMaxDepth {
		t.Errorf("default = %d, want %d", got, DefaultMaxDepth)
	}
	t.Setenv(EnvMaxDepth, "3")
	if got := MaxDepth(); got != 3 {
		t.Errorf("env=3 → %d", got)
	}
	for _, bad := range []string{"0", "-2", "abc", "1.5"} {
		t.Setenv(EnvMaxDepth, bad)
		if got := MaxDepth(); got != DefaultMaxDepth {
			t.Errorf("env=%q → %d, want default %d", bad, got, DefaultMaxDepth)
		}
	}
}

func TestProjectOf(t *testing.T) {
	for rel, want := range map[string]string{
		"payments/adr.html": "payments",
		"a/b/c.md":          "a",
		"overview.html":     "",
	} {
		if got := ProjectOf(rel); got != want {
			t.Errorf("ProjectOf(%q) = %q, want %q", rel, got, want)
		}
	}
}
