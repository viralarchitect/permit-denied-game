---
name: gen-asset-master
description: Write a single Grok/image-model prompt for one PERMIT DENIED master sprite or tile. Use when the user wants new art for a named frame or tile id.
---

# One master, one prompt

Do not generate a sheet. Generate one subject at the slot size.

Read assets/LEDGER.md, assets/usable/ASSETS.md, internal/render/palette.go. Copy the locked neighbor of the same family as style reference (e.g. dozer_up_00 for other dozer facings).

Prompt must include:
- exact pixel size
- transparent background
- SNES/Genesis nearest-neighbor, no smoothing, no drop shadow outside the box
- palette hexes from palette.go
- facing lock if relevant (0=north, 4=east, 8=south, 12=west, already rotated in-frame)
- “do not draw other frames, HUD, or text”

Save as assets/src/sprites/<name>.png or assets/src/tiles/<id>_<name>.png. Then use skill stamp-asset. Leave usable/ alone until stamp succeeds.
