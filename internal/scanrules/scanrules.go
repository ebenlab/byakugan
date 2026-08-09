// Package scanrules centralizes which parts of a docs tree byakugan reads:
// the directory skip rules, the nesting-depth limit, and path→project
// mapping. The indexer and the watcher both consume these, so the two can
// never drift apart — and the depth limit also caps how many directories
// the watcher pins (inotify watches are a finite kernel resource on Linux).
package scanrules

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DefaultMaxDepth is the folder-nesting limit when EnvMaxDepth is unset:
// directories nested more than this many levels below the root are ignored.
const DefaultMaxDepth = 8

// EnvMaxDepth names the environment variable that overrides DefaultMaxDepth.
const EnvMaxDepth = "BYAKUGAN_MAX_DEPTH"

// SkipName reports whether a directory name is never scanned or watched:
// hidden directories and node_modules.
func SkipName(name string) bool {
	return strings.HasPrefix(name, ".") || name == "node_modules"
}

// MaxDepth returns the nesting limit, reading EnvMaxDepth fresh on each
// call so a Rebuild always honors the current environment. Invalid or
// non-positive values fall back to DefaultMaxDepth.
func MaxDepth() int {
	if v := os.Getenv(EnvMaxDepth); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return n
		}
	}
	return DefaultMaxDepth
}

// TooDeep reports whether a directory at rel (relative to the scan root,
// native or slash separators) is nested more than limit levels below it.
func TooDeep(rel string, limit int) bool {
	if rel == "." || rel == "" {
		return false
	}
	return strings.Count(filepath.ToSlash(rel), "/")+1 > limit
}

// ProjectOf returns the first segment of a slash-relative page path — the
// project the page belongs to — or "" for root-level pages.
func ProjectOf(rel string) string {
	if i := strings.Index(rel, "/"); i >= 0 {
		return rel[:i]
	}
	return ""
}
