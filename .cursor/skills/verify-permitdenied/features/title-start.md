# Title and start

Title and start shows the `PERMIT DENIED` card, starts a run with Space or Enter, and returns to the card with Esc.

## Sub-features

- `title-card` shows the wordmark and `PRESS SPACE`.
- `title-space` starts a run from Space.
- `title-enter` starts a run from Enter.
- `title-esc` returns from play to title and wipes the run.

## How to get to it (user POV)

- Launch the game; it opens on the title card.
- Press `Space` or `Enter` to start.
- Press `Esc` during play to quit to title.

## Driving it with control-permitdenied

Preconditions:

- Doctor reports `ok`, `title=PERMIT DENIED`, `scene=title`.
- `--out` is a new artifact directory.

- **Title card.** Capture the card before any key. Run `drive --out <out> .cursor/skills/verify-permitdenied/scripts/title-start.script`. `title.json` has `scene=title` and `title=PERMIT DENIED`. `title.png` shows `PERMIT DENIED` and `PRESS SPACE`.
- **Start with Space.** The same script taps `space`. `play.json` has `scene=play`, `stance=BLADE UP`, `x=320`, `y=1180`, `clock=0:00`, `plates=4`. `play.png` shows `BLADE UP` and `0:00`.
- **Start with Enter.** Fresh drive: `tap enter`, `dump enter.json`, `assert scene=play`.
- **Esc to title.** From play: `tap esc`, `dump esc.json`, `assert scene=title`. `esc.json` has spawn zeros and no `BLADE UP` HUD.

## Gotchas

- Space on title starts the run. Space in play toggles the blade. They are not the same key effect.
- Esc destroys the run. A second Space after Esc is a new spawn, not a resume.
- Identity is the title dump `title=PERMIT DENIED` plus the wordmark in `title.png`. A play-scene PNG alone is not title proof.
