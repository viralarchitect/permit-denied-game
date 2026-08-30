# Blade

Blade is a single toggle. The HUD reads `BLADE UP` or `BLADE DOWN`. Down wrecks; up glances.

## Sub-features

- `blade-hud` shows `BLADE UP` at spawn.
- `blade-toggle` Space in play flips the stance and the dump `blade` field.
- `blade-toggle-back` a second Space returns to `BLADE UP`.

## How to get to it (user POV)

- Start a run. The HUD bottom-left reads `BLADE UP`.
- Press `Space` to drop the blade. Press `Space` again to raise it.

## Driving it with control-permitdenied

Preconditions:

- Doctor is green.
- Fresh drive from title.

- **Spawn stance.** Run `drive --out <out> .cursor/skills/verify-permitdenied/scripts/blade.script`. After the opening Space, `blade-up.json` has `blade=up` and `stance=BLADE UP`. `blade-up.png` shows `BLADE UP`.
- **Drop blade.** The script taps Space in play. `blade-down.json` has `blade=down` and `stance=BLADE DOWN`. `blade-down.png` shows `BLADE DOWN`.
- **Raise blade.** A third Space (second in play) returns `blade=up`.
- **Proof.** Both PNGs show the matching stance string. A dump without the PNG is incomplete.

## Gotchas

- The first Space of a drive starts the run and must not be counted as a blade toggle.
- Holding Space does not auto-repeat; only the first tick of a `hold space N` toggles.
- Blade-up vs a building glances (speed reverses a fraction). Blade-down wrecks HP. This feature only proves the toggle, not wreck.
