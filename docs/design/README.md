# Byakugan design assets

Produced with Claude Design (handoff 2026-08-08). The comms/*.txt files in
this folder are the original design↔code handoff records.

## Assets

| File | Use |
| --- | --- |
| `byakugan-mark.svg` | Canonical eye mark, `currentColor` — inherits ink from context. Used inline in `landing.html` / `inject.js`. |
| `byakugan-logo.svg` | Mark + wordmark lockup (`currentColor`, system-sans `<text>` — metrics vary per platform; prefer mark + real text in fixed-metric contexts). |
| `byakugan-mark-light.svg` / `byakugan-mark-dark.svg` | Fixed-ink variants (`#111113` / `#ededf0`) for contexts without `currentColor`, e.g. the GitHub README `<picture>` header. |
| `internal/server/assets/favicon.svg` | Fixed-ink tile (dark rect + white eye) — carries its own contrast, served at `/__byakugan/favicon.svg`. |

## Rules (settled direction)

- **Monochrome applies to Byakugan chrome only** (landing page, overlay,
  drawer): black/white/gray palette, no accent color.
- **Docs keep the blueprint editorial palette**: drafting-blue accent and
  semantic diagram colors (blue actor / green datastore / amber async).
  `doc.css` is never restyled by chrome changes.
- The eye emoji is retired everywhere the brand appears; the SVG mark
  replaces it (chrome, README, release art).
