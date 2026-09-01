# PERMIT DENIED

Top-down tank-steer dozer. You are slow, heavy, and punished for filling the street you still need. Flatten the county lot, then a town, a city, and the capitol tower. One life. The run ends when the penthouse is rubble, the tower drops, or the machine stops.

Window title: **PERMIT DENIED**.

## Run

Requires Go 1.24+ and Ebitengine v2.9.

```powershell
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

**CI:** GitHub Actions (`.github/workflows/windows-package.yml`) builds a GUI exe (`-H windowsgui`), uploads `permitdenied-windows-amd64` on push/PR, and attaches it to a GitHub Release on `v*` tags.

## Keys

| Key | Action |
|-----|--------|
| `A` / `D` | Tank-steer left / right (the machine’s left, not screen-left) |
| `W` / `S` | Forward / reverse |
| `Space` | Toggle blade |
| `Enter` / `Space` | Start, close the assessment, file amendments |
| `Esc` | Title |
| `F1` | Debug: peel one plate |
| `F2` | Debug overlay |
| `F3` | Debug: skip +15 s |

Phone: left thumb is a tiller (desired heading), right hold is throttle (slide up = forward, down = reverse), right tap toggles the blade.

## The run

**County** — spawn at the south gate, facing north. Named targets: **SHERIFF** (PIT revoked), **YARD** (dumps gone, jersey brittle), **PLANT** (mix never sets). Blade down wrecks. Heat cooks if you keep pushing. Clear all three to enter town.

**Town** — three streets, civic targets (courthouse / school / depot). Layout shuffles. Fire truck and roadblocks stall you.

**City** — steel mid-rise, garage, bus barn, overpass. Steel wants a ripper, ball, or driver. Buses park in the way.

**Capitol** — campus plus a stacked tower. Wreck floors into ramps and punch the **PENTHOUSE**, or drop the **CORE** and the tower falls with it. Either is a win.

Die on any map (heat, thrown track, pin, bury, clock) and the assessment closes. Meta unlocks stay on disk; attachments do not.

## Attachments and meta

Pickups on later maps: ripper, wide blade, wrecking ball, pile driver, extra plate. Lost on death.

Between runs, a list of up to eight filed amendments (engine, armor, starting tools, alt civic layout). Each one changes how a building falls or how you get stuck. Save file: `%AppData%\permitdenied\meta.json`.

## Assessment

Stop causes: `CLEARED` / `ENGINE COOKED` / `TRACK THROWN` / `PINNED` / `BURIED` / `COUNTY CLOCK`. Copy is a denied permit, not a trailer.

## Tunables

All numbers live in `internal/game/const.go` (speeds, heat, wreck rates, clocks, map sizes). County lot coordinates stay in `internal/lot/lot.go`.

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
- Wreck spawns rubble; blade-up stall pins; meta save/load; capitol core-drop and penthouse reach

## Contributing

See [`.github/CONTRIBUTING.md`](.github/CONTRIBUTING.md).

- Issues: bug, playtest, or change forms in [`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE/)
- PRs: [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md)
- Security: [`.github/SECURITY.md`](.github/SECURITY.md)
- Help: [`.github/SUPPORT.md`](.github/SUPPORT.md)
