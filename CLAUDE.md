# Byakugan — agent guidelines

Byakugan is a single-binary Go live server for agent-generated architecture
docs and PRDs. Read this before changing anything; it is the contract every
agent and contributor works under.

## Commands

```bash
go test ./...                  # unit tests — must pass before any commit
go vet ./...                   # must be clean
go build -o byakugan .         # build the binary
go run . ./testdata/demo       # run against the demo docs (port 4664)
cd e2e && npx playwright test  # browser e2e tests (builds the binary itself)
```

## Architecture map

| Path | Responsibility |
| --- | --- |
| `main.go` | CLI flags, wiring; nothing else lives here |
| `internal/index` | Folder scanning, HTML text extraction, search index snapshot |
| `internal/server` | HTTP routes, SSE hub, HTML injection, embedded UI assets |
| `internal/server/assets` | The entire frontend (landing, overlay, search core) |
| `internal/watcher` | Recursive fsnotify watcher with debounce |
| `e2e` | Playwright tests exercising the real binary in a real browser |

Data flow: watcher → `Index.Rebuild()` → `Server.Broadcast("reload")` → every
open tab refetches `/api/index.json`.

## Rules

1. **Stdlib first.** The dependency set is `fsnotify` + `golang.org/x/net`.
   Adding any dependency requires an issue explaining why stdlib can't do it.
2. **One binary, always.** Everything the browser needs is embedded via
   `//go:embed`. Never reference a CDN, never add a JS build step. The
   frontend stays framework-free vanilla JS.
3. **Never break `byakugan <folder>`.** Zero-config must keep working; new
   behavior arrives as optional flags with sensible defaults.
4. **Injected code is a guest.** `inject.js` runs inside users' documents:
   it must stay idempotent, self-contained, prefixed (`bk-`/`__byakugan`),
   and must never alter the host page's own DOM or styles.
5. **Tests accompany behavior.** New server behavior gets a `httptest` unit
   test; new UI behavior gets a Playwright test. Both suites run on every PR
   and both must be green to merge.
6. **Windows is a first-class target.** Use `filepath` for disk paths,
   `path`/forward slashes for URLs; never concatenate path strings.
7. **Comments state constraints, not narration.** Exported identifiers get
   doc comments (Go convention); inline comments only for the non-obvious.

## Release

Releases are tag-driven: update ROADMAP.md checkboxes, then
`git tag vX.Y.Z && git push --tags`. GoReleaser builds Mac/Linux/Windows
binaries and publishes the GitHub release. Never hand-upload artifacts.
