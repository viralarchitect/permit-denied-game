package game

import (
	"permitdenied/internal/lot"
	"permitdenied/internal/threats"
)

// combat.go — contact / combat.
// Writes boon flags and world solidity when a target dies (dumps despawn,
// jersey HP clamp, concrete HP/solid). Beat reads those flags; combat writes them.

func (g *Game) wreckWithBlade(bx, by, bw, bh float64) {
	for i := range g.lot.Buildings {
		b := &g.lot.Buildings[i]
		if b.State == lot.InRubble {
			continue
		}
		if !aabbOverlap(bx, by, bw, bh, b.X, b.Y, b.W, b.H) {
			continue
		}
		if b.ApplyDamage(WreckRateDown * Dt) {
			g.destroyBuilding(b)
		}
	}
	for i := range g.blockers {
		bl := &g.blockers[i]
		if !bl.Alive || !bl.Solid {
			continue
		}
		if !aabbOverlap(bx, by, bw, bh, bl.X, bl.Y, bl.W, bl.H) {
			continue
		}
		bl.HP -= WreckRateDown * Dt
		if bl.HP <= 0 {
			g.killBlocker(bl)
		}
	}
}

func (g *Game) glanceBuildings() {
	glanced := false
	for i := range g.lot.Buildings {
		b := &g.lot.Buildings[i]
		if b.State == lot.InRubble {
			continue
		}
		if _, _, _, hit := CircleAABB(g.dozer.X, g.dozer.Y, DozerBodyR, b.X, b.Y, b.W, b.H); hit {
			if !glanced {
				g.dozer.Speed *= GlanceScale
				glanced = true
			}
			if b.ApplyDamage(WreckRateUp * Dt) {
				g.destroyBuilding(b)
			}
		}
	}
}

func (g *Game) destroyBuilding(b *lot.Building) {
	g.lot.AddRubble(b.X, b.Y, b.W, b.H, RubbleInset)
	g.run.StructCash += b.Value
	cx, cy := b.Center()
	g.fx.SpawnDollar(cx, cy, b.Value, DollarLife)
	g.fx.HitStop = HitStopTicks
	g.fx.Shake += ShakeWreck
	g.audio.Crunch()
	if b.ID != lot.TargetNone {
		g.fireBoon(b.ID)
	}
}

func (g *Game) killBlocker(bl *threats.Blocker) {
	bl.Alive = false
	bl.Solid = false
	g.lot.AddRubble(bl.X, bl.Y, bl.W, bl.H, RubbleInset)
	g.fx.HitStop = HitStopTicks
	g.fx.Shake += ShakeWreck * 0.6
	g.audio.Crunch()
	if bl.Kind == threats.BlockerDump {
		g.run.VehicleCash += DumpCash
		g.fx.SpawnDollar(bl.X+bl.W/2, bl.Y+bl.H/2, DumpCash, DollarLife)
	}
}

func (g *Game) fireBoon(id lot.TargetID) {
	switch id {
	case lot.TargetSheriff:
		if g.run.SheriffDown {
			return
		}
		g.run.SheriffDown = true
		g.run.CruiserPIT = false
		g.fx.SetBanner("PERMIT: PIT MANEUVER REVOKED", BannerLife)
	case lot.TargetYard:
		if g.run.YardDown {
			return
		}
		g.run.YardDown = true
		g.run.DumpTrucks = false
		g.run.WallsBrittle = true
		for i := range g.blockers {
			b := &g.blockers[i]
			if b.Kind == threats.BlockerDump && b.Alive {
				b.Alive = false
				b.Solid = false
			}
			if b.Kind == threats.BlockerJersey && b.Alive {
				if b.HP > JerseyBrittleHP {
					b.HP = JerseyBrittleHP
				}
			}
		}
		g.fx.SetBanner("PERMIT: PUBLIC WORKS CLOSED", BannerLife)
	case lot.TargetPlant:
		if g.run.PlantDown {
			return
		}
		g.run.PlantDown = true
		g.run.ConcreteSets = false
		for i := range g.blockers {
			b := &g.blockers[i]
			if b.Kind == threats.BlockerConcrete && !b.Set {
				applyConcreteLaw(b, false)
			}
		}
		g.fx.SetBanner("PERMIT: MIX NEVER SETS", BannerLife)
	}
	g.run.RecountTargets()
}

// applyConcreteLaw stamps Set/Solid/HP for a concrete blocker.
// Wet spawn is Solid:false (threats.Concrete). Both writers currently force
// Solid:true and stamp HP. Spec wants unset to stay shoveable. Do not fix here.
func applyConcreteLaw(b *threats.Blocker, sets bool) {
	if sets {
		b.Set = true
		b.Solid = true
		b.HP = ConcreteSetHP
	} else {
		b.Set = false
		b.Solid = true
		b.HP = ConcreteUnsetHP
	}
}

func (g *Game) stepHeat(stalled, overlapping, bladeHitSolid, deepRubble bool) {
	h := g.dozer.Heat
	switch {
	case stalled:
		h += HeatCookStall * Dt
	case g.dozer.BladeDown && (bladeHitSolid || deepRubble):
		h += HeatCookPush * Dt
	case !g.dozer.BladeDown && !overlapping:
		h -= HeatCoolAsphalt * Dt
	default:
		h -= HeatCoolIdle * Dt
	}
	g.dozer.Heat = clamp(h, 0, HeatMax)
}

func (g *Game) peel(_ string) {
	if g.dozer.Plates <= 0 {
		return
	}
	g.dozer.Plates--
	g.dozer.IFrames = IFramesPeel
	g.fx.Shake += ShakePeel
	g.audio.Crunch()
}

// checkDeath ends the run on thrown track or cooked engine.
// Order: plates then heat. Returns true if the run ended.
func (g *Game) checkDeath() bool {
	if g.dozer.Plates <= 0 {
		g.endRun("track")
		return true
	}
	if g.dozer.Heat >= HeatMax {
		g.endRun("cooked")
		return true
	}
	return false
}
