# Prompt template — ADR

Fill the `<angle-bracket>` slots, then paste everything below the line into
your agent.

---

Write an architecture decision record (ADR) as a single self-contained HTML
page in the byakugan doc style.

Decision: <the choice being recorded, e.g. "Postgres as the ledger store">
Project folder: docs/<project>/
Output file: docs/<project>/adr-<NNN>-<slug>.html
  (NNN = next unused number in the folder, zero-padded to three digits;
   never reuse a number, even for superseded ADRs)

Setup — do this first:
1. Run `byakugan rules` and follow every rule in it (page skeleton, .d-*
   diagram grammar, self-containment, relative links).
2. If `docs/_shared/doc.css` does not exist, create it with
   `byakugan style > docs/_shared/doc.css`. Link it from the page:
   `<link rel="stylesheet" href="../_shared/doc.css">`.

Required structure:
- `.doc-head`: kicker `<project> · decision record · <status badge> · <date>`
  where status is `badge--accepted`, `badge--draft`, or `badge--superseded`;
  h1 `ADR-<NNN>: <decision>`; a lede that states the decision and its core
  rationale in at most two sentences.
- Sections, in order (each a `<section>` with an `<h2>`):
  1. Context — the forces at play: constraints, load numbers, team reality.
     No solutions here.
  2. Options considered — a table with columns Option / For / Against.
     Include the rejected options honestly; 2–4 rows.
  3. Decision — a `callout--decision` stating the choice in the present
     tense ("Use X for Y"), plus, when flow matters, one `.plate` diagram in
     the `.d-*` grammar showing how the chosen design behaves.
  4. Consequences — bullet list of what becomes easier, harder, or
     mandatory; end with a `callout--risk` labeled "Revisit when" naming the
     concrete trigger that would reopen this decision.
- `.doc-foot`: `<project> · ADR-<NNN> · supersedes: <ADR link or "none"> ·
  style: ../_shared/doc.css`. If this ADR supersedes another, also edit the
  old ADR's badge to `badge--superseded` and link forward from its foot.

Tone: an engineer recording a decision for a future reader. Past forces,
present-tense decision, honest consequences. Never sell the choice; make the
trade-off legible enough that a reader could disagree intelligently.

Miniature example of the shape (real ADRs go deeper per section):

```html
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>ADR-001: Postgres as the ledger store</title>
<link rel="stylesheet" href="../_shared/doc.css">
</head>
<body class="doc">
<header class="doc-head">
  <p class="kicker">payments · decision record ·
     <span class="badge badge--accepted">accepted</span> · 2026-08-01</p>
  <h1>ADR-001: Postgres as the ledger store</h1>
  <p class="lede">The journal needs a store that can refuse to lie. We choose
  Postgres and lean on its strongest isolation level.</p>
</header>
<main>
<section>
  <h2>Context</h2>
  <p>The ledger needs serializable transactions and strong constraints, and
  it has to be operable by a two-person team.</p>
</section>
<section>
  <h2>Options considered</h2>
  <table>
    <thead><tr><th>Option</th><th>For</th><th>Against</th></tr></thead>
    <tbody>
      <tr><td><strong>Postgres</strong></td><td>True SERIALIZABLE, boring ops</td>
          <td>Single-writer ceiling ~10k TPS</td></tr>
      <tr><td>DynamoDB</td><td>Managed scale</td>
          <td>25-item transaction cap</td></tr>
    </tbody>
  </table>
</section>
<section>
  <h2>Decision</h2>
  <div class="callout callout--decision">
    <strong>Decision</strong>
    <p>Use Postgres with SERIALIZABLE isolation for journal writes; the
    ledger core retries serialization failures.</p>
  </div>
</section>
<section>
  <h2>Consequences</h2>
  <ul><li>Retry loops are mandatory — serialization failures are normal.</li></ul>
  <div class="callout callout--risk">
    <strong>Revisit when</strong>
    <p>Sustained journal writes approach 10k TPS on one partition.</p>
  </div>
</section>
</main>
<footer class="doc-foot">payments · ADR-001 · supersedes: none · style: ../_shared/doc.css</footer>
</body>
</html>
```
