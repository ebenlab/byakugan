# Contributing to Byakugan

Thanks for helping architects see everything.

## Ground rules

- **Scope**: Byakugan is a tiny live server for folders of HTML docs. Features
  that turn it into a CMS, wiki, or SSG are out of scope — say no early.
- **Dependencies**: stdlib first. New modules need an issue + justification.
- **Compatibility**: `byakugan <folder>` with zero config must always work,
  on macOS, Linux, and Windows alike.

## Workflow

1. Fork/branch from `main`.
2. Make the change, with tests:
   - Go behavior → table-driven unit tests next to the code (`go test ./...`).
   - Anything a user sees in the browser → Playwright test in `e2e/`.
3. `go vet ./... && go test ./...` locally; run `cd e2e && npx playwright test`
   if you touched the server or frontend.
4. Open a PR. CI runs unit tests on three OSes, a cross-compile check, and
   the Playwright suite — all must be green.
5. One approving review merges. Squash-merge with a descriptive message.

## Code style

- Standard `gofmt` formatting (CI assumes it; your editor should too).
- Exported identifiers carry doc comments; errors are wrapped with context.
- Frontend: vanilla JS, no build step, everything namespaced `bk-`.

## Releases

Maintainers cut releases by pushing a `v*` tag; GoReleaser does the rest.
