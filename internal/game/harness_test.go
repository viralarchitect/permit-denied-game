package game

import (
	"math"
	"testing"
)

func TestHarnessTitleSpaceStartsRun(t *testing.T) {
	g := New()
	g.Silence()
	if g.Snapshot().Scene != "title" {
		t.Fatalf("new game scene=%s want title", g.Snapshot().Scene)
	}
	if err := g.Drive(Input{BladeToggle: true}, Keys{}); err != nil {
		t.Fatal(err)
	}
	s := g.Snapshot()
	if s.Scene != "play" {
		t.Fatalf("space on title: scene=%s want play", s.Scene)
	}
	if s.Title != WindowTitle {
		t.Fatalf("title=%q", s.Title)
	}
	wantX, wantY, _ := countySpawn(t)
	if s.Y != wantY || s.X != wantX {
		t.Fatalf("spawn %v,%v want %v,%v", s.X, s.Y, wantX, wantY)
	}
	if s.Stance != "BLADE UP" {
		t.Fatalf("stance=%s", s.Stance)
	}
}

func TestHarnessTitleEnterStartsRun(t *testing.T) {
	g := New()
	g.Silence()
	if err := g.Drive(Input{}, Keys{Enter: true}); err != nil {
		t.Fatal(err)
	}
	if g.Snapshot().Scene != "play" {
		t.Fatalf("enter on title: scene=%s want play", g.Snapshot().Scene)
	}
}

func TestHarnessWMovesNorthThroughUpdate(t *testing.T) {
	g := New()
	g.Silence()
	if err := g.Drive(Input{BladeToggle: true}, Keys{}); err != nil {
		t.Fatal(err)
	}
	y0 := g.dozer.Y
	for i := 0; i < TPS; i++ {
		if err := g.Drive(Input{Throttle: 1}, Keys{}); err != nil {
			t.Fatal(err)
		}
	}
	if g.dozer.Y >= y0-40 {
		t.Fatalf("W 1s via Drive: Y=%v want < %v", g.dozer.Y, y0-40)
	}
}

func TestHarnessATurnsWestThroughUpdate(t *testing.T) {
	g := New()
	g.Silence()
	if err := g.Drive(Input{BladeToggle: true}, Keys{}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < TPS/2; i++ {
		if err := g.Drive(Input{Steer: -1}, Keys{}); err != nil {
			t.Fatal(err)
		}
	}
	h := g.dozer.Heading
	west := (h > math.Pi && h < 2*math.Pi) || (h < 0 && h > -math.Pi)
	if !west {
		t.Fatalf("A 0.5s via Drive: heading=%v want west-of-north", h)
	}
}

func TestHarnessEscReturnsTitle(t *testing.T) {
	g := New()
	g.Silence()
	if err := g.Drive(Input{BladeToggle: true}, Keys{}); err != nil {
		t.Fatal(err)
	}
	if err := g.Drive(Input{}, Keys{Escape: true}); err != nil {
		t.Fatal(err)
	}
	if g.Snapshot().Scene != "title" {
		t.Fatalf("esc: scene=%s want title", g.Snapshot().Scene)
	}
}

func TestHarnessBladeToggle(t *testing.T) {
	g := New()
	g.Silence()
	if err := g.Drive(Input{BladeToggle: true}, Keys{}); err != nil {
		t.Fatal(err)
	}
	if err := g.Drive(Input{BladeToggle: true}, Keys{}); err != nil {
		t.Fatal(err)
	}
	if !g.Snapshot().BladeDown {
		t.Fatal("space in play should drop the blade")
	}
	if g.Snapshot().Stance != "BLADE DOWN" {
		t.Fatalf("stance=%s", g.Snapshot().Stance)
	}
}
