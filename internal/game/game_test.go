package game

import (
	"math"
	"testing"

	"permitdenied/internal/lot"
)

func TestForwardVector(t *testing.T) {
	cases := []struct {
		h, x, y float64
	}{
		{0, 0, -1},
		{math.Pi / 2, 1, 0},
		{math.Pi, 0, 1},
		{3 * math.Pi / 2, -1, 0},
	}
	for _, c := range cases {
		fx, fy := Forward(c.h)
		if math.Abs(fx-c.x) > 1e-9 || math.Abs(fy-c.y) > 1e-9 {
			t.Fatalf("heading %v: got %v,%v want %v,%v", c.h, fx, fy, c.x, c.y)
		}
	}
}

func TestMultTable(t *testing.T) {
	m := []float64{1.0, 1.25, 1.6, 2.0}
	for i, want := range m {
		if Mult(i) != want {
			t.Fatalf("targets %d: got %v want %v", i, Mult(i), want)
		}
	}
}

func TestSpawnWOneSecond(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	for i := 0; i < TPS; i++ {
		g.stepPlay(Input{Throttle: 1})
	}
	if g.dozer.Y >= SpawnY-40 {
		t.Fatalf("W 1s: Y=%v want < %v (forward should be north / decreasing Y)", g.dozer.Y, SpawnY-40)
	}
}

func TestSpawnAHalfSecond(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	for i := 0; i < TPS/2; i++ {
		g.stepPlay(Input{Steer: -1})
	}
	h := g.dozer.Heading
	west := (h > math.Pi && h < 2*math.Pi) || (h < 0 && h > -math.Pi)
	if !west {
		t.Fatalf("A 0.5s: heading=%v want west-of-north in (-π,0) or (π,2π)", h)
	}
}

func TestBladeDownDropsSheriffHP(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	g.dozer.X = 224
	g.dozer.Y = 736
	g.dozer.Heading = 0
	g.dozer.BladeDown = true
	g.dozer.Speed = 0
	b := g.lot.BuildingByID(lot.TargetSheriff)
	if b == nil {
		t.Fatal("missing sheriff")
	}
	hp0 := b.HP
	for i := 0; i < 30; i++ {
		g.stepPlay(Input{})
	}
	if b.HP >= hp0 {
		t.Fatalf("blade-down vs sheriff: HP %v -> %v, expected drop", hp0, b.HP)
	}
}

func TestCowardVsTwoTargetScore(t *testing.T) {
	coward := 100 + 20 + 210
	two := int(float64(coward) * Mult(2))
	one0 := int(float64(coward) * Mult(0))
	if two == one0 {
		t.Fatalf("coward and two-target cashed out the same: %d", two)
	}
	if Mult(0) != 1.0 || Mult(2) != 1.6 {
		t.Fatalf("mult table drifted")
	}
}

func TestEndRunTallyDeathAndScore(t *testing.T) {
	deaths := []string{"cooked", "track", "buzzer"}
	for _, death := range deaths {
		g := New()
		g.Silence()
		g.startRun()
		g.endRun(death)
		if g.scene != SceneTally {
			t.Fatalf("death %s: scene=%v want SceneTally", death, g.scene)
		}
		if g.run.Death != death {
			t.Fatalf("death %s: got %q", death, g.run.Death)
		}
	}
	g := New()
	g.Silence()
	g.startRun()
	g.run.StructCash = 100
	g.run.VehicleCash = 20
	g.run.TimeAlive = 210
	zero := g.run.Final(Mult(0))
	two := g.run.Final(Mult(2))
	if zero == two {
		t.Fatalf("Final(Mult(0))=%d == Final(Mult(2))=%d", zero, two)
	}
}

func TestRightAtNorthIsEast(t *testing.T) {
	rx, ry := Right(0)
	if math.Abs(rx-1) > 1e-9 || math.Abs(ry) > 1e-9 {
		t.Fatalf("Right(0)=%v,%v want 1,0", rx, ry)
	}
}

func TestFacingIndexCardinals(t *testing.T) {
	if FacingIndex(0) != 0 {
		t.Fatalf("north: %d", FacingIndex(0))
	}
	if FacingIndex(math.Pi/2) != 4 {
		t.Fatalf("east: %d", FacingIndex(math.Pi/2))
	}
	if FacingIndex(math.Pi) != 8 {
		t.Fatalf("south: %d", FacingIndex(math.Pi))
	}
	if FacingIndex(3*math.Pi/2) != 12 {
		t.Fatalf("west: %d", FacingIndex(3*math.Pi/2))
	}
}
