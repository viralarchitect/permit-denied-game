package game

import (
	"testing"

	"permitdenied/internal/threats"
)

func TestBeatCruisersAtTZero(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	if g.run.Tick != 0 {
		t.Fatalf("Tick=%d want 0", g.run.Tick)
	}
	g.fireBeats()
	alive := 0
	for _, c := range g.cruisers {
		if c.Alive {
			alive++
			if !countyPointWithinWorld(g, c.X, c.Y, CruiserRadius) {
				t.Fatalf("cruiser spawned out of bounds at %v,%v on %vx%v map", c.X, c.Y, g.worldW(), g.worldH())
			}
		}
	}
	if alive != 2 {
		t.Fatalf("living cruisers=%d want 2", alive)
	}
}

func TestBeatCountyBlockersSkipOffMapSpawns(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	before := len(g.blockers)
	g.run.PressureBonus = 0
	g.run.Tick = TBlockers * TPS
	g.fireBeats()
	if len(g.blockers) != before {
		t.Fatalf("blockers len=%d want %d; stale county blocker spawns should stay off the 320x288 pack map", len(g.blockers), before)
	}
	for _, b := range g.blockers {
		if !countyRectWithinWorld(g, b.X, b.Y, b.W, b.H) {
			t.Fatalf("off-map blocker present at (%v,%v) size %vx%v on %vx%v map", b.X, b.Y, b.W, b.H, g.worldW(), g.worldH())
		}
	}
}

func TestBeatExcavatorSkipsOffMapCountySlice(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	g.run.PressureBonus = 0
	g.run.Tick = TExAnnounce * TPS
	g.fireBeats()
	if g.excavator.Announced || g.fx.Banner == "HEAVY EN ROUTE" {
		t.Fatalf("off-map county excavator should not announce on pack slice: announced=%v banner=%q", g.excavator.Announced, g.fx.Banner)
	}
	g.run.Tick = TExArrive * TPS
	g.fireBeats()
	if g.excavator.Arrived || g.excavator.Alive {
		t.Fatalf("off-map county excavator should not spawn on pack slice: Arrived=%v Alive=%v", g.excavator.Arrived, g.excavator.Alive)
	}
}

func TestBeatConcreteSetFlag(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	g.run.PressureBonus = 0
	g.run.ConcreteSets = true
	g.run.Tick = TConcreteSet * TPS
	g.fireBeats()
	for _, b := range g.blockers {
		if b.Kind != threats.BlockerConcrete || !b.Alive {
			continue
		}
		if !b.Set {
			t.Fatalf("ConcreteSets=true: concrete at (%v,%v) Set=%v want true", b.X, b.Y, b.Set)
		}
	}

	g2 := New()
	g2.audio = nil
	g2.startRun()
	g2.run.PressureBonus = 0
	g2.run.ConcreteSets = false
	g2.run.Tick = TConcreteSet * TPS
	g2.fireBeats()
	for _, b := range g2.blockers {
		if b.Kind != threats.BlockerConcrete || !b.Alive {
			continue
		}
		if b.Set {
			t.Fatalf("ConcreteSets=false: concrete at (%v,%v) Set=%v want false", b.X, b.Y, b.Set)
		}
	}
}
