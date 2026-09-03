# GTA-style engine roadmap

**Source:** [issue #18](https://github.com/viralarchitect/permit-denied-game/issues/18)  
**Product:** a top-down GTA-style engine that loads custom maps, vehicles, destructibles, and missions.  
**First pack:** dozer demolition. Not the ceiling. The testbed for the systems that make the clone worth building.

Existing PERMIT DENIED code is a parts bin. Its design docs are not binding.

---

## What "GTA clone" means here

Top-down. File-authored world. Vehicles as data. Buildings you can flatten. Cops that escalate. Missions that are files.

| In | Out of v1 |
|---|---|
| Roamable map from JSON | 3D, interiors, on-foot |
| Multiple vehicle chassis from one spec | Soft-body / Teardown physics |
| Destructibles that rewrite navigation | Full city / ambient traffic sim |
| Wanted from crime (damage, hitting cops) | Star HUD as a religion |
| Police: patrol, pursue, ram, blockade, pin | Lua/Starlark before JSON missions work |
| Mission file: win, lose, triggers, spawns | Third-party zip mods before pack 1 plays |
| Pack loads without a Go edit | |

Honest analog: GTA 2 + destructible sandbox. Not GTA V.

---

## Vertical slice — Dozer Pack

You drive a bulldozer through a small authored town-lot. You smash things. The world clogs with rubble. Cops show up harder the more you wreck. You win the mission or the machine/cops stop you.

This pack exists to prove the engine, not to be the whole game.

### Must prove

1. **Map is a file.** Player can ignore the objective, roam, and still crush + generate response.
2. **Vehicle is a file.** Dozer: heavy, slow, high crush, heat-on-load. At least one other chassis in the world (cruiser) from the same spec.
3. **Destruction rewrites the map.** Intact → damaged → rubble that blocks. Crush is the verb.
4. **Wanted ≠ engine heat.** Property $ and cop hits raise wanted. Wanted changes spawn count and AI aggression.
5. **Police play like GTA, not a beat chart.** Spawn, pursue, ram, blockade, pin. Bust = pinned N seconds. Dozer flattens cars. Cars cannot DPS-kill the dozer.
6. **Mission is a file.** Win: dollar quota and/or named structures. Lose: overheat, armor gone, busted. Triggers: `on_destroy`, `on_area_enter`, `timer_tick`.
7. **Load the pack and play it without editing Go.**

Fail the slice if it is still "sheriff / yard / plant then buzzer on a hardcoded lot."

### Keep out of the slice

Zip mods, hot reload, scripting VM, flowfield, huge city, player vehicle switching, on-foot, soft-body, ambient traffic schedules, multi-map campaign.

### Steal from current code

Ebitengine loop, tank-steer, 3-state buildings, immobilize, threat types, `assets/usable/lot.json`, `Game.Drive` / `Snapshot`, stamp-asset atlas. Throw away beat-chart-as-the-game.

---

## Workstreams

Build these together. The dozer pack is the integration test for all of them.

### 1. Contract / schema

JSON. One runtime map format. Tiled may author; runtime does not speak two dialects.

```
packs/dozer/
  pack.json
  map.json
  mission.json
  vehicles/dozer.json
  vehicles/cruiser.json
  destructibles/*.json
  sprites/ ...
```

- **Map:** layers `ground` / `decal`. Objects `spawn`, `building`, `blocker`, `marker`, `trigger`.
- **Vehicle:** mass, top speed, turn rate, armor, engine_heat_rate, crush.
- **Destructible:** health, yield_dollars, states (start with 3), rubble rule.
- **Mission:** objectives, fail, triggers, spawn tables by wanted level.

Start from `lot.json` + `const.go` + `immobilize` fields. Add fields the clone needs (wanted, police states, $ quota). Do not preserve old omissions.

### 2. Vehicle + physics

One solver, many chassis.

- Dozer and cruiser both instantiate the vehicle spec.
- Arcade top-down: accel, momentum, surface friction, mass-based ram damage.
- Dozer extras (blade, heat-on-load) may stay dozer-only components attached via the spec (`role: "dozer"` or a parts list). A fully generic vehicle class is the Pack 2 exam, not a slice blocker.
- Do not add resolv unless the current circle-AABB is the measured bottleneck.

### 3. Destruction

- Dynamic swap: whole building → damage states → rubble obstacle.
- Debris is optional garnish. Collision rubble is required.
- Yield $ feeds wanted + score.

### 4. Mission runner

JSON state machine. No VM in v1.

- `on_destroy` / `on_area_enter` / `timer_tick` / `on_pin`
- Win / lose from the mission file
- Spawn rules keyed to wanted

### 5. Police / wanted

- Wanted rises from property $ and hitting cops. Falls on quiet time if you want that later; not required in the slice.
- States: patrol, pursuit (intercept/ram), blockade, bust (pin timer).
- Pathfinding: grid A* when seek-offset fails on the slice map. Flowfield later.

### 6. Tools

- `cmd/validate-pack` — missing textures, bad IDs, schema errors
- Headless runner on the existing harness (no window)
- Hot reload and zip packs after the slice is playable from a folder

---

## Phases

Interleave engine and pack. Do not build a dark engine for four phases and "add Killdozer."

### P0 — Hardcoded wreck window (days, disposable)

Steal current dozer + smash + rubble. Flatten a block. Get stuck in your pile. Lock feel. This is not the product. If this window is still the boot path after P2, the slice failed.

**Exit:** you can crush and bury yourself in a window.

### P1 — Contract + map file

Freeze the four schemas. Load `packs/dozer/map.json`. Drive, roam, smash two buildings. Rubble blocks.

**Exit:** no `lot.go` literals required to play the slice map. Deleting a building from JSON changes the lot.

### P2 — Two chassis + crush math

Dozer + cruiser from vehicle JSON. Mass-based ram. Engine heat on sustained push. Pin detection.

**Exit:** a parked cruiser and a brick office feel different to hit. Heat cooks if you keep pushing.

### P3 — Mission file

Wire objectives, fail, triggers. Default slice mission:

- Win: hit a $ quota (start at something readable, e.g. mid-six / low-seven figures of yield — tune later; #18's $10M is a pack number)
- Lose: engine cooked, armor gone, or pinned ≥ N seconds (start N = 3)

**Exit:** editing `mission.json` changes win/lose without a rebuild.

### P4 — Wanted + police states

Damage-to-wanted. Patrol → pursuit → blockade → bust. Spawn tables in the mission file.

**Exit:** a coward who taps one shed gets a cruiser. A player who flattens a block gets a pile-on and a pin attempt.

### P5 — Slice lock

Art, juice, readable heat/armor, assessment sheet or equivalent. Playtest.

**Exit (slice ships when):**

1. New player can steer, crush, draw cops, and die or cash out in one sitting with no tooltip wall.
2. Ignoring the objective still produces a wanted response.
3. Rubble you made has already stolen a street you needed.
4. A second person can change the $ quota and a spawn table in JSON and feel it in-game.
5. `validate-pack` rejects a broken trigger ID and a missing sprite.

### P6 — Engine beyond the dozer

Only after P5.

- Second player vehicle spec instance (car) on the same map, or a second pack that is not a dozer
- More mission types (survive N minutes, destroy listed targets, escape after quota)
- Ambient traffic
- A* upgrade / flowfield if the map grew
- Hot reload
- Zip pack loader
- Scripting VM only if JSON triggers cannot express the mission

---

## Child issues

File under #18. Change form.

**Contract**
- Freeze map / vehicle / destructible / mission JSON schemas
- Create `packs/dozer/` layout

**Slice map + wreck**
- Load dozer pack map from JSON (no lot.go dependency)
- Building damage states + blocking rubble from destructible spec

**Vehicles**
- Vehicle spec drives dozer
- Cruiser instantiates the same spec
- Engine heat, crush, pin from spec fields

**Mission**
- Mission runner: on_destroy, on_area_enter, timer_tick, on_pin
- Dozer pack win/lose from mission.json

**Police**
- Wanted from property $ and cop hits
- Police states: patrol, pursuit, blockade, bust

**Tools**
- `cmd/validate-pack`
- Headless pack runner (extend `Game.Drive` / `Snapshot`)

**After slice**
- Second player vehicle pack or car chassis
- Hot-reload watcher
- Zip pack loader
- Pathfinding upgrade (gated)

---

## First week

1. Write the four schemas from current types plus the fields the clone needs (`wanted`, police states, $ quota, crush).
2. `packs/dozer/map.json` loads; you can drive and flatten two buildings.
3. One cruiser exists from `vehicles/cruiser.json`.
4. One validator test fails on a missing sprite.

No Lua. No zip. No second map. No campaign wrapper.

## Stack

Keep Go + Ebitengine. Custom collision stays until it is the bottleneck. New packages as needed (`internal/pack`, `internal/mission`, `internal/wanted`). Old `beat.go` / `campaign.go` are not the architecture.
