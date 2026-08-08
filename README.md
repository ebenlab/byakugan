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

## CLI reference

```
byakugan [flags] [folder]   serve a docs folder (default ".")
byakugan <subcommand>
```

### Flags (serve mode)

| Flag | Default | Description |
| --- | --- | --- |
| `--port` | `4664` | Port to listen on |
| `--host` | `127.0.0.1` | Interface to bind (`0.0.0.0` to expose on the network) |
| `--open` | `false` | Open the default browser after starting |
| `--no-watch` | `false` | Disable file watching and live reload |
| `--version` | | Print version and exit |

### Subcommands

| Subcommand | Description |
| --- | --- |
| `help` | Print full usage, subcommands, and flags |
| `version` | Print version and exit (same as `--version`) |
| `style` | Print the shared doc stylesheet (`doc.css`) to stdout |
| `rules` | Print the doc authoring guide for agents |
| `template <kind>` | Print a doc generation prompt template — kinds: `prd`, `adr`, `architecture`, `overview` |

`template` with a missing or unknown kind lists the available kinds on
stderr and exits with status 2.

Examples:

```bash
byakugan ./docs                            # serve
byakugan --port 8080 --open ~/architecture # serve with flags
byakugan style > docs/_shared/doc.css      # install the design system
byakugan rules                             # read the authoring guide
byakugan template adr                      # prompt template for an ADR
```

## For agents

Byakugan ships everything a coding agent needs to generate docs that look
like they belong together — a design system ("blueprint editorial"), an
authoring guide, and per-kind prompt templates. The workflow:

1. **Install the stylesheet once per docs folder:**

   ```bash
   byakugan style > docs/_shared/doc.css
   ```

   Every generated page links it relatively (`../_shared/doc.css` from
   inside a project, `_shared/doc.css` at the root).

2. **Feed the rules to the agent.** `byakugan rules` prints a concise
   markdown guide: the page skeleton (`body.doc`, doc-head, numbered
   sections, callouts, stats, tables, doc-foot), the `.d-*` inline-SVG
   diagram grammar, file/folder conventions, and the self-containment rule
   (no CDNs, no external anything).

3. **Start from a template.** `byakugan template <kind>` prints a
   ready-to-paste prompt for a `prd`, `adr`, `architecture`, or `overview`
   doc — structure, tone, required sections, and a miniature example:

   ```bash
   byakugan template prd | pbcopy   # paste into your agent
   ```

Pages produced this way are plain, self-contained HTML: they render
correctly from disk, in light and dark mode, with or without byakugan
running.

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
go vet ./...            # must be clean
go test ./...           # unit tests (CLI dispatch, agentkit assets, index, server)
go run . ./testdata/demo
cd e2e && npx playwright test   # browser e2e suite (builds the binary itself)
```

The canonical copy of the design system lives at
`internal/agentkit/assets/doc.css` (what `byakugan style` prints);
`testdata/demo/_shared/doc.css` is a byte-identical copy, and a unit test
fails if the two ever drift — update both together.

Releases are cut by tagging: `git tag v0.x.y && git push --tags` builds and
publishes binaries for all platforms via GoReleaser.

## License

[MIT](LICENSE) © Eben Labs
