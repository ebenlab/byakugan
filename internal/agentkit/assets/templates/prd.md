# Prompt template — PRD

Fill the `<angle-bracket>` slots, then paste everything below the line into
your agent.

---

Write a product requirements document (PRD) as a single self-contained HTML
page in the byakugan doc style.

Feature: <what is being proposed and why now>
Project folder: docs/<project>/
Output file: docs/<project>/prd-<slug>.html

Setup — do this first:
1. Run `byakugan rules` and follow every rule in it (page skeleton, .d-*
   diagram grammar, self-containment, relative links).
2. If `docs/_shared/doc.css` does not exist, create it with
   `byakugan style > docs/_shared/doc.css`. Link it from the page:
   `<link rel="stylesheet" href="../_shared/doc.css">`.

Required structure:
- `.doc-head`: kicker `<project> · product requirements ·
  <span class="badge badge--draft">draft</span> · <date>`; h1 `PRD: <name>`;
  a lede stating the change and the expected outcome in one sentence.
- Sections, in order (each a `<section>` with an `<h2>`):
  1. Problem — the user pain, with a `.stats` row of 2–4 measured numbers.
  2. Proposal — what we build; include one `.plate` diagram (sequence or
     flow) in the `.d-*` grammar showing the new user/system flow.
  3. Requirements — a table with columns Requirement / Priority / Notes.
  4. Success metrics — targets and the measurement window, as a
     `callout--decision` or a second `.stats` row.
  5. Out of scope — bullet list of explicit non-goals.
  6. Open questions — bullet list; when a question is resolved, move the
     answer into the relevant section and delete the bullet.
- `.doc-foot`: `<project> · prd-<slug> · owner: <who> · style: ../_shared/doc.css`.

Tone: product-precise. Numbers over adjectives; every claim in the Problem
section carries a measurement. The proposal says what changes for the user
before it says what changes in the system.

Miniature example of the shape (real PRDs go much deeper per section):

```html
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>PRD: Signup Flow</title>
<link rel="stylesheet" href="../_shared/doc.css">
</head>
<body class="doc">
<header class="doc-head">
  <p class="kicker">onboarding · product requirements ·
     <span class="badge badge--draft">in review</span> · 2026-08-05</p>
  <h1>PRD: Signup Flow</h1>
  <p class="lede">Replace password-and-verify signup with magic links, and move
  profile questions to after the first successful login.</p>
</header>
<main>
<section>
  <h2>Problem</h2>
  <p>Activation drops 40% at the email-verification step: users create a
  password, get bounced to their inbox, and a large share never come back.</p>
  <div class="stats">
    <div class="stat"><b>40%</b><span>drop at verification</span></div>
    <div class="stat"><b>75%</b><span>target completion</span></div>
    <div class="stat"><b>2 wks</b><span>measurement window</span></div>
  </div>
</section>
<section>
  <h2>Proposal</h2>
  <p>Magic-link signup: the form asks for an email address and nothing else.
  The link both verifies the address and starts the session.</p>
  <!-- .plate figure with a sequence diagram (.d-life lifelines, .d-flow /
       .d-async messages, marker id="arr" in defs) goes here -->
</section>
<section>
  <h2>Requirements</h2>
  <table>
    <thead><tr><th>Requirement</th><th>Priority</th><th>Notes</th></tr></thead>
    <tbody>
      <tr><td>Link expires in 15 minutes</td><td>P0</td><td>single use</td></tr>
      <tr><td>Resend with 60s cooldown</td><td>P1</td><td>rate-limited per address</td></tr>
    </tbody>
  </table>
</section>
</main>
<footer class="doc-foot">onboarding · prd-signup · owner: growth · style: ../_shared/doc.css</footer>
</body>
</html>
```
