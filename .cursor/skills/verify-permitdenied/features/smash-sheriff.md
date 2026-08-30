# Smash sheriff

Smash sheriff is the mid-drag named target. Blade-down contact drops its HP; rubble awards the PIT-revoked boon.

## Sub-features

- `sheriff-path` drives from the south gate around spawn concrete toward the amber SHERIFF pad.
- `sheriff-wreck` blade-down contact drops `sheriff_hp` below spawn HP (60).
- `sheriff-boon` full rubble sets dump `sheriff=rubble` and `pit=false`.

## How to get to it (user POV)

- From spawn, go north up the drag. The amber pip is the sheriff office.
- Skirt the wet concrete just north of the gate (it occupies the spawn X).
- Drop the blade and push the building until it is rubble.

## Driving it with control-permitdenied

Preconditions:

- Doctor is green.
- Fresh drive from title.
- Blade down before the push.

- **Path and wreck.** Run `drive --out <out> .cursor/skills/verify-permitdenied/scripts/smash-sheriff.script`. `at-sheriff.json` is still `scene=play`. `after-push.json` has `sheriff_hp<60`.
- **Boon (if HP reached 0).** `sheriff=rubble`, `pit=false`, banner `PERMIT: PIT MANEUVER REVOKED`. A partial wreck (`sheriff_hp<60` with `sheriff=intact` or `cracked`) still counts as `sheriff-wreck`; it does not count as `sheriff-boon`.
- **Proof.** `after-push.json` plus `after-push.png`. Teleporting the dozer onto the pad is not this path.

## Gotchas

- Cruisers spawn at 0:00 and hunt sides/rear while blade is up. This script drops the blade immediately. A peel (`plates<4`) is a real fail — rerun, do not skip to a dump that cheated position.
- Spawn concrete is a wall on the straight-north line. The script steers A before a long W so X moves west of 256.
- Sheriff spawn HP is 60. `sheriff_hp<60` is the wreck assert; `sheriff_hp<300` would pass without contact.
