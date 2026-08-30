# Tank-steer

Tank-steer is vehicle-relative: A/D rotate about the machine, W goes forward along heading, S reverses. Spawn faces north.

## Sub-features

- `steer-w` holds W for one second from spawn; Y decreases (north).
- `steer-a` holds A for half a second from spawn; heading is west-of-north.
- `steer-d` is the east-of-north mirror of `steer-a` (same ticks, opposite sign).

## How to get to it (user POV)

- From title, press Space to spawn at the south gate facing north.
- Hold `W` to roll forward. Hold `S` to reverse.
- Hold `A` / `D` to tank-steer (machine left / right, not screen left).

## Driving it with control-permitdenied

Preconditions:

- Doctor is green.
- Two fresh drives; do not chain W and A in one process if you need spawn heading.

- **Forward north.** Run `drive --out <out> .cursor/skills/verify-permitdenied/scripts/tank-steer.script`. `spawn.json` has `y=1180`. `after-w.json` has `scene=play` and `y<1140`. `after-w.png` is not the title card.
- **Turn west.** Run `drive --out <out> .cursor/skills/verify-permitdenied/scripts/tank-steer-a.script`. `after-a.json` has `heading<0` and `heading>-3.1416`.
- **Proof.** Both drives record dump+PNG at spawn and after the hold. Y after W must be less than spawn Y; heading after A must be negative.

## Gotchas

- Wet concrete sits on the north line from spawn (`x` 256–336, `y` 1100–1124). One second of W stays south of a hard pin; longer W without steering hits the patch.
- A is machine-left. At north that is west (negative heading), not screen-left of a rotated camera — the camera does not yaw.
- Blade-down crawl is slower. These scripts start blade-up; do not tap Space twice.
