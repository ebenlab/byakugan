# Byakugan doc authoring rules

How to write HTML docs that match the "blueprint editorial" design system.
Get the stylesheet with `byakugan style`; get a ready-to-paste prompt for a
specific doc kind with `byakugan template <kind>`.

## Files and folders

- First-level folders under the docs root are projects
  (`docs/payments/architecture.html`). HTML files at the root
  (`docs/overview.html`) are top-level documents.
- One shared stylesheet lives at `docs/_shared/doc.css`. Create it once with
  `byakugan style > docs/_shared/doc.css`, then link it from every page:
  `<link rel="stylesheet" href="../_shared/doc.css">` inside a project,
  `<link rel="stylesheet" href="_shared/doc.css">` at the root.
- Cross-link docs with relative hrefs (`../payments/architecture.html`),
  never absolute paths or URLs.
- Name files by kind: `architecture.html`, `adr-001-<slug>.html`,
  `prd-<slug>.html`, `overview.html`. ADR numbers never get reused.
- Every page must be self-contained: no CDNs, no external fonts, no scripts,
  no build step. Images and extra assets sit next to the page and are linked
  relatively.
- Every page needs a `<title>` (it names the page in navigation and search),
  `<meta charset="utf-8">`, and the viewport meta. The shared stylesheet
  declares `color-scheme: light dark`; a page that does not link it must add
  `<meta name="color-scheme" content="light dark">` itself.

## Reuse the design system — never restyle

- The shared stylesheet is the whole design system: skeleton classes,
  callouts, tables, plates, diagram grammar, and **light/dark theming**.
  Reuse it. Do not write per-page `<style>` blocks for anything it already
  covers, and do not copy rules out of it into a page.
- Dark mode is three-state and comes free with `doc.css`: the OS scheme by
  default, overridden by an explicit viewer choice the Byakugan theme toggle
  stores as `data-bk-theme` on `<html>`. Pages never manage this themselves —
  never set `data-bk-theme`, never add your own `prefers-color-scheme` blocks.
- If a page truly needs a custom element, style it with the shared CSS
  variables (`--paper`, `--card`, `--ink`, `--muted`, `--line`, `--line-soft`,
  `--accent`, `--accent-soft`, `--ok`, `--warn`, `--risk`, and their `-soft`
  tints) so it adapts to both themes automatically. Hard-coded hex colors are
  a bug: they will be wrong in one of the two modes.
- Never redefine the `:root` variables in a page — that desynchronizes it
  from every other page in the dossier.

## Page skeleton

```html
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Title of the document</title>
<link rel="stylesheet" href="../_shared/doc.css">
</head>
<body class="doc">

<header class="doc-head">
  <p class="kicker">payments · architecture ·
     <span class="badge badge--accepted">accepted</span> · 2026-08-08</p>
  <h1>Title of the document</h1>
  <p class="lede">One-paragraph summary of what this doc claims.</p>
</header>

<main>
  <section>
    <h2>First section</h2>
    <p>…</p>
  </section>

  <figure class="plate">
    <svg class="diagram" viewBox="0 0 780 340" role="img" aria-label="…">…</svg>
    <figcaption><b>Fig 1</b> — caption text.</figcaption>
  </figure>
</main>

<footer class="doc-foot">payments · ADR-001 · supersedes: none · style: ../_shared/doc.css</footer>
</body>
</html>
```

- `body class="doc"` is required — every style hangs off it.
- `.doc-head` holds the `.kicker` (project · doc type · status badge · date),
  the `h1`, and a one- or two-sentence `.lede`.
- Status badges: `badge--accepted`, `badge--draft`, `badge--superseded`.
- Wrap each section in `<section>` with an `<h2>` — headings number
  themselves (01, 02, …); never hand-number. `<h3>` for subsections.
- Callouts: `<div class="callout"><strong>Label</strong><p>…</p></div>`.
  The first `<strong>` becomes the small-caps label. Variants:
  `callout--decision` (green) for the decision itself, `callout--risk`
  (red) for risks and revisit triggers.
- Stat rows (PRD metrics): `<div class="stats">` containing
  `<div class="stat"><b>40%</b><span>drop at verification</span></div>`.
- Tables: plain `<table>` with `<thead>`/`<tbody>` — styling is automatic.
- `.doc-foot`: one mono provenance line, e.g.
  `payments · ADR-001 · supersedes: none · style: ../_shared/doc.css`.

## Diagrams — the .d-* SVG grammar

Diagrams are hand-written inline SVG inside a `.plate` figure with a
`figcaption`. Skeleton:

```html
<figure class="plate">
  <svg class="diagram" viewBox="0 0 780 340" role="img" aria-label="Component diagram">
    <defs>
      <marker id="arr" viewBox="0 0 10 10" refX="9" refY="5"
              markerWidth="7" markerHeight="7" orient="auto-start-reverse">
        <path d="M0 0L10 5L0 10z" class="d-arrhead"/>
      </marker>
    </defs>
    <rect class="d-box" x="236" y="96" width="140" height="60" rx="8"/>
    <text class="d-title" x="306" y="122">Payments API</text>
    <text class="d-sub" x="306" y="138">validation · idempotency</text>
    <line class="d-flow" x1="148" y1="126" x2="230" y2="126" marker-end="url(#arr)"/>
  </svg>
  <figcaption><b>Fig 1</b> — caption text.</figcaption>
</figure>
```

Vocabulary:

| Class | Use |
| --- | --- |
| `.d-box` | component box (`<rect rx="8">`) |
| `.d-actor` | external actor (blue tint) |
| `.d-store` | datastore (green tint, use `<rect rx="10">`) |
| `.d-queue` | async infra — queues, buses (dashed outline) |
| `.d-title` | box title text, centered on the box |
| `.d-sub` | small mono annotation under the title |
| `.d-flow` | solid arrow line (sync call) |
| `.d-async` | dashed arrow line (async/event) |
| `.d-label` | small free-floating annotation text |
| `.d-life` | sequence-diagram lifeline (dotted vertical) |
| `.d-lane` | swimlane / boundary background |
| `.d-num` | numbered step circle (`<circle r="9">`) |
| `.d-numtext` | the number inside the circle |
| `.d-arrhead` | arrowhead path inside the `<marker>` |

Rules:

- Size with `viewBox` only — never `width`/`height` attributes on the SVG;
  CSS makes it fluid.
- Every arrow (`.d-flow`, `.d-async`) ends with `marker-end="url(#arr)"`,
  and each SVG defines its own `<marker id="arr">` in `<defs>` — ids are
  per-document, keep them local.
- Boxes pair a `<rect>` with centered `.d-title` and optional `.d-sub` text
  (`text-anchor` is centered by the classes; put x at the box center).
- Numbered steps: `<circle class="d-num" r="9">` plus `.d-numtext`.
- Color only through the `.d-*` classes so dark mode works — never
  hard-code fills or strokes.
- Give each diagram `role="img"` and an `aria-label`.

## ASCII diagrams and code snippets — the zero-cost forms

Not every figure needs SVG. Two lighter forms are part of the system, and
both render as fast as plain text because they are plain text — no
libraries, no highlighter, nothing to parse or execute:

- **ASCII diagrams** — `<pre class="ascii">` inside a `.plate`. Use them for quick topologies, request
  paths, and anything reviewed in diffs; keep the `.d-*` SVG grammar for
  figures that earn the polish. Rules: lines ≤ 78 columns (the container
  scrolls sideways rather than wraps), Unicode box-drawing characters
  (`─ │ ┌ ┐ └ ┘ ├ ▶`) over `+--+`, solid arrows for synchronous calls and
  `─ ─ ▶` dashed for async — the same conventions as the SVG grammar.
- **Code snippets** — a bare `<pre><code>` block. Quote real code, short and load-bearing: the invariant, the
  contract, the one function that explains the section — never page-long
  listings. A leading comment line naming the file (`// server.go — …`)
  replaces a caption. There is deliberately no syntax highlighter; do not
  add one.

## Tone

Write like an engineer explaining a decision, not a brochure: short
paragraphs, concrete numbers, named trade-offs. The lede states what the doc
claims in one or two sentences. Prefer a table to a paragraph when comparing
options; prefer a diagram to a table when describing flow.
