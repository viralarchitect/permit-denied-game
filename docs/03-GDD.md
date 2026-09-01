# GDD (lean)

## Camera and control

- World-fixed north. Camera follows dozer, clamped to map.
- Tank steer: forward/back + rotate. Blade is the front.
- Logical 320×224. All new art is 16px-grid native.

## Core loop

Spawn → drive to a named target → push / rip / collapse it → rubble becomes terrain → threats spend themselves on you → next target → extract to the next tier or die.

Score is dollars of assessed damage. Score is garnish. Clearing targets is the run.

## Run structure

```
County (1 map)
  → Town (1 of a pool)
    → City (1 of a pool)
      → Capitol (fixed boss map)
```

Fail anywhere: run over. Keep meta unlocks. Lose per-run attachments.

## Tiers

| Tier | Map feel | Buildings | Threats | Win |
|---|---|---|---|---|
| County | One strip / lot | Wood, brick, yard, plant | Cruisers, jersey, excavator, garnish chopper | Listed structures down |
| Town | Small grid, 2–3 streets | Concrete, school, courthouse stand-in (generic) | Fire trucks, roadblocks, more cars | All marked civic buildings |
| City | Dense blocks, overpass | Steel frame, parking garage, bus barn | SWAT wagons, buses as blockers, useless air | District targets + overpass optional |
| Capitol | Tower campus | High-rise as stacked floors | Everything at once, still cannot delete you | Penthouse rubble or tower collapse |

## Buildings

Each building is an AABB stack with HP by material, a dollar value, and a collapse rule.

- Wood: shears fast, light rubble
- Brick: chunks, medium rubble
- Concrete: slow, heavy blocks that pin
- Steel: needs ripper / ball / driver; falls as frames you can climb

Rubble is solid. Late-game rubble is the ramp. That is how a top-down dozer reaches a penthouse: undermine, pile, drive up, punch the next floor rectangle.

## Threats

Threats waste time and create immobilization risk. They do not out-DPS the player.

- Cruiser: rams, bounces, calls friends
- Blocker: jersey, dump, wet→set concrete
- Heavy: excavator / truck that can contest the blade
- Air: presence and heat, not a boss
- Crowd: garnish only, no gameplay duty

Wanted level only changes *which* of the above spawn, not a new verb.

## Dozer stats (design, not final numbers)

- Mass, torque, turn rate, blade width, armor, heat cap
- Immobilize if: buried past axle, heat max, or pinned with no forward vector for N seconds

## Upgrades

**Per-run (find or buy mid-map, lost on death)**  
Blade width, ripper, wrecking ball, pile driver, extra plate, torque shot.

**Meta (kept between runs)**  
Engine tier, armor tier, one attachment unlocked for future runs, map-pool entries.

Upgrade rule: every unlock must change *how a building falls* or *how you get stuck*. Cosmetics wait.

## Capitol / penthouse

The tower is N floor AABBs stacked in world-Y bands or a side-stack the camera treats as “up.”  
Destroy the core supports → upper floors become rubble ramps → drive to the penthouse rect → collapse it.  
If the core dies first, the penthouse dies with the building. Both count as a win.

## UI

Title, run HUD (time, heat, targets left, $), result sheet that looks like a county assessment. No lore screens.

## Audio (later)

Engine loop, blade scrape, collapse stinger, radio chatter as noise not plot. Silence is allowed.
