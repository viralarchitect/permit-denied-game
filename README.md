# PERMIT DENIED

One-lot top-down dozer gauntlet. You are slow, heavy, and punished for filling the street you still need.

Window title: **PERMIT DENIED**. Same lot every run. No meta, no second player, no campaign.

## Run

Requires Go 1.24+ and Ebitengine v2.9.

```bash
go test ./...
go run ./cmd/permitdenied
```

Window starts at 1280×896 (4× a 320×224 logical screen). Resize is enabled; the sim stays 60 TPS.

## Build

**Local (Windows):** double-click or run `build.bat`. It runs tests, then writes `dist\permitdenied.exe` (console left on for debugging). Do not commit the exe.

```bat
build.bat
dist\permitdenied.exe
```

**CI:** GitHub Actions (`.github/workflows/windows-package.yml`) builds a GUI exe (`-H windowsgui`), uploads it as the `permitdenied-windows-amd64` workflow artifact on push/PR, and attaches it to a GitHub Release when you push a `v*` tag (for example `v1.0.0`).

## Keys

| Key | Action |
|-----|--------|
| `A` / `D` | Tank-steer left / right (the machine’s left, not screen-left) |
| `W` / `S` | Forward / reverse |
| `Space` | Toggle blade |
| `Enter` / `Space` | Start, or **again** on the tally |
| `Esc` | Title |
| `F1` | Debug: peel one plate |
| `F2` | Debug overlay |
| `F3` | Debug: skip +15 s |

Phone: left thumb is a tiller (desired heading), right hold is throttle (slide up = forward, down = reverse), right tap toggles the blade.

## The strip

Spawn at the south gate, facing north. Named targets, marked with rim pips:

- **SHERIFF** (amber) — smash it and cruisers lose PIT
- **YARD** (rust) — smash it and dumps despawn, jersey walls go brittle
- **PLANT** (cyan) — smash it and concrete never sets

Blade down wrecks. Blade up glances and leaves your sides open. Heat cooks the engine if you keep pushing a wall or sit wedged. Four palette plates; the last peel throws a track.

## Show-off tests

1. New player, no tooltip wall: steer, toggle blade, die inside 90 s.
2. A learned player can look at the plant pip and judge time.
3. Coward run (0 targets) and two-target run do **not** cash out the same (`×1.0` vs `×1.6`).
4. Heat readable from the sprite; plates readable from palette. No numeric bars required.
5. At least one street on the drag is stolen by player-made rubble in a typical run.

## Agent control tests

Covered by `go test ./...`:

- Spawn, W, 1 second → `dozer.Y < 1180 - 40`
- Spawn, A 0.5 s → heading west-of-north
- Blade down, push sheriff → HP drops
- `TestForwardVector` and `TestMultTable`

## Contributing

One lot. Same strip every run. See [`.github/CONTRIBUTING.md`](.github/CONTRIBUTING.md).

- Issues: bug, playtest, or change forms in [`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE/)
- PRs: [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md)
- Security: [`.github/SECURITY.md`](.github/SECURITY.md)
- Help: [`.github/SUPPORT.md`](.github/SUPPORT.md)
