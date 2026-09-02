# PERMIT DENIED — Go / Ebitengine Implementation Guide

**For:** an implementing agent (or a human solo) building the one-lot show-off.  
**Authority:** this file owns *how*. [`../PERMIT_DENIED.md`](../PERMIT_DENIED.md) owns *what*. If they conflict, the design doc wins on fantasy; this file wins on numbers, signs, and APIs.  
**Do not expand the design.** No campaign, no meta, no second lot, no Heemeyer biography, no title `KILLDOZER`.

---

## 0. Agent contract (read first)

You are building **one** playable binary:

- Title on the window and title screen: **`PERMIT DENIED`**
- Engine: **Ebitengine v2** (`github.com/hajimehoshi/ebiten/v2`, use **v2.9.x**)
- Language: **Go 1.22+**
- Loop: 60 TPS, world-fixed north camera, tank-steer dozer, 3:30 gauntlet
- Art: **procedural canvas/rects first**. PNG atlases are optional polish after the lot is fun.
- Success: the five show-off tests in §19. If those fail, do not add systems.

**Forbidden without a human saying so**

- ECS frameworks, Godot, Raylib, Pixel, Fyne
- `net/http` game server, accounts, leaderboards, databases
- Soft-body, particle debris as collision, locational armor
- Chopper rockets, second player dozer, fuel-depot body boons
- Between-run unlocks
- `ColorM` (deprecated) — use `ColorScale`
- `ebiten.TouchIDs()` (deprecated) — use `ebiten.AppendTouchIDs`

**Implementation order is §18.** Do not write threats before the dozer can wreck a building and die to heat.

---

## 1. Stack lock

| Piece | Use this |
|---|---|
| Module | `permitdenied` (or `github.com/<user>/permitdenied`) |
| Engine | `github.com/hajimehoshi/ebiten/v2` **v2.9.10 or later v2.9** |
| Text | `github.com/hajimehoshi/ebiten/v2/text/v2` |
| Vector (debug + procedural) | `github.com/hajimehoshi/ebiten/v2/vector` |
| Audio | `github.com/hajimehoshi/ebiten/v2/audio` |
| Input | `ebiten.IsKeyPressed` / `IsKeyJustPressed` / `AppendTouchIDs` |
| Entry | `ebiten.RunGame(g)` once in `main` |
| Tick | default **60 TPS** (`ebiten.DefaultTPS`). Do not change. |
| Logical screen | **320 × 224** (SNES). Window starts at **1280 × 896** (4× integer). |
| Filter | `ebiten.FilterNearest` on all gameplay sprites |

`ebiten.Game` (mandatory three methods):

```go
type Game interface {
    Update() error
    Draw(screen *ebiten.Image)
    Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}
```

`Layout` **always** returns `(320, 224)` regardless of `outsideWidth/Height`. Integer scale is the window’s job (`SetWindowSize(1280, 896)`, `SetWindowResizingModeEnabled` optional). Do not letterbox inside `Draw` unless you must; prefer Ebitengine’s scale.

```go
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return 320, 224
}
```

`Update` is logic. `Draw` is render. **Never** mutate sim state in `Draw`. `Draw` may run at display Hz; `Update` is the clock of the beat chart.

Y in Ebitengine is **down**. `GeoM.Rotate(theta)` rotates **clockwise**. All headings in this doc use that convention.

---

## 2. Repository layout

```text
permitdenied/
  go.mod
  go.sum
  README.md                          # how to run, keys, show-off tests
  cmd/permitdenied/main.go           # window title, RunGame
  internal/
    game/game.go                     # scene switch, Update/Draw/Layout
    game/input.go                    # keyboard + touch → Input
    game/camera.go
    game/const.go                    # ALL tunables live here
    run/run.go                       # clock, score, freeze-frame, reset
    dozer/dozer.go
    lot/lot.go                       # strip geometry, buildings, rubble
    lot/building.go
    threats/cruiser.go
    threats/blocker.go
    threats/excavator.go
    threats/chopper.go
    threats/ped.go                   # garnish only
    fx/fx.go                         # shake, hit-stop, dollar ticks
    audio/audio.go                   # optional until milestone 6
    render/palette.go
    render/draw.go                   # world → screen
  testdata/                          # empty ok
```

No `pkg/`. No interfaces-for-one-implementation. Unexported fields unless a test in another package needs them.

`cmd/permitdenied/main.go` is ~25 lines:

```go
package main

import (
    "log"
    "github.com/hajimehoshi/ebiten/v2"
    "permitdenied/internal/game"
)

func main() {
    ebiten.SetWindowSize(1280, 896)
    ebiten.SetWindowTitle("PERMIT DENIED")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    g := game.New()
    if err := ebiten.RunGame(g); err != nil {
        log.Fatal(err)
    }
}
```

---

## 3. Coordinate system (do not improvise)

### 3.1 World

| Axis | Direction | Unit |
|---|---|---|
| +X | east, right on screen | world pixels |
| +Y | **south, down on screen** | world pixels |
| Origin | north-west corner of the lot | (0, 0) |

Lot size: **640 × 1280** world px (40 × 80 tiles of 16px).  
Logical camera: 320 × 224, centered on the dozer, **clamped** so it never shows past the lot.

North is **up the screen** = **decreasing Y**.

### 3.2 Heading

- `heading` is radians.
- `heading == 0` → **north** → forward `(0, -1)`.
- `heading` **increases clockwise** (Ebiten-native).
- Forward vector:

```go
func Forward(heading float64) (fx, fy float64) {
    return math.Sin(heading), -math.Cos(heading)
}
```

Proof table (write a unit test):

| heading | forward | world |
|---|---|---|
| 0 | (0, -1) | north |
| π/2 | (1, 0) | east |
| π | (0, 1) | south |
| 3π/2 | (−1, 0) | west |

Right vector (for side/rear tests):

```go
func Right(heading float64) (rx, ry float64) {
    return math.Cos(heading), math.Sin(heading)
}
```

At heading 0: right = (1, 0) = east. Correct.

### 3.3 16 facings (draw only)

Physics uses continuous `heading`. Draw snaps:

```go
func FacingIndex(heading float64) int {
    const step = 2 * math.Pi / 16
    h := math.Mod(heading+step/2, 2*math.Pi)
    if h < 0 {
        h += 2 * math.Pi
    }
    return int(h / step) // 0 = north, 4 = east, 8 = south, 12 = west
}
```

Until you have a 16-frame atlas, draw a yellow body rect rotated by `heading` plus a darker blade quad. Rotation: `op.GeoM.Translate(-cx, -cy); op.GeoM.Rotate(heading); op.GeoM.Translate(screenX, screenY)`.

### 3.4 Camera

World-fixed north. **Camera does not yaw.**

```go
camX = clamp(dozer.X - 160, 0, lotW-320)
camY = clamp(dozer.Y - 112, 0, lotH-224)
screenX = worldX - camX + shakeX
screenY = worldY - camY + shakeY
```

---

## 4. Input

Unify keyboard and touch into one struct **once per Update**, then simulate from that. Never read `IsKeyPressed` from `dozer` or `threats`.

```go
type Input struct {
    Throttle    float64 // -1 reverse … +1 forward
    Steer       float64 // -1 vehicle-left … +1 vehicle-right
    BladeToggle bool    // edge, once per press
}

type Probe struct {
    Heading, X, Y, Speed float64
    BladeDown            bool
    Plates               int
    Heat                 float64
}
```

### 4.1 Keyboard (source of truth)

| Key | Action |
|---|---|
| `ebiten.KeyA` | Steer −1 (vehicle **left**, CCW, heading **decreases**) |
| `ebiten.KeyD` | Steer +1 (vehicle **right**, CW, heading **increases**) |
| `ebiten.KeyW` | Throttle +1 |
| `ebiten.KeyS` | Throttle −1 (reverse, first-class) |
| `ebiten.KeySpace` | `IsKeyJustPressed` → BladeToggle |
| `ebiten.KeyEnter` / Space on tally | Again |
| `ebiten.KeyEscape` | Title (optional) |

If A and D both held, Steer = 0. If W and S both held, Throttle = 0.

**Self-test (mandatory, spawn facing north, south gate):**

1. Hold **W** → Y decreases, dozer travels **up the drag** toward the plant. If Y increases, negate forward Y.
2. Hold **A** (optionally with W) → nose sweeps **west** (X decreases, heading goes negative). If the nose goes east, **negate Steer once**. Do not also flip `Forward`.
3. Hold **D** → nose sweeps **east**.
4. Hold **S** → travel south, reverse slower than forward.

This game’s camera is **not** a chase cam. “A = left” means **the machine’s left**, which at south-facing is screen-right. That is correct tank-steer. Do not “fix” it to screen-absolute left.

### 4.2 Touch (phone tiller)

Two non-overlapping zones. Screen space, logical 320×224.

| Zone | Rect | Meaning |
|---|---|---|
| Tiller | x < 140, y > 80 | Left thumb. Vector from pad center to finger = **desired heading**. Convert angular error to Steer (clamp ±1). Deadzone 8 px. |
| Throttle | x > 180, y > 80 | Right thumb **hold**. Finger Y relative to touch-down: up (smaller Y) = forward, down = reverse. Release → throttle 0. |
| Blade tap | right zone, movement < 12 px and duration < 200 ms on release | BladeToggle |

Do **not** implement auto-creep. Reverse must be reachable.

Track touches with:

```go
ids := ebiten.AppendTouchIDs(nil)
x, y := ebiten.TouchPosition(id)
```

Map logical coords: `lx = x * 320 / outsideW` (or use `ebiten.CursorPosition` equivalent; if you implement `Layout` fixed 320×224, `TouchPosition` is already in outside pixels — convert using the same scale Ebitengine uses, or store last `Layout` outside size on the Game).

### 4.3 Integrate

```go
const (
    TurnRateBladeUp   = 2.4 // rad/s
    TurnRateBladeDown = 1.1
    SpeedFwdUp        = 110 // px/s
    SpeedFwdDown      = 48
    SpeedRevUp        = 55
    SpeedRevDown      = 28
)

turn := TurnRateBladeUp
if d.BladeDown {
    turn = TurnRateBladeDown
}
d.Heading += in.Steer * turn * dt
// wrap to [0, 2π)

max := SpeedFwdUp
if in.Throttle < 0 {
    max = SpeedRevUp
    if d.BladeDown {
        max = SpeedRevDown
    }
} else if d.BladeDown {
    max = SpeedFwdDown
}
// approach max * Throttle with accel; see dozer.go
```

Accel: `180 px/s²` forward, `220 px/s²` brake-to-zero, reverse accel `120`. Dozer is heavy. Instant max-speed is wrong.

`dt = 1.0 / 60.0` every Update. Do not use frame delta from Draw.

---

## 5. `internal/game/const.go` — every tunable

```go
package game

const (
    TPS          = 60
    Dt           = 1.0 / 60.0
    ScreenW      = 320
    ScreenH      = 224
    Tile         = 16
    LotW         = 640
    LotH         = 1280

    RunSeconds   = 210.0 // 3:30
    BuzzerTick   = 210 * TPS

    // beat chart (seconds)
    TBlockers    = 40
    TChopper     = 60
    TExAnnounce  = 80
    TExArrive    = 105
    TConcreteSet = 135
    TTwoFamilies = 165

    DozerBodyR   = 14.0
    BladeW       = 32.0
    BladeHDown   = 10.0
    BladeHUp     = 6.0
    BladeReach   = 18.0 // center → blade center along forward

    PlatesMax    = 4
    HeatMax      = 100.0
    HeatCookPush = 28.0  // /s blade-down into wall or deep rubble
    HeatCookStall= 45.0  // /s wedged (speed < 8 and overlapping solid)
    HeatCoolAsphalt = 18.0 // /s blade-up, not overlapping solid
    HeatCoolIdle = 8.0

    WreckHPPerTile = 2.0 // each 16×16 of a building
    WreckRateDown  = 14.0 // HP/s while blade-down overlapping
    WreckRateUp    = 1.5  // glance

    CruiserSpeed   = 95.0
    CruiserRadius  = 8.0
    PedRadius      = 4.0

    HitStopTicks   = 2
    ShakeDecay     = 8.0

    Mult0 = 1.00
    Mult1 = 1.25
    Mult2 = 1.60
    Mult3 = 2.00
)
```

Do not sprinkle magic numbers elsewhere. If you need a new number, add it here and comment which spec line it implements.

---

## 6. Core types

```go
type Scene int
const (
    SceneTitle Scene = iota
    ScenePlay
    SceneTally
)

type BuildingState int
const (
    Intact BuildingState = iota
    Cracked
    Rubble // colliding
)

type TargetID int
const (
    TargetNone TargetID = iota
    TargetSheriff
    TargetYard
    TargetPlant
)

type Building struct {
    ID      TargetID // TargetNone for mundane
    Label   string   // "SHERIFF", "YARD", "PLANT", or ""
    X, Y, W, H float64
    HP, MaxHP  float64
    State      BuildingState
    Value      int // structure $ when fully rubble
}

type Rubble struct { // axis-aligned, colliding
    X, Y, W, H float64
}

type Dozer struct {
    X, Y    float64
    Heading float64
    Speed   float64 // signed: +forward along heading, −reverse
    BladeDown bool
    Plates  int     // 4 paint … 1 frame; 0 = dead this tick
    Heat    float64 // 0..100
    IFrames int     // ticks after a peel
}

type Cruiser struct {
    X, Y, Heading float64
    Alive bool
}

type BlockerKind int
const (
    BlockerJersey BlockerKind = iota
    BlockerDump
    BlockerConcrete // sets at TConcreteSet unless plant dead
)

type Blocker struct {
    Kind BlockerKind
    X, Y, W, H float64
    Set        bool    // concrete only
    HP         float64
}

type Excavator struct {
    X, Y, Heading float64
    Announced, Arrived, Alive bool
    BoomPhase float64 // 0..1 swipe
    Corridor  int
}

type Chopper struct {
    X, Y    float64
    Active  bool
    SpotR   float64 // 72
}

type Ped struct {
    X, Y    float64
    Alive   bool
}

type Dollar struct {
    X, Y float64
    Amt  int
    Life float64
}

type Run struct {
    Tick     int
    Over     bool
    Death    string // "", "cooked", "track", "buzzer"
    Struct$  int
    Vehicle$ int
    TimeAlive float64
    Targets  int // 0..3
    SheriffDown, YardDown, PlantDown bool
    CruiserPIT bool // true until sheriff smashed
    DumpTrucks bool
    WallsBrittle bool
    ConcreteSets bool
}
```

Booleans on `Run` are the boons. Default at `NewRun`:

- `CruiserPIT = true`
- `DumpTrucks = true`
- `WallsBrittle = false`
- `ConcreteSets = true`

---

## 7. Lot geometry (exact)

Tile = 16. Coordinates are top-left of rects unless noted. Dozer spawn is **center**.

```
Lot 640 × 1280

DRAG (asphalt): x=240..400, y=0..1280     // 160 px / 10 tiles wide
SOUTH GATE spawn: dozer center (320, 1180), heading 0 (north)
Shoulder dirt: everything else a sandy fill

SHERIFF (easy, mid-drag, west side of drag so you see it while rolling)
  rect  176, 640,  96, 80     // left of drag, label SHERIFF
  value $400

YARD (east side lot, behind a choke)
  choke jersey at 400, 780, 48, 16  (plugs the yard mouth)
  yard buildings:
    448, 720, 128, 96   // PUBLIC WORKS, target
    value $500
  dump spawn slots (if DumpTrucks): (456, 840), (520, 840)

PLANT (far north pad, greedy, open)
  pad asphalt 200, 48, 240, 200
  silos/building  248, 72, 144, 112   // BATCH PLANT
  value $700

MUNDANE smashables (give the drag something to eat)
  west shacks:  (64, 900, 64, 48), (80, 500, 80, 48), (48, 300, 64, 64)
  east shacks:  (448, 1000, 72, 48), (480, 560, 64, 40)
  each value $40–$80
  All start Intact. HP = (W/16)*(H/16)*WreckHPPerTile

RAIL SPUR (visual + one extra choke, north-east)
  y=200..216, x=400..640  dark ties, no extra system

BLOCKER DROP POINTS on the drag (first drop at 0:40)
  (248, 1000, 48, 16) jersey
  (344, 860, 48, 16) jersey
  (248, 480, 56, 16) jersey
  dump truck if DumpTrucks still true: (300, 720, 40, 24)

CONCRETE PATCHES (inert until TConcreteSet, then become solid if ConcreteSets)
  (256, 1100, 80, 24)
  (280, 400, 80, 24)
  (260, 200, 96, 32) // plant approach — the greedy tax

EXCAVATOR arrive cell (plants IN a corridor, does not chase)
  prefer the drag cell (288, 520) facing south
  if that cell is already rubble-blocked, plant at (288, 360)
  boom swipe: 48 px arc in front, 0.8 s period
```

**Visibility rule:** at spawn `(320, 1180)` with camera centered, the player must see the drag going north. Sheriff is off-screen until they drive; that is OK **only if** a 16-bit label or map pips mark all three targets on the HUD edge. Spec: “all three named targets visible from spawn.” Satisfy it with **edge pips** (three colored ticks on the screen rim pointing at sheriff/yard/plant) plus the buildings themselves when in view. Do not fog. Do not make the lot 200 px tall.

Pip color: sheriff amber, yard rust, plant cyan. Hide a pip once that target is rubble.

---

## 8. Simulation order (every `Update` in ScenePlay)

If `hitStop > 0`: decrement, skip this list except clock (clock still runs). Hit-stop is **one-frame juice**, max `HitStopTicks`.

1. Read `Input`.
2. `run.Tick++`. If `run.Tick >= BuzzerTick` → end `"buzzer"`.
3. Spawn threats according to beat chart (§11) **once** (flags).
4. Chopper linger check: if dozer inside spot, `run.Tick += 1` extra every 12 ticks (accelerates the chart ~8%). Hard cap: cannot skip past a beat that has not announced. Simpler allowed implementation: `nextBeatShift -= Dt` while inside, applied to remaining schedule offsets stored on `Run`. Pick **one** and comment it. Recommended: `Run.PressureBonus` float seconds, added into beat comparisons: `t := float64(run.Tick)/TPS + run.PressureBonus`.
5. Dozer steer / throttle / integrate **candidate** position.
6. Blade toggle (flip `BladeDown`).
7. Resolve dozer vs solids (buildings Intact/Cracked, Rubble, Blockers, Excavator body, lot bounds). Slide along; do not tunnel. If overlapping solid and `|speed| < 8` → stalled.
8. Blade-down wreck: any building/blocker overlapping the **blade rect** loses HP. Cracked at 50% HP. At 0: Intact/Cracked → **Rubble** of the same AABB (or inset 2 px), add `Value` to `Struct$`, spawn dollar tick, hit-stop, shake. If `ID != TargetNone`, fire boon (§10).
9. Blade-up glance: overlap body-vs-building reflects speed `*= -0.3`, tiny HP chip, no hit-stop.
10. Heat (§9).
11. Cruisers seek, collide (§12).
12. Excavator boom if arrived.
13. Chopper follows lazily toward dozer (max 70 px/s), not a dogfight.
14. Peds wander; body overlap + `|speed| > 20` → flatten, `$5`, dollar tick. No peel, no heat.
15. If `Plates <= 0` → end `"track"`. If `Heat >= HeatMax` → end `"cooked"`.
16. Decay shake, dollar lives.

**Never** delete rubble. Rubble is the maze.

Bounds: dozer center clamped to `[DozerBodyR, LotW-DozerBodyR] × [DozerBodyR, LotH-DozerBodyR]`.

---

## 9. Heat and armor

### Heat

```
if stalled:
    heat += HeatCookStall * dt
else if BladeDown and (blade overlapping solid or deep rubble):
    heat += HeatCookPush * dt
else if !BladeDown and not overlapping solid:
    heat -= HeatCoolAsphalt * dt
else:
    heat -= HeatCoolIdle * dt
clamp 0..100
```

“Deep rubble” = overlapping a `Rubble` whose area ≥ 16×16.

**Draw:** no numeric bar. Palette-swap the body toward red as `Heat/100`. At `Heat > 70` draw two 1-px vent dashes beside the hood, flickering on `Tick%12 < 6`. At `Heat > 90` whole body pulses.

### Armor (onion, 4 plates)

| Plates | Name | Body color |
|---|---|---|
| 4 | Paint | `#E0C040` |
| 3 | Primer | `#C9A14A` |
| 2 | Rust hull | `#8A5A32` |
| 1 | Black frame | `#2A2A28` |
| 0 | dead | — |

**Peel rules**

- Cruiser contact on **side or rear**, blade **up**, `IFrames==0`, `CruiserPIT==true` → `Plates--`, `IFrames = 45`, shake. Cruiser bounces.
- Same contact, `CruiserPIT==false` (sheriff down) → cruiser bounces, **no peel**.
- Cruiser contact on **front** + blade **down** → cruiser dies, `$25` vehicle, no peel.
- Excavator **boom overlap** → peel **any stance**, `IFrames = 60`.
- Chopper: never peels.
- Buildings: never peel.

Side/rear test:

```go
fx, fy := Forward(d.Heading)
dx, dy := cruiser.X-d.X, cruiser.Y-d.Y
along := dx*fx + dy*fy          // > 0 in front
if along > 4 {
    // front
} else {
    // side or rear
}
```

Last plate + one peel or boom → `Plates = 0` → thrown track.

---

## 10. Targets and boons

When a named building first reaches `Rubble`:

| ID | Boon (set flags, show a 1.2 s banner) |
|---|---|
| Sheriff | `CruiserPIT = false`. Banner: `PERMIT: PIT MANEUVER REVOKED` |
| Yard | `DumpTrucks = false`; despawn any alive dump blockers; `WallsBrittle = true` (jersey HP → 1). Banner: `PERMIT: PUBLIC WORKS CLOSED` |
| Plant | `ConcreteSets = false`; any unset concrete stays shoveable (treat as weak rubble HP 3, shoveable like a dump). Banner: `PERMIT: MIX NEVER SETS` |

`run.Targets = count of the three flags`. Ignore-targets is allowed: you can drive circles until the buzzer.

Chopper is **not** deletable. No fourth target.

---

## 11. Beat chart (code)

```go
func (r *Run) Time() float64 {
    return float64(r.Tick)/60.0 + r.PressureBonus
}
```

| `Time()` ≥ | Event (once) |
|---|---|
| 0 | 2 cruisers spawn south of sheriff, hunt |
| 40 | jersey blockers at drop points |
| 60 | chopper `Active=true` |
| 80 | excavator `Announced=true` (dust puff + `HEAVY EN ROUTE` banner at yard) |
| 105 | if `!YardDown` → excavator `Arrived=true` at corridor cell. If yard already rubble, **skip forever** |
| 135 | if `ConcreteSets` → concrete patches `Set=true` (solid). Else they stay shoveable |
| 165 | spawn +2 cruisers AND, if excavator alive, start boom swipes if not already; if excavator dead, spawn extra dump **only if** `DumpTrucks` |
| 210 | buzzer |

Pressure bonus: while dozer center is inside chopper spot radius, `PressureBonus += 0.35 * dt` (linger accelerates the town). Chopper never fires a weapon.

---

## 12. Threats

### Cruisers

Seek a point `dozer - Forward*22 + Right*side*16` (alternate side per id). Max speed 95. They **do not** occupy streets as blockers (you can overlap and flatten or peel). Light-bar: two 2-px pixels blinking.

Count: start 2, +1 at 40 s, +2 at 165 s. Cap 6 alive.

### Blockers

AABB solids. Dump trucks if `DumpTrucks`. Jersey walls: HP 8, or 1 if `WallsBrittle`. Blade-down wrecks them like buildings. Dead jersey becomes rubble.

### Excavator

Silhouette **must not** be the player dozer: longer body, boom rect sticking out. Does not chase. Boom overlap = peel. Blade-down **body ram**: excavator HP 40, you cook (`HeatCookPush`). If HP 0: `$200` vehicle, dead. If swing-side cell is rubble (`Rubble` overlapping boom origin + right*20), boom cannot fire (buried).

### Chopper

Sprite: 16×10 diamond + rotor line. Spotlight: translucent cone/circle, radius 72, drawn under the dozer. No rockets.

### Peds

Max 8. Spawn on shoulders. Garnish.

---

## 13. Collision (keep it dumb)

- Dozer **body**: circle `R=14`.
- Dozer **blade**: AABB in local space, transformed by heading (4 corners, then SAT vs AABB **or** convert blade to a world AABB using min/max of corners — the latter is good enough).
- Buildings, rubble, jersey, dumps, lot props: **AABB**.
- Circle vs AABB: closest-point clamp, push out along the vector, kill speed into the wall (`speed = min(speed, 0)` if moving into the normal).

Pseudo:

```go
func CircleAABB(cx, cy, r, x, y, w, h float64) (nx, ny, pen float64, hit bool) {
    closestX := clamp(cx, x, x+w)
    closestY := clamp(cy, y, y+h)
    dx, dy := cx-closestX, cy-closestY
    d2 := dx*dx + dy*dy
    if d2 >= r*r {
        return 0, 0, 0, false
    }
    d := math.Sqrt(d2)
    if d < 1e-6 {
        return 0, -1, r, true
    }
    return dx / d, dy / d, r - d, true
}
```

Resolve each solid, max 3 iterations. No physics engine.

---

## 14. Score

```
final = (Struct$ + Vehicle$ + int(TimeAlive)) * Mult[Targets]
```

`TimeAlive = min(run.Time(), 210)` on end.  
`Mult = {0: 1.0, 1: 1.25, 2: 1.6, 3: 2.0}`.

Dollar ticks during the run: `+$N` at the wreck, rising 20 px over 0.7 s, color `#F8E070`.

Tally screen (same for cooked / track / buzzer):

1. Freeze the last framebuffer **or** keep drawing the lot with sim paused (pause is easier: `SceneTally` still `Draw`s the world, no `Update` sim).
2. Overlay panel, tally **rolls** over ~1.4 s: STRUCTURE → VEHICLE → TIME → TARGETS ×MULT → TOTAL.
3. Death line in rust: `ENGINE COOKED` / `TRACK THROWN` / `COUNTY CLOCK`.
4. Prompt: `SPACE / TAP — AGAIN` (same lot, `NewRun()`, dozer at south gate, buildings reset).

Do not change the lot layout between runs.

---

## 15. Rendering

**Order (back to front)**

1. Dirt fill `#5A6B3A` / asphalt drag `#3A3A3E` with 16-px tile noise (hash, not random each frame)
2. Concrete patches (wet `#8A8A7A` or set `#B0B0A4`)
3. Building shadows (offset 2,2, `#0006`)
4. Buildings by state: intact `#6E7A84`, cracked darker + X crack lines, rubble `#4A4036` scatter rects **inside** the AABB
5. Named labels (8 px, `text/v2`)
6. Blockers, peds, cruisers
7. Excavator
8. Dozer (body + tracks + blade + vents)
9. Chopper + spotlight (spotlight multiply/alpha 0.25 before chopper sprite)
10. Dollar ticks, banners
11. HUD
12. Title / tally overlays

**HUD (high contrast, no HP number, no heat number)**

- Top-left: elapsed `1:23` in `#F2E6C4` (or countdown `1:07` — pick **elapsed**; beat chart is easier to feel)
- Top-center: `PERMIT DENIED` tiny, or nothing
- Top-right: `$` running `(Struct$+Vehicle$)` and `×1.6` if targets > 0
- Rim pips for remaining named targets
- Bottom: `BLADE UP` / `BLADE DOWN` in the stance color

Use `text/v2` with a bundled TTF (e.g. a small pixel font under `internal/render/fonts`). If you cannot embed a font on day one, `ebitenutil.DebugPrint` is acceptable **only** for milestone 1.

**Palette (Genesis-ish, 12 colors)**

```
#1B1B22 bg night-green-dark
#5A6B3A dirt
#3A3A3E asphalt
#E0C040 dozer paint
#C04040 heat / siren
#3A6FBF cruiser
#F2E6C4 HUD
#8A5A32 rust
#2A2A28 frame
#6E7A84 building
#F8E070 money
#4AD2C4 plant pip
```

---

## 16. Audio (milestone 6, not a blocker)

Two channels:

1. **Chase loop** — 8-bar square/triangle, ~140 BPM, generated PCM or a tiny WAV embed. Duck it 30% on tally. `M` toggles mute for this loop only (session-only; SFX stay audible).
2. **Crunch** — one-shot noise burst on wreck / peel.

`audio.NewContext(44100)`. Do not add voice or radio chatter.

---

## 17. Scenes

**Title:** black-green, big `PERMIT DENIED`, yellow dozer silhouette, `A/D STEER  W GO  S BACK  SPACE BLADE`, `PRESS SPACE`. No story crawl. Optional one-liner: `THE COUNTY SAID NO.`

**Play:** §8.

**Tally:** §14. Space / tap → `NewRun()`.

Game struct holds `Scene`, `Run`, `Dozer`, `Lot`, `FX`. `New()` starts at Title. First Space → Play.

---

## 18. Milestone order (stop at 5 if fun fails)

| M | Done when |
|---|---|
| **0** | Window 320×224 scaled, Title screen, Space to Play, yellow rect, camera follow, **control self-test §4.1 passes** |
| **1** | Lot asphalt + three named AABBs + mundane shacks drawn. Blade down wrecks Intact→Cracked→Rubble. Rubble collides. Reverse works. |
| **2** | Heat cook/cool + 4 palette plates. Die cooked. Die plates=0 (debug key `F1` peels, remove later). Tally + Again. |
| **3** | Score dollars, multiplier, freeze tally roll. Pips. Banners on targets. Boons actually change flags. |
| **4** | Cruisers + PIT rules. Blockers at 0:40. Concrete at 2:15. Chopper spotlight + pressure. |
| **5** | Excavator announce/arrive/skip-if-yard-dead. Boom peel. Two-family beat. Show-off tests. |
| **6** | Touch tiller. Audio. 16-facing atlas if you want. WASM. |

Do not start M4 until a human (or you, if you are the agent playing it) can fail M1 on purpose by boxing the dozer in its own rubble.

---

## 19. Show-off tests (definition of done)

Copy these into `README.md`. All must pass.

1. New player, no tooltip wall: steer, toggle blade, die inside 90 s.
2. A learned player can look at the plant pip and judge time.
3. Coward run (0 targets) and two-target run do **not** cash out the same (`×1.0` vs `×1.6`).
4. Heat readable from the sprite; plates readable from palette. No numeric bars required.
5. At least one street on the drag is stolen by player-made rubble in a typical run.

**Control tests (agent must run)**

- Spawn, W, 1 second → `dozer.Y < 1180 - 40`
- Spawn, A 0.5 s → heading in `(-π, 0)` or `(π, 2π)` equivalent west-of-north
- Blade down, push sheriff → HP drops
- `go test ./...` includes `TestForwardVector` and `TestMultTable`

```go
func TestMultTable(t *testing.T) {
    m := []float64{1.0, 1.25, 1.6, 2.0}
    for i, want := range m {
        if Mult(i) != want {
            t.Fatalf("targets %d: got %v want %v", i, Mult(i), want)
        }
    }
}
```

---

## 20. WASM (optional, after the lot is fun)

```bash
go install github.com/hajimehoshi/wasmserve@latest
wasmserve ./cmd/permitdenied
```

Or:

```bash
GOOS=js GOARCH=wasm go build -o permitdenied.wasm ./cmd/permitdenied
```

Serve with `wasm_exec.js` from the same Go version. Touch path in §4.2 must work or the phone build is a lie.

Do not add a React wrapper. This game is the binary.

---

## 21. Debug overlay (`F2` toggle, strip for show-off)

```
T=83.2 PIT=1 DUMP=1 SET=1 YARD=0
HEAT=42 PLATES=3 BLADE=DOWN SPD=31
```

Never on by default.

`F1` in `go run` builds: peel one plate. `F3`: skip +15 s on the clock. Both compile out or no-op behind `ebiten.IsKeyPressed` is fine for this prototype; do not build tags unless you want them.

---

## 22. Copy (use verbatim)

- Window / title: `PERMIT DENIED`
- Title one-liner: `THE COUNTY SAID NO.`
- Stance: `BLADE UP` / `BLADE DOWN`
- Banners: see §10
- Deaths: `ENGINE COOKED` / `TRACK THROWN` / `COUNTY CLOCK`
- Again: `SPACE / TAP — AGAIN`
- Do **not** use: Killdozer, Heemeyer, Granby, manifesto, wanted stars

Fiction note (not in-game): later lots *may* be a crusade about a shed permit. **This strip does not tell that story.**

---

## 23. `Update` skeleton (copy and fill)

```go
func (g *Game) Update() error {
    in := g.readInput()
    switch g.scene {
    case SceneTitle:
        if in.BladeToggle || ebiten.IsKeyJustPressed(ebiten.KeyEnter) {
            g.startRun()
        }
    case ScenePlay:
        if g.fx.HitStop > 0 {
            g.fx.HitStop--
            g.run.Tick++
            return nil
        }
        g.stepPlay(in)
    case SceneTally:
        g.fx.TallyT += Dt
        if in.BladeToggle || ebiten.IsKeyJustPressed(ebiten.KeyEnter) {
            g.startRun()
        }
    }
    return nil
}
```

`startRun` resets dozer, lot HP/states, threats, score, `scene=Play`. Same geometry every time.

---

## 24. What “simple” means in code

- One `Lot` slice of buildings. No quadtree until you measure.
- No goroutines. The sim is single-threaded in `Update`.
- No JSON level files. Geometry is Go literals in `lot/lot.go`.
- No shader for heat. `ColorScale.Scale`.
- If a threat does not change steering, delete it.

When the five show-off tests pass, **stop**.

---

## 25. Quick reference — files vs spec

| Spec idea | Code |
|---|---|
| Tank-steer | `Input.Steer` → heading clockwise on D |
| Blade stance | `Dozer.BladeDown` |
| Onion plates | `Dozer.Plates` 4..0 |
| Heat as sprite | `Dozer.Heat` + `ColorScale` |
| 3 building states | `BuildingState` |
| Rubble collides | `[]Rubble` never culled |
| Sheriff / yard / plant | `TargetID` + booleans on `Run` |
| Beat chart | `Run.Time()` vs constants |
| Chopper | spotlight + `PressureBonus` |
| Excavator | terrain event, skip if yard dead |
| Score | `(S+V+T)*Mult[targets]` |
| Same tally for all deaths | `SceneTally` |
| No meta | no save besides optional `localStorage` equivalent — **none** |

---

*End of implementation guide. If you need a new system to make the lot fun, you are probably tuning constants in `const.go` instead.*
