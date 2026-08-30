package game

import (
	"testing"

	"permitdenied/internal/lot"
	"permitdenied/internal/threats"
)

func TestDestroyBuildingSpawnsBoom(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	b := g.lot.BuildingByID(lot.TargetSheriff)
	g.destroyBuilding(b)
	if len(g.fx.Bursts) != 1 {
		t.Fatalf("bursts=%+v", g.fx.Bursts)
	}
	if g.fx.Bursts[0].FrameName() != "boom_00" {
		t.Fatalf("frame %s", g.fx.Bursts[0].FrameName())
	}
}

func TestGlanceSparksOnce(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	g.dozer.X = 224
	g.dozer.Y = 700
	g.dozer.BladeDown = false
	g.glanceBuildings()
	if len(g.fx.Bursts) != 1 {
		t.Fatalf("first glance sparks: %d", len(g.fx.Bursts))
	}
	g.glanceBuildings()
	if len(g.fx.Bursts) != 1 {
		t.Fatalf("second glance retriggered: %d", len(g.fx.Bursts))
	}
}

func TestJerseyChipSparksOnce(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	g.blockers = []threats.Blocker{threats.Jersey(208, 704, 48, 16, 8)}
	g.dozer.X = 224
	g.dozer.Y = 730
	g.dozer.Heading = 0
	g.dozer.BladeDown = true
	bx, by, bw, bh := bladeAABB(g.dozer.X, g.dozer.Y, g.dozer.Heading, true)
	g.wreckWithBlade(bx, by, bw, bh)
	if len(g.fx.Bursts) != 1 {
		t.Fatalf("jersey chip sparks: %d", len(g.fx.Bursts))
	}
	g.wreckWithBlade(bx, by, bw, bh)
	if len(g.fx.Bursts) != 1 {
		t.Fatalf("jersey retriggered: %d", len(g.fx.Bursts))
	}
}

func TestCruiserFlattenSpawnsBoom(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	g.dozer.BladeDown = true
	g.dozer.Heading = 0
	g.cruisers = g.cruisers[:0]
	g.cruisers = append(g.cruisers, threats.SpawnCruiser(g.dozer.X, g.dozer.Y-10))
	g.stepCruisers()
	if len(g.fx.Bursts) != 1 {
		t.Fatalf("cruiser flatten bursts=%d", len(g.fx.Bursts))
	}
}

func TestLastPeelSpawnsBoom(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	g.dozer.Plates = 1
	g.peel("test")
	if len(g.fx.Bursts) != 1 {
		t.Fatalf("last peel bursts=%d", len(g.fx.Bursts))
	}
}
