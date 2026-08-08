---
name: ship-release
description: Cut a Byakugan release — pre-flight checks, tagging, and verifying the GoReleaser pipeline. Use when asked to release, ship, or publish a new version.
---

# Shipping a Byakugan release

## Pre-flight

1. `git status` clean, on `main`, up to date with origin.
2. `go vet ./... && go test ./...` green locally.
3. `cd e2e && npx playwright test` green.
4. CI on `main` is green (`gh run list --limit 3`).
5. ROADMAP.md updated: check off shipped items under the milestone.

## Version choice

Semver, pre-1.0 (Go community convention): new features or breaking
changes → bump minor; pure fixes → bump patch. Check the last tag with
`git describe --tags --abbrev=0`.

## Cut it

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

That's the whole release: the `Release` workflow runs GoReleaser, which
builds darwin/linux/windows × amd64/arm64, stamps `main.version`, and
publishes archives + checksums to GitHub Releases.

## Verify

1. `gh run watch` the Release workflow to completion.
2. `gh release view vX.Y.Z` — six archives + checksums.txt present.
3. Download one binary, run `byakugan --version`, confirm the stamp.

## If it fails

Delete the tag (`git tag -d vX.Y.Z && git push origin :refs/tags/vX.Y.Z`),
fix on `main`, re-tag. Never edit a published release's artifacts by hand.
