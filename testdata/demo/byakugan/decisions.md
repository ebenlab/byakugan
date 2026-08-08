---
title: Decision log
status: living
date: 2026-08-08
---

# Decision log

Running record of the calls that shaped Byakugan, one line each. Full
rationale lives in the linked docs; this page exists so nobody re-litigates
by accident. It is also Byakugan's own dogfood for Markdown rendering.

## Decisions

| # | Decision | Why (short) | Where |
| - | -------- | ----------- | ----- |
| 1 | Go over Rust | Learning project that still had to ship fast | [Architecture Overview](architecture.html) |
| 2 | Stdlib first; deps need a written case | Keeps the binary one file and the surface auditable | [Internals](internals.html) |
| 3 | Server-built index, client-side search | Instant results, no JS toolchain, works offline | [Internals](internals.html) |
| 4 | Injected overlay is a guest in user documents | `bk-` prefixed, idempotent, never touches host styles | [Internals](internals.html) |
| 5 | Monochrome chrome, colored docs | Chrome recedes; semantic diagram color carries meaning | design comms 003 |
| 6 | Flat design everywhere | Borders over shadows — stated owner preference | design comms 005 |
| 7 | Markdown via goldmark, scoped hard | ~200 LOC + one zero-dep library; not an SSG | this page |

## Topology at a glance

Markdown pages get zero-cost diagrams too — a fenced block renders as fast
as the text it is, no SVG required:

```
agent ──▶ docs/ ─ ─▶ watcher ──▶ index ──▶ server ──▶ browser
            ▲                                 │
            └──────────── read on request ────┘
                     (SSE pushes: reload)
```

And code snippets render plain and instant, no highlighter attached:

```go
// the entire debounce: a burst of saves becomes one callback
if timer == nil {
    timer = time.AfterFunc(debounce, onChange)
} else {
    timer.Reset(debounce)
}
```

## Standing boundaries

- ~~Syntax highlighting~~ — deferred indefinitely; chroma alone outweighs the whole binary's ambition.
- No configuration files, templates, or themes. `byakugan <folder>` must stay the entire manual.
- Every user-visible behavior lands with a Playwright test in the same commit.

Questions about any row? Open an issue at https://github.com/ebenlab/byakugan/issues — autolinked courtesy of GFM.
