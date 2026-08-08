# Prompt template — platform overview

Fill the `<angle-bracket>` slots, then paste everything below the line into
your agent.

---

Write the top-level overview of the whole system as a single self-contained
HTML page in the byakugan doc style. This page lives at the docs root and is
the first thing a reader opens.

Platform: <what the overall system does>
Docs root: docs/
Output file: docs/overview.html   (root level — NOT inside a project folder)

Setup — do this first:
1. Run `byakugan rules` and follow every rule in it (page skeleton, .d-*
   diagram grammar, self-containment, relative links).
2. If `docs/_shared/doc.css` does not exist, create it with
   `byakugan style > docs/_shared/doc.css`. This page sits at the root, so
   link it as `<link rel="stylesheet" href="_shared/doc.css">` (no `../`).

Required structure:
- `.doc-head`: kicker `platform · system landscape ·
  <span class="badge badge--accepted">current</span> · <date>`; h1
  `<Platform> Overview` (or just "Platform Overview"); a lede telling the
  reader this is the map — what runs, what talks to what, and where each
  decision is written down.
- Sections, in order (each a `<section>` with an `<h2>`):
  1. System landscape — one `.plate` diagram in the `.d-*` grammar showing
     every deployed service (`.d-box`), external actors (`.d-actor`),
     shared infra (`.d-store`, `.d-queue`), and the arrows between them
     (`.d-flow` sync, `.d-async` events). The prose names each hop and
     links each service name to its project doc with a relative href
     (`<a href="payments/architecture.html">Payments</a>`).
  2. Projects — a table with columns Project / Owns / Key docs, one row per
     first-level docs folder, linking architecture/PRD/ADR pages relatively.
  3. Conventions — the cross-cutting rules of the platform (how services
     communicate, who owns data, where events flow); use a `callout` for
     the single most important rule.
  4. Reading order — a short ordered list telling a new engineer which
     three docs to read first, with relative links.
- `.doc-foot`: `platform · overview · projects: <n> · style: _shared/doc.css`.
- Keep this page short: it is a map, not a territory. Push detail into the
  project docs and link to them; if a section grows past ~3 paragraphs,
  move the detail out.

Tone: orientation for a newcomer. Every sentence either places a component
on the map or points at the doc that explains it.

Miniature example of the head and landscape section:

```html
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Platform Overview</title>
<link rel="stylesheet" href="_shared/doc.css">
</head>
<body class="doc">
<header class="doc-head">
  <p class="kicker">platform · system landscape ·
     <span class="badge badge--accepted">current</span> · 2026-08-08</p>
  <h1>Platform Overview</h1>
  <p class="lede">The top-level map of the system: what runs, what talks to
  what, and where each decision is written down.</p>
</header>
<main>
<section>
  <h2>System landscape</h2>
  <p>Client apps reach everything through a single edge gateway. Behind it sit
  <a href="onboarding/prd-signup.html">Onboarding</a> and
  <a href="payments/architecture.html">Payments</a>; cross-domain facts travel
  asynchronously over the event bus.</p>
  <figure class="plate">
    <svg class="diagram" viewBox="0 0 780 340" role="img" aria-label="System landscape">
      <defs>
        <marker id="arr" viewBox="0 0 10 10" refX="9" refY="5"
                markerWidth="7" markerHeight="7" orient="auto-start-reverse">
          <path d="M0 0L10 5L0 10z" class="d-arrhead"/>
        </marker>
      </defs>
      <rect class="d-actor" x="24" y="118" width="128" height="60" rx="8"/>
      <text class="d-title" x="88" y="144">Client apps</text>
      <rect class="d-box" x="236" y="118" width="136" height="60" rx="8"/>
      <text class="d-title" x="304" y="144">API Gateway</text>
      <rect class="d-queue" x="360" y="288" width="200" height="36" rx="8"/>
      <text class="d-title" x="460" y="311">Event bus</text>
      <line class="d-flow" x1="152" y1="148" x2="230" y2="148" marker-end="url(#arr)"/>
    </svg>
    <figcaption><b>Fig 1</b> — every service, one picture.</figcaption>
  </figure>
</section>
</main>
<footer class="doc-foot">platform · overview · projects: 2 · style: _shared/doc.css</footer>
</body>
</html>
```
