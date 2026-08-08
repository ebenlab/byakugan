# Prompt template — architecture doc

Fill the `<angle-bracket>` slots, then paste everything below the line into
your agent.

---

Write an architecture document for one system/service as a single
self-contained HTML page in the byakugan doc style.

System: <the service or subsystem this doc describes>
Project folder: docs/<project>/
Output file: docs/<project>/architecture.html

Setup — do this first:
1. Run `byakugan rules` and follow every rule in it (page skeleton, .d-*
   diagram grammar, self-containment, relative links).
2. If `docs/_shared/doc.css` does not exist, create it with
   `byakugan style > docs/_shared/doc.css`. Link it from the page:
   `<link rel="stylesheet" href="../_shared/doc.css">`.

Required structure:
- `.doc-head`: kicker `<project> · architecture ·
  <span class="badge badge--accepted">current</span> · <date>`; h1
  `<System> Architecture`; a lede naming the core design idea and the
  invariants it protects.
- Sections, in order (each a `<section>` with an `<h2>`):
  1. Context — what this system owns and which properties are
     non-negotiable (its invariants). Two paragraphs maximum.
  2. Component overview — one `.plate` component diagram in the `.d-*`
     grammar (`.d-box` components, `.d-actor` externals, `.d-store` data,
     `.d-queue` async infra, `.d-lane` for the service boundary), plus a
     paragraph naming each component's single responsibility.
  3. Key flows — the 1–2 flows where correctness is won or lost, each as a
     `.plate` sequence diagram (`.d-life` lifelines, `.d-num` step circles)
     with a short narration; write the failure path, not just the happy one.
  4. Data & storage — what is stored where and why; a table with columns
     Store / Holds / Why here.
  5. Failure modes — a table with columns Failure / Blast radius /
     Mitigation, plus a `callout--risk` for the one that keeps you up at
     night.
  6. Decision log — bullet list of relative links to the ADRs behind this
     design (`<a href="adr-001-<slug>.html">ADR-001 — title</a>`), newest
     first.
- `.doc-foot`: `<project> · architecture · decisions: see §6 ·
  style: ../_shared/doc.css`.

Tone: explain the machine, then defend it. Every component earns its place
by protecting an invariant named in Context; if a component protects
nothing, question it in the doc. Concrete numbers (TPS, latencies, retention)
beat qualifiers.

Miniature example of the head and one section (a real doc covers all six):

```html
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Payments Architecture</title>
<link rel="stylesheet" href="../_shared/doc.css">
</head>
<body class="doc">
<header class="doc-head">
  <p class="kicker">payments · architecture ·
     <span class="badge badge--accepted">current</span> · 2026-08-08</p>
  <h1>Payments Architecture</h1>
  <p class="lede">A double-entry ledger at the core, an outbox relay at the
  edge: how money moves through the system without ever going missing.</p>
</header>
<main>
<section>
  <h2>Context</h2>
  <p>Payments owns every money movement. Two properties are non-negotiable:
  the ledger must always balance, and downstream systems must hear about
  every movement exactly once. Everything here protects those invariants.</p>
</section>
<section>
  <h2>Component overview</h2>
  <figure class="plate">
    <svg class="diagram" viewBox="0 0 780 380" role="img" aria-label="Component diagram">
      <defs>
        <marker id="arr" viewBox="0 0 10 10" refX="9" refY="5"
                markerWidth="7" markerHeight="7" orient="auto-start-reverse">
          <path d="M0 0L10 5L0 10z" class="d-arrhead"/>
        </marker>
      </defs>
      <rect class="d-lane" x="196" y="16" width="404" height="348" rx="10"/>
      <rect class="d-box" x="236" y="96" width="140" height="60" rx="8"/>
      <text class="d-title" x="306" y="122">Payments API</text>
      <text class="d-sub" x="306" y="138">validation · idempotency</text>
      <rect class="d-store" x="356" y="216" width="150" height="64" rx="10"/>
      <text class="d-title" x="431" y="242">Postgres</text>
      <line class="d-flow" x1="148" y1="126" x2="230" y2="126" marker-end="url(#arr)"/>
    </svg>
    <figcaption><b>Fig 1</b> — components inside the payments boundary.</figcaption>
  </figure>
</section>
</main>
<footer class="doc-foot">payments · architecture · decisions: see §6 · style: ../_shared/doc.css</footer>
</body>
</html>
```
