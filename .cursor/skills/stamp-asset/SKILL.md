---
name: stamp-asset
description: Reserve or stamp one PERMIT DENIED sprite frame or tileset cell without moving locked slots. Use when adding or replacing atlas art, tileset pixels, or when an agent generated a master PNG.
---

# Stamp one slot

1. Read assets/LEDGER.md and assets/usable/sprites.json.
2. If the name/id is locked and the human did not name that slot, stop.
3. New content: `-reserve` a non-overlapping rect first, append LEDGER row `reserved`, write the master under assets/src/sprites/<name>.png (or tiles/<id>_<name>.png) at exact w×h.
4. `go run ./cmd/stamp-asset -frame <name>` or `-tile <id>`.
5. Mark the LEDGER row `locked`.
6. `go test ./internal/render ./cmd/stamp-asset`.

Do not run a rebuild-all. Do not ask an image model to redraw a locked sheet. Do not GeoM.Rotate atlas facings. Filter nearest.
