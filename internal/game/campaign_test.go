package game

import (
	"math"
	"testing"

	"permitdenied/internal/lot"
	"permitdenied/internal/threats"
)

func TestWreckBladeDownSpawnsRubble(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	b := poseOnWestShack(g)
	if b == nil {
		t.Fatal("missing shack")
	}
	before := len(g.lot.Rubble)
	hp0 := b.HP
	g.dozer.BladeDown = true
	for i := 0; i < 20*TPS; i++ {
		g.stepPlay(Input{Throttle: 1})
		if b.State == lot.InRubble {
			break
		}
	}
	if b.HP >= hp0 {
		t.Fatalf("HP did not drop: %v -> %v", hp0, b.HP)
	}
	if b.State != lot.InRubble {
		t.Fatalf("state=%v want InRubble", b.State)
	}
	if len(g.lot.Rubble) <= before {
		t.Fatalf("rubble %d -> %d, want spawn", before, len(g.lot.Rubble))
	}
}

func TestImmobilizeStalledOverlapPins(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	b := g.lot.BuildingByID(lot.TargetSheriff)
	if b == nil {
		t.Fatal("missing sheriff")
	}
	poseSouthOf(g, b, false)
	g.blockers = append(g.blockers, threats.Jersey(b.X, b.Y+b.H+2, b.W, 20, 99))
	for i := 0; i < int(PinSeconds*TPS)+5; i++ {
		g.stepImmobilize(true, true)
		if g.run.Over {
			break
		}
	}
	if !g.run.Over || g.run.Death != "pinned" {
		t.Fatalf("Over=%v Death=%q heat=%.1f want pinned", g.run.Over, g.run.Death, g.dozer.Heat)
	}
	if g.scene != SceneTally {
		t.Fatalf("scene=%v want tally", g.scene)
	}
}

func TestInputADecreasesHeading(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	h0 := g.dozer.Heading
	g.integrateDozer(Input{Steer: -1})
	if g.dozer.Heading >= h0 && h0 == 0 {
		if g.dozer.Heading == 0 {
			t.Fatal("A did not change heading")
		}
	}
	// heading 0, steer -1 → heading decreases, wraps near 2π
	if g.dozer.Heading < math.Pi {
		t.Fatalf("A from north: heading=%v want west-of-north (>π)", g.dozer.Heading)
	}
}

func TestCountyClearEntersTown(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	for _, id := range []lot.TargetID{lot.TargetSheriff, lot.TargetYard, lot.TargetPlant} {
		b := g.lot.BuildingByID(id)
		if b == nil {
			t.Fatalf("missing %v", id)
		}
		g.destroyBuilding(b)
	}
	if !g.mapCleared() {
		t.Fatal("county should be cleared")
	}
	g.stepPlay(Input{})
	if g.scene != SceneTown {
		t.Fatalf("scene=%v want town", g.scene)
	}
	if g.mapW != TownW || g.mapH != TownH {
		t.Fatalf("town size %vx%v", g.mapW, g.mapH)
	}
}

func TestTownClearCityCapitolCleared(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	g.enterTown()
	wreckNamed(g)
	g.stepPlay(Input{})
	if g.scene != SceneCity {
		t.Fatalf("after town: scene=%v want city", g.scene)
	}
	wreckNamed(g)
	g.stepPlay(Input{})
	if g.scene != SceneCapitol {
		t.Fatalf("after city: scene=%v want capitol", g.scene)
	}
	if g.tower.CoreIdx < 0 {
		t.Fatal("capitol missing core")
	}
	core := &g.lot.Buildings[g.tower.CoreIdx]
	g.destroyBuilding(core)
	g.stepPlay(Input{})
	if g.run.Death != "cleared" || g.scene != SceneTally {
		t.Fatalf("core drop: death=%q scene=%v want cleared/tally", g.run.Death, g.scene)
	}
}

func wreckNamed(g *Game) {
	for i := range g.lot.Buildings {
		b := &g.lot.Buildings[i]
		if b.Label == "" || b.Role == lot.RoleMundane {
			continue
		}
		if b.State != lot.InRubble {
			g.destroyBuilding(b)
		}
	}
}

func TestCapitolPenthouseReachWins(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	g.enterCapitol()
	ph := &g.lot.Buildings[g.tower.PenthouseIdx]
	cx, cy := ph.Center()
	g.dozer.X, g.dozer.Y = cx, cy
	g.tower.MarkReached(cx, cy, g.lot)
	g.destroyBuilding(ph)
	g.stepPlay(Input{})
	if g.run.Death != "cleared" {
		t.Fatalf("death=%q want cleared", g.run.Death)
	}
}
