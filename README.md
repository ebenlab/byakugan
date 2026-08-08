# 👁️ Byakugan

**A tiny live server for architecture docs and PRDs.**

Byakugan serves a folder of HTML documents — the kind coding agents generate
during the software development process — as a navigable, searchable site.
Point it at a docs folder and every subfolder becomes a project, every HTML
file becomes a page, and everything is indexed for instant search.

Built for one use case: letting architects and engineers navigate project
architecture and the decisions made over time by agents and other devs.

## Install

Download a binary for macOS, Linux, or Windows from
[Releases](https://github.com/ebenlab/byakugan/releases), or build from source:

```bash
go install github.com/ebenlab/byakugan@latest
```

## Usage

```bash
byakugan ./docs
```

```
byakugan v0.1.0 — serving /Users/you/docs
→ http://127.0.0.1:4664
```

| Flag | Default | Description |
| --- | --- | --- |
| `--port` | `4664` | Port to listen on |
| `--host` | `127.0.0.1` | Interface to bind (`0.0.0.0` to expose on the network) |
| `--open` | `false` | Open the default browser after starting |
| `--no-watch` | `false` | Disable file watching and live reload |
| `--version` | | Print version and exit |

## What you get

- **Landing page** — every first-level subfolder is listed as a project with
  its pages; visiting `/some-project/` scopes the view to that project.
- **Search** — the server indexes titles, headings, and body text of every
  HTML file; search runs instantly in the browser (`/` to focus).
- **Navigation overlay** — every served page gets a floating **👁️ Byakugan**
  button (`b` to toggle) opening a drawer with the full project tree, search,
  and prev/next links.
- **Live reload** — the folder is watched; when an agent regenerates docs,
  the index refreshes and every open browser tab reloads.
- **One static binary** — no runtime, no config, no database. The UI is
  embedded in the executable.

## Folder convention

```
docs/
├── payments/               ← project
│   ├── architecture.html
│   └── adr-001-ledger.html
├── onboarding/             ← project
│   └── prd-signup-flow.html
└── overview.html           ← top-level document
```

Anything that isn't HTML (images, CSS, PDFs) is served as-is, so pages can
reference their own assets with relative links. Hidden folders and
`node_modules` are ignored.

## Development

```bash
go test ./...
go run . ./testdata/demo
```

Releases are cut by tagging: `git tag v0.x.y && git push --tags` builds and
publishes binaries for all platforms via GoReleaser.

## License

[MIT](LICENSE) © Eben Labs
