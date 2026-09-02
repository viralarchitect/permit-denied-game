# Asset LEDGER

Shipped art lives in `assets/usable/`. Masters live in `assets/src/` and are **not** embedded.
`sprites.json` name -> rect is a lock. Tile IDs 0-18 are frozen.
Never repack. Never reuse a tile ID. Never rewrite `lot.json` to refresh art.
Stamp only through `go run ./cmd/stamp-asset`. `cmd/genfx` may overwrite `boom_*` / `spark_*` pixels in their current rects and must not move those rects.

## How to add a slot

1. Reserve a non-overlapping rect (`go run ./cmd/stamp-asset -reserve name=x,y,w,h`), or grow the sheet downward / add `sprites_b.png` later.
2. Append a LEDGER row with `status=reserved`.
3. Write one master at exact w x h under `assets/src/sprites/<name>.png` or `assets/src/tiles/<id>_<name>.png`.
4. `go run ./cmd/stamp-asset -frame <name>` or `-tile <id>`.
5. Mark the LEDGER row `locked`.
6. `go test ./internal/render ./cmd/stamp-asset`.

Next free tile id is **19** at `x=48,y=32` (same row as 16-18). Grow `tileset.png` only if the decoded PNG is shorter than that cell. New sprite frames: append below y=256 or pack into free gaps without overlap.

## Sprites (`sprites.png`)

| name | sheet | x | y | w | h | status |
|---|---|---:|---:|---:|---:|---|
| dozer_up_00 | sprites.png | 0 | 0 | 32 | 32 | locked |
| dozer_up_01 | sprites.png | 32 | 0 | 32 | 32 | locked |
| dozer_up_02 | sprites.png | 64 | 0 | 32 | 32 | locked |
| dozer_up_03 | sprites.png | 96 | 0 | 32 | 32 | locked |
| dozer_up_04 | sprites.png | 128 | 0 | 32 | 32 | locked |
| dozer_up_05 | sprites.png | 160 | 0 | 32 | 32 | locked |
| dozer_up_06 | sprites.png | 192 | 0 | 32 | 32 | locked |
| dozer_up_07 | sprites.png | 224 | 0 | 32 | 32 | locked |
| dozer_up_08 | sprites.png | 256 | 0 | 32 | 32 | locked |
| dozer_up_09 | sprites.png | 288 | 0 | 32 | 32 | locked |
| dozer_up_10 | sprites.png | 320 | 0 | 32 | 32 | locked |
| dozer_up_11 | sprites.png | 352 | 0 | 32 | 32 | locked |
| dozer_up_12 | sprites.png | 384 | 0 | 32 | 32 | locked |
| dozer_up_13 | sprites.png | 416 | 0 | 32 | 32 | locked |
| dozer_up_14 | sprites.png | 448 | 0 | 32 | 32 | locked |
| dozer_up_15 | sprites.png | 480 | 0 | 32 | 32 | locked |
| dozer_down_00 | sprites.png | 0 | 32 | 32 | 32 | locked |
| dozer_down_01 | sprites.png | 32 | 32 | 32 | 32 | locked |
| dozer_down_02 | sprites.png | 64 | 32 | 32 | 32 | locked |
| dozer_down_03 | sprites.png | 96 | 32 | 32 | 32 | locked |
| dozer_down_04 | sprites.png | 128 | 32 | 32 | 32 | locked |
| dozer_down_05 | sprites.png | 160 | 32 | 32 | 32 | locked |
| dozer_down_06 | sprites.png | 192 | 32 | 32 | 32 | locked |
| dozer_down_07 | sprites.png | 224 | 32 | 32 | 32 | locked |
| dozer_down_08 | sprites.png | 256 | 32 | 32 | 32 | locked |
| dozer_down_09 | sprites.png | 288 | 32 | 32 | 32 | locked |
| dozer_down_10 | sprites.png | 320 | 32 | 32 | 32 | locked |
| dozer_down_11 | sprites.png | 352 | 32 | 32 | 32 | locked |
| dozer_down_12 | sprites.png | 384 | 32 | 32 | 32 | locked |
| dozer_down_13 | sprites.png | 416 | 32 | 32 | 32 | locked |
| dozer_down_14 | sprites.png | 448 | 32 | 32 | 32 | locked |
| dozer_down_15 | sprites.png | 480 | 32 | 32 | 32 | locked |
| dozer_plate_paint | sprites.png | 0 | 64 | 32 | 32 | locked |
| dozer_plate_paint_down | sprites.png | 128 | 64 | 32 | 32 | locked |
| dozer_plate_primer | sprites.png | 32 | 64 | 32 | 32 | locked |
| dozer_plate_primer_down | sprites.png | 160 | 64 | 32 | 32 | locked |
| dozer_plate_rust | sprites.png | 64 | 64 | 32 | 32 | locked |
| dozer_plate_rust_down | sprites.png | 192 | 64 | 32 | 32 | locked |
| dozer_plate_frame | sprites.png | 96 | 64 | 32 | 32 | locked |
| dozer_plate_frame_down | sprites.png | 224 | 64 | 32 | 32 | locked |
| dozer_heat | sprites.png | 256 | 64 | 32 | 32 | locked |
| cruiser_00 | sprites.png | 0 | 96 | 16 | 16 | locked |
| cruiser_01 | sprites.png | 16 | 96 | 16 | 16 | locked |
| cruiser_02 | sprites.png | 32 | 96 | 16 | 16 | locked |
| cruiser_03 | sprites.png | 48 | 96 | 16 | 16 | locked |
| cruiser_04 | sprites.png | 64 | 96 | 16 | 16 | locked |
| cruiser_05 | sprites.png | 80 | 96 | 16 | 16 | locked |
| cruiser_06 | sprites.png | 96 | 96 | 16 | 16 | locked |
| cruiser_07 | sprites.png | 112 | 96 | 16 | 16 | locked |
| cruiser_08 | sprites.png | 128 | 96 | 16 | 16 | locked |
| cruiser_09 | sprites.png | 144 | 96 | 16 | 16 | locked |
| cruiser_10 | sprites.png | 160 | 96 | 16 | 16 | locked |
| cruiser_11 | sprites.png | 176 | 96 | 16 | 16 | locked |
| cruiser_12 | sprites.png | 192 | 96 | 16 | 16 | locked |
| cruiser_13 | sprites.png | 208 | 96 | 16 | 16 | locked |
| cruiser_14 | sprites.png | 224 | 96 | 16 | 16 | locked |
| cruiser_15 | sprites.png | 240 | 96 | 16 | 16 | locked |
| dump_00 | sprites.png | 24 | 112 | 24 | 24 | locked |
| dump_04 | sprites.png | 72 | 112 | 24 | 24 | locked |
| dump_08 | sprites.png | 120 | 112 | 24 | 24 | locked |
| dump_12 | sprites.png | 168 | 112 | 24 | 24 | locked |
| dump_02 | sprites.png | 48 | 112 | 24 | 24 | locked |
| dump_06 | sprites.png | 96 | 112 | 24 | 24 | locked |
| dump_10 | sprites.png | 144 | 112 | 24 | 24 | locked |
| dump_14 | sprites.png | 192 | 112 | 24 | 24 | locked |
| jersey | sprites.png | 0 | 136 | 16 | 16 | locked |
| jersey_broken | sprites.png | 16 | 136 | 16 | 16 | locked |
| excavator_00 | sprites.png | 0 | 152 | 32 | 32 | locked |
| excavator_02 | sprites.png | 32 | 152 | 32 | 32 | locked |
| excavator_04 | sprites.png | 64 | 152 | 32 | 32 | locked |
| excavator_06 | sprites.png | 96 | 152 | 32 | 32 | locked |
| excavator_08 | sprites.png | 128 | 152 | 32 | 32 | locked |
| excavator_10 | sprites.png | 160 | 152 | 32 | 32 | locked |
| excavator_12 | sprites.png | 192 | 152 | 32 | 32 | locked |
| excavator_14 | sprites.png | 224 | 152 | 32 | 32 | locked |
| chopper_0 | sprites.png | 256 | 96 | 16 | 16 | locked |
| chopper_1 | sprites.png | 272 | 96 | 16 | 16 | locked |
| chopper_2 | sprites.png | 288 | 96 | 16 | 16 | locked |
| chopper_3 | sprites.png | 304 | 96 | 16 | 16 | locked |
| spotlight | sprites.png | 320 | 88 | 48 | 48 | locked |
| ped | sprites.png | 368 | 96 | 8 | 8 | locked |
| building_intact | sprites.png | 0 | 184 | 16 | 16 | locked |
| building_cracked | sprites.png | 16 | 184 | 16 | 16 | locked |
| building_rubble | sprites.png | 32 | 184 | 16 | 16 | locked |
| pip_amber | sprites.png | 48 | 184 | 4 | 4 | locked |
| pip_rust | sprites.png | 56 | 184 | 4 | 4 | locked |
| pip_cyan | sprites.png | 64 | 184 | 4 | 4 | locked |
| dollar | sprites.png | 72 | 184 | 8 | 8 | locked |
| boom_00 | sprites.png | 0 | 208 | 48 | 48 | locked |
| boom_01 | sprites.png | 48 | 208 | 48 | 48 | locked |
| boom_02 | sprites.png | 96 | 208 | 48 | 48 | locked |
| boom_03 | sprites.png | 144 | 208 | 48 | 48 | locked |
| boom_04 | sprites.png | 192 | 208 | 48 | 48 | locked |
| boom_05 | sprites.png | 240 | 208 | 48 | 48 | locked |
| spark_00 | sprites.png | 288 | 208 | 16 | 16 | locked |
| spark_01 | sprites.png | 304 | 208 | 16 | 16 | locked |
| spark_02 | sprites.png | 320 | 208 | 16 | 16 | locked |
| spark_03 | sprites.png | 336 | 208 | 16 | 16 | locked |
| fire_00 | sprites.png | 0 | 256 | 32 | 16 | locked |
| fire_02 | sprites.png | 32 | 256 | 32 | 16 | locked |
| fire_04 | sprites.png | 64 | 256 | 32 | 16 | locked |
| fire_06 | sprites.png | 96 | 256 | 32 | 16 | locked |
| fire_08 | sprites.png | 128 | 256 | 32 | 16 | locked |
| fire_10 | sprites.png | 160 | 256 | 32 | 16 | locked |
| fire_12 | sprites.png | 192 | 256 | 32 | 16 | locked |
| fire_14 | sprites.png | 224 | 256 | 32 | 16 | locked |
| wagon_00 | sprites.png | 0 | 272 | 48 | 24 | locked |
| wagon_02 | sprites.png | 48 | 272 | 48 | 24 | locked |
| wagon_04 | sprites.png | 96 | 272 | 48 | 24 | locked |
| wagon_06 | sprites.png | 144 | 272 | 48 | 24 | locked |
| wagon_08 | sprites.png | 192 | 272 | 48 | 24 | locked |
| wagon_10 | sprites.png | 240 | 272 | 48 | 24 | locked |
| wagon_12 | sprites.png | 288 | 272 | 48 | 24 | locked |
| wagon_14 | sprites.png | 336 | 272 | 48 | 24 | locked |

## Tiles (`tileset.png`)

Cell formula: `x=(id%8)*16`, `y=(id/8)*16`, size 16x16. Columns = 8.

| name | sheet | x | y | w | h | status |
|---|---|---:|---:|---:|---:|---|
| 0_empty | tileset.png | 0 | 0 | 16 | 16 | locked |
| 1_dirt_0 | tileset.png | 16 | 0 | 16 | 16 | locked |
| 2_dirt_1 | tileset.png | 32 | 0 | 16 | 16 | locked |
| 3_dirt_2 | tileset.png | 48 | 0 | 16 | 16 | locked |
| 4_dirt_3 | tileset.png | 64 | 0 | 16 | 16 | locked |
| 5_asphalt | tileset.png | 80 | 0 | 16 | 16 | locked |
| 6_asphalt_dash | tileset.png | 96 | 0 | 16 | 16 | locked |
| 7_asphalt_edge_w | tileset.png | 112 | 0 | 16 | 16 | locked |
| 8_asphalt_edge_e | tileset.png | 0 | 16 | 16 | 16 | locked |
| 9_asphalt_stop | tileset.png | 16 | 16 | 16 | 16 | locked |
| 10_asphalt_oil | tileset.png | 32 | 16 | 16 | 16 | locked |
| 11_pad | tileset.png | 48 | 16 | 16 | 16 | locked |
| 12_rail | tileset.png | 64 | 16 | 16 | 16 | locked |
| 13_rail_tie | tileset.png | 80 | 16 | 16 | 16 | locked |
| 14_concrete_wet | tileset.png | 96 | 16 | 16 | 16 | locked |
| 15_concrete_set | tileset.png | 112 | 16 | 16 | 16 | locked |
| 16_skid | tileset.png | 0 | 32 | 16 | 16 | locked |
| 17_gate | tileset.png | 16 | 32 | 16 | 16 | locked |
| 18_gravel | tileset.png | 32 | 32 | 16 | 16 | locked |

