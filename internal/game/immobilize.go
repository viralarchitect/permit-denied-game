package game

import (
	"permitdenied/internal/attach"
	"permitdenied/internal/lot"
)

func (g *Game) stepImmobilize(stalled, overlapping bool) {
	if stalled && overlapping && !g.dozer.BladeDown {
		g.stallTicks++
	} else {
		g.stallTicks = 0
	}
	if g.stallTicks >= int(PinSeconds*TPS) {
		g.endRun("pinned")
		return
	}

	deep := 0
	for i := range g.lot.Rubble {
		r := g.lot.Rubble[i]
		if r.Ramp || r.W*r.H < DeepRubbleA {
			continue
		}
		if _, _, _, hit := CircleAABB(g.dozer.X, g.dozer.Y, DozerBodyR, r.X, r.Y, r.W, r.H); hit {
			deep++
		}
	}
	if deep >= BuriedNeed {
		g.buryTicks++
	} else {
		g.buryTicks = 0
	}
	if g.buryTicks >= int(BuriedSeconds*TPS) {
		g.endRun("buried")
	}
}

func wreckMulFor(g *Game, b lot.Building) float64 {
	if g.scene == ScenePlay {
		return 1
	}
	return g.kit.WreckMul(b.Material)
}

func (g *Game) bladeWidth() float64 {
	if g.kit.Wide {
		return BladeW * WideBladeScale
	}
	return BladeW
}

func (g *Game) collectPickups() {
	for i := range g.pickups {
		p := &g.pickups[i]
		if !p.Alive {
			continue
		}
		if _, _, _, hit := CircleAABB(g.dozer.X, g.dozer.Y, DozerBodyR, p.X, p.Y, p.W, p.H); !hit {
			continue
		}
		p.Alive = false
		g.kit.Grant(p.Kind)
		if p.Kind == attach.ExtraPlate {
			g.dozer.Plates++
		}
		g.fx.SetBanner("PERMIT: "+p.Kind.Label()+" ISSUED", BannerLife)
		g.fx.SpawnDollar(p.X, p.Y, 0, DollarLife)
	}
}
