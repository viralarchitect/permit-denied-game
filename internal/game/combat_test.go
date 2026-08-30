package game

import (
	"testing"

	"permitdenied/internal/lot"
	"permitdenied/internal/threats"
)

func TestDestroySheriffRevokesPIT(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	b := g.lot.BuildingByID(lot.TargetSheriff)
	if b == nil {
		t.Fatal("missing sheriff")
	}
	g.destroyBuilding(b)
	if g.run.CruiserPIT {
		t.Fatal("CruiserPIT still true after sheriff destroy")
	}
	if !g.run.SheriffDown {
		t.Fatal("SheriffDown false after sheriff destroy")
	}
	if g.run.Targets != 1 {
		t.Fatalf("Targets=%d want 1", g.run.Targets)
	}
	// Idempotent boon: second call must not change the three asserts.
	g.destroyBuilding(b)
	if g.run.CruiserPIT || !g.run.SheriffDown || g.run.Targets != 1 {
		t.Fatalf("second destroy: PIT=%v SheriffDown=%v Targets=%d", g.run.CruiserPIT, g.run.SheriffDown, g.run.Targets)
	}
}

func TestDestroyYardDespawnsDumps(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	living := 0
	for _, b := range g.blockers {
		if b.Kind == threats.BlockerDump && b.Alive {
			living++
		}
	}
	if living < 2 {
		t.Fatalf("startRun living dumps=%d want >=2", living)
	}
	yard := g.lot.BuildingByID(lot.TargetYard)
	if yard == nil {
		t.Fatal("missing yard")
	}
	g.destroyBuilding(yard)
	if g.run.DumpTrucks {
		t.Fatal("DumpTrucks still true after yard destroy")
	}
	for _, b := range g.blockers {
		if b.Kind == threats.BlockerDump && b.Alive {
			t.Fatalf("living dump at (%v,%v) after yard destroy", b.X, b.Y)
		}
	}
}

func TestPeelFourEndsTrack(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	if g.dozer.Plates != 4 {
		t.Fatalf("Plates=%d want 4", g.dozer.Plates)
	}
	for i := 0; i < 4; i++ {
		g.peel("test")
	}
	if g.dozer.Plates != 0 {
		t.Fatalf("after 4 peels Plates=%d want 0", g.dozer.Plates)
	}
	if !g.checkDeath() {
		t.Fatal("checkDeath returned false, want true")
	}
	if !g.run.Over || g.run.Death != "track" {
		t.Fatalf("Over=%v Death=%q want Over=true Death=track", g.run.Over, g.run.Death)
	}
}

func TestSheriffDownNoCruiserPeel(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	sheriff := g.lot.BuildingByID(lot.TargetSheriff)
	if sheriff == nil {
		t.Fatal("missing sheriff")
	}
	g.destroyBuilding(sheriff)
	if g.run.CruiserPIT {
		t.Fatal("CruiserPIT still true")
	}
	g.dozer.X = SpawnX
	g.dozer.Y = SpawnY
	g.dozer.Heading = 0
	g.dozer.BladeDown = false
	g.dozer.IFrames = 0
	g.dozer.Speed = 0
	plates := g.dozer.Plates
	g.cruisers = []threats.Cruiser{
		{X: SpawnX + 16, Y: SpawnY, Alive: true},
	}
	g.stepCruisers()
	if g.dozer.Plates != plates {
		t.Fatalf("plates %d -> %d after side overlap with PIT off; want no peel", plates, g.dozer.Plates)
	}
}
