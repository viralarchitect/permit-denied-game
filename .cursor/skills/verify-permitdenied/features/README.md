# PERMIT DENIED verification map

This directory is the maintained source for verifying the player-facing behavior of PERMIT DENIED. Read this index before driving, then use the matching feature file as the recipe.

## Baseline preconditions

- Work from the repo root.
- Build `control-permitdenied` once per session (see the skill Launch section).
- Run `control-permitdenied doctor` and require `ok`, title `PERMIT DENIED`, scene `title`, instance `in-process`.
- Each drive is a new process. `--out` is `.cursor/skills/verify-permitdenied/artifacts/<run-id>` and must be unique.
- Never launch or keystroke `dist\permitdenied.exe`.

## Driving conventions

- Start every recipe from the title scene unless its preconditions say otherwise.
- Treat script lines as literal. Keep key names and assert fields unchanged.
- 60 TPS: `hold w 60` is one second of forward.
- Restore nothing between drives — there is no disk state. Start a new drive instead.
- Cleanup must leave the artifact directory in place.

## Proof and skip reporting

- Capture dump+PNG before the action and dump+PNG after.
- Play proof includes `stance` matching the HUD (`BLADE UP` / `BLADE DOWN`) and a 320×224 PNG.
- Title proof includes `scene=title` and `title=PERMIT DENIED` with the wordmark visible in the PNG.
- Tally proof includes `death` (`cooked` / `track` / `buzzer`) matching the tally heading.
- Record the feature ID and script path with every artifact directory name.
- An unreachable path is reported with the command run and the unmet assert, not as verified through a different script.

## Feature entry contract

Each feature file starts with an H1 title and one paragraph describing the user-visible behavior. It then uses exactly four H2 sections in this order.

1. `Sub-features` lists short IDs with one line for each behavior.
2. `How to get to it (user POV)` lists every user entry point.
3. `Driving it with control-permitdenied` starts with `Preconditions:` and uses labeled bullets that pair each user action with an exact command and observable result.
4. `Gotchas` lists traps that can waste or invalidate a verification run.

Keep implementation details out of the map. Name only user paths, stable handles, required state, commands, and observable proof.

## Features

- [Title and start](./title-start.md) covers the title card, Space/Enter into play, and Esc back to title.
- [Tank-steer](./tank-steer.md) covers W north from spawn and A heading west-of-north.
- [Blade](./blade.md) covers Space toggling BLADE UP / BLADE DOWN on the HUD.
- [Smash sheriff](./smash-sheriff.md) covers driving the south-gate → sheriff path with blade down until HP drops.
- [Tally](./tally.md) covers cooking the engine on spawn concrete and reading ENGINE COOKED.
