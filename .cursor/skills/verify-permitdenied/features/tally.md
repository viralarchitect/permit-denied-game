# Tally

Tally is the end card. Death copy is `ENGINE COOKED`, `TRACK THROWN`, or `COUNTY CLOCK`. Space or Enter starts a new run.

## Sub-features

- `tally-cook` blade-down into spawn concrete until `death=cooked`.
- `tally-copy` the overlay heading matches that death.
- `tally-again` Space on tally starts a fresh play at spawn.

## How to get to it (user POV)

- Cook: hold blade-down against a wall or deep rubble until the engine dies.
- Track: lose the last plate.
- Clock: survive 3:30.
- On the overlay, press Space or Enter to go again.

## Driving it with control-permitdenied

Preconditions:

- Doctor is green.
- Fresh drive from title.

- **Cook on spawn concrete.** Run `drive --out <out> .cursor/skills/verify-permitdenied/scripts/tally.script`. The script starts, drops the blade, holds W for 720 ticks. `tally.json` has `scene=tally` and `death=cooked`. `tally.png` shows `ENGINE COOKED`.
- **Again.** From that state in a follow-up script: `tap space`, `dump again.json`, `assert scene=play`, `assert y=1180`, `assert death=`.
- **Proof.** Tally PNG must show the heading, not only the play HUD behind it. `death=cooked` with a play-scene PNG fails.

## Gotchas

- Tally rolls in copy over ~1.4s of tally time. `wait 90` after `scene=tally` if `tally.png` still lacks `TOTAL` / `SPACE / TAP — AGAIN`. `ENGINE COOKED` appears immediately.
- F3 skips 15s of clock and is a debug cheat. Do not use it as the cook path.
- Blade-up on asphalt cools. The cook script must drop the blade before the W hold.
