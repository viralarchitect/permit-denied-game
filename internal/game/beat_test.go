package game

import (
	"testing"

	"permitdenied/internal/lot"
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
		}
	}
	if alive != 2 {
		t.Fatalf("living cruisers=%d want 2", alive)
	}
}

func livingDumpAt(blockers []threats.Blocker, x, y float64) bool {
	for _, b := range blockers {
		if b.Kind == threats.BlockerDump && b.Alive && b.X == x && b.Y == y {
			return true
		}
	}
	return false
}

func TestBeatBlockersDumpRespectsFlags(t *testing.T) {
	// DumpTrucks=true → beat dump at (300, 720)
	g := New()
	g.audio = nil
	g.startRun()
	g.run.PressureBonus = 0
	g.run.DumpTrucks = true
	g.run.Tick = TBlockers * TPS
	g.fireBeats()
	if !livingDumpAt(g.blockers, 300, 720) {
		t.Fatal("DumpTrucks=true: expected living dump at (300, 720)")
	}

	// DumpTrucks=false → no beat dump at (300, 720)
	g2 := New()
	g2.audio = nil
	g2.startRun()
	g2.run.PressureBonus = 0
	g2.run.DumpTrucks = false
	g2.run.Tick = TBlockers * TPS
	g2.fireBeats()
	if livingDumpAt(g2.blockers, 300, 720) {
		t.Fatal("DumpTrucks=false: unexpected living dump at (300, 720)")
	}

	// Yard destroyed → DumpTrucks=false → no new living dump from blockers beat
	g3 := New()
	g3.audio = nil
	g3.startRun()
	g3.run.PressureBonus = 0
	yard := g3.lot.BuildingByID(lot.TargetYard)
	if yard == nil {
		t.Fatal("missing yard")
	}
	g3.destroyBuilding(yard)
	if g3.run.DumpTrucks {
		t.Fatal("after yard destroy DumpTrucks should be false")
	}
	g3.run.Tick = TBlockers * TPS
	g3.fireBeats()
	for _, b := range g3.blockers {
		if b.Kind == threats.BlockerDump && b.Alive {
			t.Fatalf("YardDown path: unexpected living dump at (%v,%v)", b.X, b.Y)
		}
	}
}

func TestBeatExcavatorArriveRespectsYardDown(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	g.run.PressureBonus = 0
	g.run.YardDown = false
	g.run.Tick = TExArrive * TPS
	g.fireBeats()
	if !g.excavator.Arrived || !g.excavator.Alive {
		t.Fatalf("YardDown=false: Arrived=%v Alive=%v want both true", g.excavator.Arrived, g.excavator.Alive)
	}

	g2 := New()
	g2.audio = nil
	g2.startRun()
	g2.run.PressureBonus = 0
	g2.run.YardDown = true
	g2.run.Tick = TExArrive * TPS
	g2.fireBeats()
	if g2.excavator.Arrived || g2.excavator.Alive {
		t.Fatalf("YardDown=true: Arrived=%v Alive=%v want not placed", g2.excavator.Arrived, g2.excavator.Alive)
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
