# Byakugan Roadmap

Structured as milestones with checklist items so it can be imported into a
project tracker (Linear) as-is. Status reflects 2026-08-08.

## Milestone 1 — v0.1.0: Core server (done, pending release)

- [x] CLI: `byakugan [flags] <folder>` with port/host/open/no-watch/version
- [x] Index: scan projects (first-level folders) + pages, extract title/headings/text
- [x] Landing page: project cards, per-project scoped view at `/project/`
- [x] Search: server-built JSON index, instant client-side scoring with highlights
- [x] Navigation overlay injected into every page (tree, search, prev/next, `b` shortcut)
- [x] Live reload: recursive watcher, debounced re-index, SSE push
- [x] CI: vet + test + build on macOS/Linux/Windows, cross-compile check
- [x] Playwright e2e suite (10 tests) running on every PR
- [x] Agent/contributor guidelines: CLAUDE.md, CONTRIBUTING.md, repo skills
- [x] Agent-facing subcommands: `help`, `version`, `style`, `rules`, `template <kind>`
- [x] `byakugan style`: shared doc design system (doc.css) printable to stdout
- [x] `byakugan rules`: authoring guide (skeleton, `.d-*` diagram grammar, conventions)
- [x] `byakugan template prd|adr|architecture|overview`: doc generation prompt templates
- [x] README: CLI reference + "For agents" workflow
- [x] Release pipeline: GoReleaser on `v*` tags
- [x] Tag and publish v0.1.0 (2026-08-08)

## Shipped between milestones

- [x] v0.2.0 — responsive mobile UI, navigation redesign, theme toggle,
  agent-facing CLI (`style`/`rules`/`template`) (2026-08-08)
- [x] Monochrome chrome + SVG mark + favicon route (design patch 003)
- [x] Flat design pass (no shadows/elevation, chrome + docs)
- [x] Back-button support (search state in `?q=`, history-integrated drawer)
- [x] Upgrade support (startup update notice + `byakugan upgrade` self-update)
- [x] Self-docs demo (byakugan documents itself in testdata/demo)

## Milestone 2 — v0.2.0: Navigation depth

- [ ] Nested project support (subfolders below first level as sections)
- [ ] Breadcrumbs on generated pages
- [ ] Sort options on landing page (recently updated first)
- [ ] Markdown rendering (serve `.md` as HTML) — agents don't always emit HTML
- [ ] `--base-path` flag for serving behind a reverse-proxy prefix

## Milestone 3 — v0.3.0: Retrieval quality

- [ ] Smarter ranking (prefix/word-boundary boosts, per-project weighting)
- [ ] Search over headings anchors: jump straight to a section, not just a page
- [ ] Keyboard-first result navigation (arrow keys, enter to open)
- [ ] Optional `byakugan.json` per project (display name, description, order)

## Milestone 4 — Distribution

- [ ] Homebrew tap (`brew install ebenlab/tap/byakugan`)
- [ ] Install script (`curl | sh`)
- [ ] Docker image for CI/preview deployments

## Maintenance (recurring)

- [ ] Dependency bumps (fsnotify, x/net) — monthly
- [ ] Go toolchain bump on new minor releases
- [ ] Triage issues weekly
