# Contributing

PERMIT DENIED is a **one-lot** top-down dozer gauntlet. Same strip every run. No campaign, no between-run meta, no second player.

Read this before you open an issue or PR:

1. [`README.md`](../README.md) — run, keys, show-off tests
2. [`PERMIT_DENIED.md`](../PERMIT_DENIED.md) — design (*what*)
3. [`PERMIT_DENIED_GO_IMPLEMENTATION.md`](../PERMIT_DENIED_GO_IMPLEMENTATION.md) — numbers and APIs (*how*)
4. [`assets/usable/ASSETS.md`](../assets/usable/ASSETS.md) — tiles draw, objects collide

If those two design files conflict: fantasy follows the design doc; numbers and APIs follow the implementation guide.

## Stay on the strip

Tune [`internal/game/const.go`](../internal/game/const.go) before adding a system. A threat that does not change **steering** does not ship.

Do not add: unlocks, a second lot, body boons, locational armor, soft-body / particle debris as collision, chopper rockets, a second dozer, accounts / leaderboards, or the title `KILLDOZER`.

Copy to keep: `PERMIT DENIED`, `THE COUNTY SAID NO.`, `BLADE UP` / `BLADE DOWN`, death lines `ENGINE COOKED` / `TRACK THROWN` / `COUNTY CLOCK`.

## Develop

Needs Go (see `go.mod`) and Ebitengine v2.9.

```bat
go test ./...
go run ./cmd/permitdenied
```

Windows package (tests first, then exe; do not commit it):

```bat
build.bat
dist\permitdenied.exe
```

CI (`.github/workflows/windows-package.yml`) builds a GUI exe on push/PR to `main`/`master` and attaches it to a GitHub Release on a `v*` tag.

## Pull requests

Use the PR template. Keep the diff on one problem. Link an issue when there is one.

Same-repo PRs auto-merge with a merge commit once the **windows** CI job is green (`go test ./...` and the GUI build). Drafts stay open until they are marked ready. The head branch is deleted on merge.

Before you ask for review:

- `go test ./...` is green
- No `dist/` or `*.exe` in the commit
- New tunables are in `const.go`
- Lot geometry stays in `internal/lot/lot.go` literals (and the Tiled mirror in `assets/usable/` if you touched objects)

## Issues

Use an issue form:

| Template | Use when |
|---|---|
| **Bug** | Crash, wrong steer, scoring, beat, render, CI |
| **Playtest** | A show-off test failed, or heat/plates/rubble are unreadable |
| **Change** | A scoped proposal that stays on this strip |

Blank issues are allowed if none of those fit. Do not file requests for a campaign or a second lot.
