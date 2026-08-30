---
name: verify-permitdenied
description: Drive PERMIT DENIED (Ebiten desktop gauntlet) through control-permitdenied and capture PNG plus JSON proof of title, steer, blade, smash, and tally. Use when proving a gameplay or HUD change, or when go test is not enough because the player-visible path must move.
---

# Verify PERMIT DENIED

Player surface is a window titled `PERMIT DENIED` (logical 320×224, 60 TPS). There is no HTTP server and no save file. Verification drives `internal/game.Game` through the same `Update`/`Draw` path the exe uses, with scripted keys. `drive` opens a short-lived unfocused window titled `PERMIT DENIED verify` so Ebiten can snapshot `Draw`. It does not launch `dist\permitdenied.exe` and does not send OS keystrokes.

Feature recipes live in [features/](features/README.md). Read the index, then the matching file, before driving.

## Launch

From the repo root, once per verification session:

```
go build -buildvcs=false -o .cursor/skills/verify-permitdenied/bin/control-permitdenied.exe ./cmd/control-permitdenied
```

Ready: the helper prints nothing and exits 0. Each drive is a new process with a fresh `Game` at the title scene. Two drives never share state. They can run sequentially; do not run two `drive` processes against the same `--out` directory.

Teardown is process exit. See Cleanup.

## Doctor

Run before the first drive, after any failed drive, and whenever the helper looks wrong:

```
.cursor/skills/verify-permitdenied/bin/control-permitdenied.exe doctor
```

Worth driving when stdout JSON has `"ok": true`, `"title": "PERMIT DENIED"`, `"scene": "title"`, `"screen": "320x224"`, `"instance": "in-process"`. Refuse if `ok` is false or `player_exe` is anything other than `not used`.

## Drive

Keys the helper accepts: `w` `a` `s` `d` `space` `enter` `esc` `f1` `f2` `f3`. `space` and `enter` are edges (one tick). Holds of `w`/`a`/`s`/`d` stay down for N ticks at 60 TPS.

```
.cursor/skills/verify-permitdenied/bin/control-permitdenied.exe drive --out .cursor/skills/verify-permitdenied/artifacts/<run-id> .cursor/skills/verify-permitdenied/scripts/<feature>.script
```

Pick `<run-id>` unique per drive (`title-start-1`, not `latest`). Exit 0 plus a dump/PNG pair is the only pass. A passing `go test` is not a substitute for a mapped user path.

Stable handles (on-screen copy, not pixels):

| Surface | Handle |
|---|---|
| Title | `PERMIT DENIED`, `THE COUNTY SAID NO.`, `PRESS SPACE` |
| Play HUD | clock `m:ss`, `$N`, `BLADE UP` / `BLADE DOWN` |
| Tally | `ENGINE COOKED` / `TRACK THROWN` / `COUNTY CLOCK`, `SPACE / TAP — AGAIN` |
| Dump fields | `scene`, `stance`, `blade`, `x`, `y`, `heading`, `sheriff_hp`, `death` |

Start every recipe from title unless its preconditions say otherwise. `esc` in play returns to title and wipes the run.

## Evidence

Write under `.cursor/skills/verify-permitdenied/artifacts/<run-id>/`. Keep dumps and PNGs; do not commit them.

Proof bar:

- Exercise the keyed user path (title → space/enter → WASD/space → death/tally). Do not teleport the dozer, set HP, or call `startRun` from outside the helper.
- Capture the action and the result: dump+PNG before the key, dump+PNG after.
- JSON `scene` / `stance` / `death` / HP must match the PNG HUD. PNG is 320×224 and must show `PERMIT DENIED` on title or the blade stance on play/tally.
- `go test ./...` is the unit box. This skill is the live box. A live pass that skipped a map entry point is incomplete.

## Cleanup

```
.cursor/skills/verify-permitdenied/bin/control-permitdenied.exe cleanup --out .cursor/skills/verify-permitdenied/artifacts/<run-id>
```

This confirms the artifact directory exists and deletes nothing in it. Drive already exited; there is no PID to kill. Never `taskkill` by `permitdenied.exe` / `control-permitdenied.exe`. You may delete `bin/control-permitdenied.exe` after the session. You may not delete artifacts.

## Helpers

`cmd/control-permitdenied` is verification scaffolding, not the player. CI still packages only `./cmd/permitdenied`.

Script commands (one per line, `#` comments):

| Command | Effect |
|---|---|
| `tap KEY` | one tick with that key's edge |
| `hold KEY[,KEY...] TICKS` | hold for TICKS (edges fire on tick 0) |
| `wait TICKS` | zero input |
| `shot RELPATH` | PNG of `Draw` |
| `dump RELPATH` | JSON snapshot |
| `assert field=value` | `= != < > <= >=` against the snapshot |

`assert blade=up` / `blade=down` and `assert stance=BLADE UP` are the HUD checks. `assert title=PERMIT DENIED` is identity.
