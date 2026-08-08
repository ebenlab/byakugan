---
name: update-architecture-docs
description: >
  Refresh byakugan's self-hosted docs (testdata/demo/byakugan/) for a release:
  update the architecture page to match the shipped code and add PRD pages for
  major features. Use when cutting a release, tagging a version, after merging
  a major feature, or when asked to "update the architecture docs",
  "document the release", or "refresh the self docs".
---

# Update byakugan's self-hosted architecture docs

Byakugan dogfoods itself: `testdata/demo/byakugan/` holds the project's own
architecture doc (and feature PRDs), served by the demo and exercised by e2e.
This skill keeps those pages truthful at every release.

## 1. Establish what changed

```bash
LAST=$(git describe --tags --abbrev=0 2>/dev/null || git rev-list --max-parents=0 HEAD)
git log --oneline "$LAST"..HEAD
git diff --stat "$LAST"..HEAD
```

Classify the changes: architecture-relevant (new packages, routes, flags,
subcommands, asset pipeline, UI voices) vs major features (deserve a PRD page)
vs noise (skip).

## 2. Refresh the architecture page

Edit `testdata/demo/byakugan/architecture.html` **in place** (never rename it —
e2e counts pages, and this file is the stable one). Verify every stated fact
against the code before keeping it:

- Routes → `internal/server/server.go` (`HandleFunc` calls)
- Packages and roles → `internal/*` package comments
- Flags/subcommands → `main.go` (`newServeFlags`, `printUsage`)
- Debounce, SSE, injection contract → `watcher.go`, `server.go`
- Template kinds → `internal/agentkit/agentkit.go`

Follow the authoring conventions exactly: `go run . rules` prints the guide.
Diagrams use the `.d-*` SVG grammar only (no hard-coded colors), each SVG has
its own `<marker id="arr">`, `role="img"`, and an `aria-label`. Keep the date
in the kicker current.

## 3. Add PRD pages for major features

One page per major feature shipped: `testdata/demo/byakugan/prd-<slug>.html`.
Get the skeleton prompt with `go run . template prd`. Link related pages with
relative hrefs.

**Reserved search terms** — some e2e tests depend on a term ranking a specific
page first or matching exactly one page. Before writing, grep the spec for its
fixture queries (`grep -E "fill\(|\?q=|keyboard.type" e2e/tests/byakugan.spec.ts`)
and keep those terms where the tests expect them — currently `internals`
(title-ranks internals.html first) and `dispatch` (single-hit drawer filter:
must appear only in internals.html, and in a heading or the first 2,000 chars
of visible text, because the index caps per-page body text at 2,000 chars).

## 4. Update the e2e fixture counts

Adding a page changes the fixture totals in `e2e/tests/byakugan.spec.ts`.
Update every count assertion (grep for `toHaveCount(` and `pageCount`):
landing card count, `#bk-meta` "N pages", mobile card count,
`.bk-card-updated` count, drawer `.bk-tree-count` count, and API `pageCount`.
Editing an existing page in place changes nothing.

## 5. Verify

```bash
go vet ./... && go test ./...
cd e2e && npx playwright test
```

Then eyeball it: `go run . --no-update-check --port 4666 testdata/demo`, open
the `byakugan` project, check both themes, confirm diagrams read in dark mode.

## 6. Ship it with the release

Commit the doc updates on the release branch/tag so the published demo matches
the published binary. No AI co-author footers in commit messages.
