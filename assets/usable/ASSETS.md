# PERMIT DENIED — usable tilemap + sprite atlas

Drop this folder into the game repo as `assets/` (or `internal/render/assets/`).  
These files match **current `lot.go` / `const.go` coordinates**. They do **not** replace collision.

## What this is

| File | Role |
|---|---|
| `tileset.png` | 16×16 tiles, 8 columns, 19 tiles |
| `tileset.tsj` | Tiled tileset |
| `lot.tmj` | Tiled map (open in Tiled) |
| `lot.json` | Game-facing map (no GID +1) |
| `lot_ground.csv` / `lot_decal.csv` | Same grids, 0 = empty |
| `sprites.png` + `sprites.json` | Named frames, 16 facings |
| `lot_preview.png` | Full 640×1280 composite |
| `view_*_3x.png` | Spawn / sheriff / plant / yard crops |

Older files in `../` (`lot_40x80.csv`, `atlas.json`) are a one-layer draft. Do not load those.

## Contract with the sim

1. **Tiles draw the ground. Objects own AABBs.**  
   Buildings, jersey, dumps, concrete, spawn, excavator cells live on the `objects` layer. Do not bake buildings into the tile grid and then also spawn `lot.Building` from literals — you will double-draw.

2. **Until a loader exists, `lot.go` literals stay source of truth.**  
   The object layer is a *mirror* of those literals so Tiled and the binary cannot drift silently. First loader milestone: parse `lot.json` objects into `lot.New`, then delete the duplicate literals.

3. **Do not use the tile layer as collision.**  
   Dozer body is a circle. Buildings / rubble / blockers are AABBs. A tile solid grid would fight `CircleAABB` and rubble-as-maze.

4. **16 facings are draw-only.**  
   Physics keeps continuous `heading`. Draw uses `FacingIndex(heading)` → `dozer_up_00`..`15` or `dozer_down_00`..`15`.  
   If you blit a facing frame, **do not also `GeoM.Rotate`**. Rotation is already baked.

5. **Filter is nearest.** Logical screen 320×224.

## Tile IDs (`tileset.png`)

| ID | Name | Use |
|---|---|---|
| 0 | empty | decal transparent |
| 1–4 | dirt_0..3 | shoulders |
| 5 | asphalt | drag fill |
| 6 | asphalt_dash | centerline |
| 7 | asphalt_edge_w | drag x=240 |
| 8 | asphalt_edge_e | drag x=384 |
| 9 | asphalt_stop | unused reserve |
| 10 | asphalt_oil | decal speckle |
| 11 | pad | plant pad |
| 12 | rail | NE spur |
| 13 | rail_tie | NE spur |
| 14 | concrete_wet | decal only |
| 15 | concrete_set | swap on `ConcreteSets` at 2:15 |
| 16 | skid | reserve |
| 17 | gate | south-gate stop bar decal |
| 18 | gravel | drag shoulder |

Drag in tiles: `tx = 15..24` (world x 240..400). Lot is 40×80 tiles = 640×1280.

Concrete decals are **paint**. Solidity still comes from `threats.Blocker`.

## Object layer (world px, top-left of rect, spawn is center)

Mirrors `internal/lot/lot.go` + `initialBlockers` + beat drop points:

- `SHERIFF` 176,640 96×80 value 400
- `YARD` 448,720 128×96 value 500
- `PLANT` 248,72 144×112 value 700
- shacks as in `lot.go`
- start jersey / dumps / wet concrete
- `when=t40` jersey + dump (beat chart)
- `excavator_primary` 288,520 heading π
- `excavator_fallback` 288,360
- `spawn` 320,1180 heading 0

`when` is documentation for `beat.go`. The map does not fire the clock.

## Sprite atlas (`sprites.json`)

Frame name pattern:

```
dozer_up_00 .. dozer_up_15     // 32×32, anchor 16,16, facing 0 = north
dozer_down_00 .. dozer_down_15
dozer_plate_paint|primer|rust|frame[_down]   // reference only
dozer_heat
cruiser_00 .. cruiser_15       // 16×16
dump_00, dump_02, ... dump_14  // 24×24, even facings
jersey / jersey_broken
excavator_00, _02, ... _14
chopper_0 .. chopper_3         // rotor
spotlight                      // optional; current Draw uses vector circle
ped
pip_amber / pip_rust / pip_cyan
building_intact|cracked|rubble // 16×16 stamp if you tile roofs later
dollar
```

Plate / heat: keep **one** 16-facing set (`dozer_up_*` / `dozer_down_*`) and `ColorScale` toward primer / rust / frame / `#C04040`. The plate preview frames are for artists, not extra draw paths.

## Minimal Go load (do not invent a framework)

```go
//go:embed assets/lot.json assets/tileset.png assets/sprites.png assets/sprites.json
var fs embed.FS
```

Draw path:

1. Camera same as now (`dozer - 160, -112`, clamp).
2. Visible tile range only (you already do this in `drawGround`).
3. For each visible cell: blit `tileset` subrect for `ground`, then `decal`.
4. Buildings still from `lot.Buildings` until the loader ships.
5. Dozer: `name := fmt.Sprintf("dozer_%s_%02d", stance, FacingIndex(heading))`.

Swap `drawGround` first. Leave `drawDozer` on rotated rects until the atlas looks right in-game. One swap at a time.

## Tiled

Open `lot.tmj` with `tileset.tsj` beside it. Tiled GIDs are `localID + 1`. `lot.json` uses raw IDs (0 = empty). If you edit in Tiled, re-export or add a tiny convert step — do not hand-edit both.

## Out of scope

- Autotile bitmasks
- Locational armor frames
- Soft-body rubble tiles as collision
- A second lot
- Replacing `const.go` numbers from this file
