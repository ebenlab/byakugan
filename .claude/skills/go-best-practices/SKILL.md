---
name: go-best-practices
description: Go conventions and review checklist for Byakugan. Use when writing or reviewing any Go code in this repo — before committing, opening a PR, or reviewing one.
---

# Go best practices (Byakugan)

Apply these when writing or reviewing Go in this repository. They encode both
general Go idiom and this project's specific constraints.

## Design

- Small packages with one responsibility; `main.go` is wiring only.
- Accept interfaces, return structs. Don't define an interface until a second
  implementation exists (tests count).
- Concurrency: guard shared state with the smallest possible mutex scope;
  prefer immutable snapshots swapped atomically (see `index.Snapshot`) over
  fine-grained locking.
- Context flows down: any function doing I/O on a request path takes the
  request's context.

## Errors

- Wrap with context: `fmt.Errorf("indexing %s: %w", path, err)`.
- Fail soft on per-item errors during folder scans (skip the file), fail hard
  on startup errors (bad flag, missing folder).
- Never ignore an error silently; `_ =` requires a comment saying why.

## Testing

- Table-driven tests; `t.TempDir()` for filesystem fixtures — never commit
  throwaway fixture trees for unit tests.
- HTTP behavior through `httptest` against `Server.ServeHTTP`, not a live port.
- Test names describe behavior: `TestPathTraversalBlocked`, not `TestServe2`.

## Cross-platform (hard requirement)

- `filepath.*` for disk paths, `path.*`/forward slashes for URLs. Convert at
  the boundary with `filepath.ToSlash`/`FromSlash`.
- No shelling out to OS tools except the browser-opener in `main.go`.

## Review checklist

Before approving or committing:

1. `go vet ./...` and `go test ./...` clean?
2. New dependency? Reject unless an issue justifies it.
3. Exported identifiers documented?
4. Any path string concatenation? Reject — see cross-platform rules.
5. User-visible behavior change? Playwright test in `e2e/` updated?
6. Does `byakugan <folder>` still work with zero flags?
