package game

import (
	"math"
	"testing"

	"permitdenied/internal/lot"
)

func TestDestroyBuildingSpawnsDollarTick(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	b := g.lot.BuildingByID(lot.TargetSheriff)
	if b == nil {
		t.Fatal("missing sheriff")
	}
	cx, cy := b.Center()
	want := b.Value
	g.destroyBuilding(b)
	if len(g.fx.Dollars) != 1 {
		t.Fatalf("dollars=%d want 1", len(g.fx.Dollars))
	}
	d := g.fx.Dollars[0]
	if d.Amt != want {
		t.Fatalf("amt=%d want %d", d.Amt, want)
	}
	if d.Life != DollarLife {
		t.Fatalf("life=%v want %v", d.Life, DollarLife)
	}
	if math.Abs(d.X-cx) > 0.01 || math.Abs(d.Y-cy) > 0.01 {
		t.Fatalf("dollar at %v,%v want center %v,%v", d.X, d.Y, cx, cy)
	}
}

func TestHitStopSkipsMotion(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	x0, y0 := g.dozer.X, g.dozer.Y
	tick0 := g.run.Tick
	g.fx.HitStop = HitStopTicks
	if err := g.Drive(Input{Throttle: 1}, Keys{}); err != nil {
		t.Fatal(err)
	}
	if g.dozer.X != x0 || g.dozer.Y != y0 {
		t.Fatalf("moved during hit-stop: %v,%v -> %v,%v", x0, y0, g.dozer.X, g.dozer.Y)
	}
	if g.run.Tick != tick0+1 {
		t.Fatalf("tick=%d want %d", g.run.Tick, tick0+1)
	}
	if g.fx.HitStop != HitStopTicks-1 {
		t.Fatalf("hit-stop=%d want %d", g.fx.HitStop, HitStopTicks-1)
	}
}
