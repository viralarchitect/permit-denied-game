package game

import (
	"math"

	"permitdenied/internal/dozerpack"
	"permitdenied/internal/threats"
)

// beat.go — calendar / beat chart.
// Reads boon flags (DumpTrucks, YardDown, ConcreteSets, WallsBrittle, and
// PressureBonus via run.Time()). Writes spawn slices + g.beat once-flags.
// May banner "HEAVY EN ROUTE". Does not wreck, peel, heat, spawn dollar/hit-stop,
// or resolve collision.

func (g *Game) fireBeats() {
	t := g.run.Time()
	if !g.beat.cruisers0 && t >= 0 {
		g.beat.cruisers0 = true
		g.spawnCruisers(2)
	}
	if !g.beat.blockers && t >= TBlockers {
		g.beat.blockers = true
		g.blockers = append(g.blockers, inBoundsCountyBlockers(g,
			threats.Jersey(248, 1000, 48, 16, jerseyHPNow(g.run.WallsBrittle)),
			threats.Jersey(344, 860, 48, 16, jerseyHPNow(g.run.WallsBrittle)),
			threats.Jersey(248, 480, 56, 16, jerseyHPNow(g.run.WallsBrittle)),
		)...)
		if g.run.DumpTrucks {
			if dump, ok := countyBlockerIfInBounds(g, threats.Dump(300, 720, 40, 24, DumpHP)); ok {
				g.blockers = append(g.blockers, dump)
			}
		}
		g.spawnCruisers(1) // §12: +1 at 40 s
	}
	if !g.beat.chopper && t >= TChopper {
		g.beat.chopper = true
		g.chopper.Active = true
		g.chopper.SpotR = ChopperSpotR
		if g.chopper.X == 0 && g.chopper.Y == 0 {
			g.chopper.X, g.chopper.Y = 320, 40
		}
	}
	if !g.beat.exAnn && t >= TExAnnounce {
		g.beat.exAnn = true
		if hasCountyExcavatorSlot(g) {
			g.excavator.Announced = true
			g.fx.SetBanner("HEAVY EN ROUTE", BannerLife)
		}
	}
	if !g.beat.exArr && t >= TExArrive {
		g.beat.exArr = true
		if !g.run.YardDown {
			g.placeExcavator()
		}
		// yard already rubble → skip forever
	}
	if !g.beat.concrete && t >= TConcreteSet {
		g.beat.concrete = true
		for i := range g.blockers {
			b := &g.blockers[i]
			if b.Kind != threats.BlockerConcrete || !b.Alive {
				continue
			}
			applyConcreteLaw(b, g.run.ConcreteSets)
		}
	}
	if !g.beat.twoFam && t >= TTwoFamilies {
		g.beat.twoFam = true
		g.spawnCruisers(2)
		if g.excavator.Alive && g.excavator.Arrived {
			g.excavator.Swinging = true
		}
		if !g.excavator.Alive && g.run.DumpTrucks {
			if dump, ok := countyBlockerIfInBounds(g, threats.Dump(310, 600, 40, 24, DumpHP)); ok {
				g.blockers = append(g.blockers, dump)
			}
		}
	}
}

func jerseyHPNow(brittle bool) float64 {
	if brittle {
		return JerseyBrittleHP
	}
	return JerseyHP
}

func (g *Game) placeExcavator() bool {
	x, y := 288.0, 520.0
	if !countyRectWithinWorld(g, x-10, y-18, ExcavatorBodyW, ExcavatorBodyH) {
		x, y = 288, 360
	}
	if !countyRectWithinWorld(g, x-10, y-18, ExcavatorBodyW, ExcavatorBodyH) {
		return false
	}
	if g.lot.RubbleBlocks(x-16, y-16, 32, 32) {
		x, y = 288, 360
	}
	if !countyRectWithinWorld(g, x-10, y-18, ExcavatorBodyW, ExcavatorBodyH) {
		return false
	}
	g.excavator = threats.Excavator{
		X: x, Y: y,
		Heading:   math.Pi, // south
		Announced: true,
		Arrived:   true,
		Alive:     true,
		HP:        ExcavatorHP,
		Swinging:  true,
	}
	return true
}

func (g *Game) spawnCruisers(n int) {
	spots := [][2]float64{
		{280, 760}, {360, 800}, {320, 900},
		{260, 700}, {380, 700}, {300, 850},
	}
	if g.scene == ScenePlay && len(g.countyCruiserSpots) > 0 {
		spots = make([][2]float64, 0, len(g.countyCruiserSpots))
		for _, s := range g.countyCruiserSpots {
			spots = append(spots, [2]float64{s.X, s.Y})
		}
	}
	alive := 0
	for _, c := range g.cruisers {
		if c.Alive {
			alive++
		}
	}
	i := 0
	for n > 0 && alive < CruiserCap && len(spots) > 0 {
		s := spots[i%len(spots)]
		off := float64(i / len(spots) * 14)
		x, y := s[0]+off, s[1]+off
		if !countyPointWithinWorld(g, x, y, CruiserRadius) {
			i++
			if i > len(spots)*2 {
				break
			}
			continue
		}
		g.cruisers = append(g.cruisers, threats.SpawnCruiser(x, y))
		n--
		alive++
		i++
	}
}

func blockersFromPack(specs []dozerpack.BlockerSpec) []threats.Blocker {
	out := make([]threats.Blocker, 0, len(specs))
	for _, spec := range specs {
		switch spec.Kind {
		case dozerpack.BlockerJersey:
			out = append(out, threats.Jersey(spec.X, spec.Y, spec.W, spec.H, JerseyHP))
		case dozerpack.BlockerDump:
			out = append(out, threats.Dump(spec.X, spec.Y, spec.W, spec.H, DumpHP))
		case dozerpack.BlockerConcrete:
			out = append(out, threats.Concrete(spec.X, spec.Y, spec.W, spec.H))
		}
	}
	return out
}

func inBoundsCountyBlockers(g *Game, blockers ...threats.Blocker) []threats.Blocker {
	out := make([]threats.Blocker, 0, len(blockers))
	for _, blocker := range blockers {
		if inBounds, ok := countyBlockerIfInBounds(g, blocker); ok {
			out = append(out, inBounds)
		}
	}
	return out
}

func countyBlockerIfInBounds(g *Game, blocker threats.Blocker) (threats.Blocker, bool) {
	if !countyRectWithinWorld(g, blocker.X, blocker.Y, blocker.W, blocker.H) {
		return threats.Blocker{}, false
	}
	return blocker, true
}

func hasCountyExcavatorSlot(g *Game) bool {
	return countyRectWithinWorld(g, 288-10, 520-18, ExcavatorBodyW, ExcavatorBodyH) ||
		countyRectWithinWorld(g, 288-10, 360-18, ExcavatorBodyW, ExcavatorBodyH)
}

func countyPointWithinWorld(g *Game, x, y, r float64) bool {
	return x >= r && y >= r && x <= g.worldW()-r && y <= g.worldH()-r
}

func countyRectWithinWorld(g *Game, x, y, w, h float64) bool {
	return x >= 0 && y >= 0 && x+w <= g.worldW() && y+h <= g.worldH()
}

func initialBlockers(dumpTrucks bool) []threats.Blocker {
	b := []threats.Blocker{
		threats.Jersey(400, 780, 48, 16, JerseyHP), // yard mouth choke
		threats.Concrete(256, 1100, 80, 24),
		threats.Concrete(280, 400, 80, 24),
		threats.Concrete(260, 200, 96, 32), // plant approach
	}
	if dumpTrucks {
		b = append(b,
			threats.Dump(456, 840, 40, 24, DumpHP),
			threats.Dump(520, 840, 40, 24, DumpHP),
		)
	}
	return b
}
