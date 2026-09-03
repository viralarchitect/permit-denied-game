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
	g.blockers = append(g.blockers,
		threats.Dump(220, 232, 40, 24, DumpHP),
		threats.Dump(264, 232, 40, 24, DumpHP),
	)
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

func TestKillJerseyLeavesRubble(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	g.blockers = append(g.blockers, threats.Jersey(176, 216, 48, 16, JerseyHP))
	jersey := &g.blockers[len(g.blockers)-1]
	before := len(g.lot.Rubble)
	g.killBlocker(jersey)
	if len(g.lot.Rubble) != before+1 {
		t.Fatalf("rubble len=%d want %d", len(g.lot.Rubble), before+1)
	}
	r := g.lot.Rubble[len(g.lot.Rubble)-1]
	cx, cy := r.X+r.W/2, r.Y+r.H/2
	_, _, _, hit := CircleAABB(cx, cy, 1, r.X, r.Y, r.W, r.H)
	if !hit {
		t.Fatal("circle at pile center should hit rubble AABB")
	}
	wantW := jersey.W - 2*RubbleInset
	wantH := jersey.H - 2*RubbleInset
	if r.W != wantW || r.H != wantH {
		t.Fatalf("rubble %vx%v want %vx%v (inset %v)", r.W, r.H, wantW, wantH, RubbleInset)
	}
}

func TestKillDumpLeavesRubble(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	g.blockers = append(g.blockers, threats.Dump(220, 232, 40, 24, DumpHP))
	before := len(g.lot.Rubble)
	g.killBlocker(&g.blockers[len(g.blockers)-1])
	if len(g.lot.Rubble) != before+1 {
		t.Fatalf("rubble len=%d want %d", len(g.lot.Rubble), before+1)
	}
	r := g.lot.Rubble[len(g.lot.Rubble)-1]
	cx, cy := r.X+r.W/2, r.Y+r.H/2
	_, _, _, hit := CircleAABB(cx, cy, 1, r.X, r.Y, r.W, r.H)
	if !hit {
		t.Fatal("circle at pile center should hit rubble AABB")
	}
	wantW := 40 - 2*RubbleInset
	wantH := 24 - 2*RubbleInset
	if r.W != wantW || r.H != wantH {
		t.Fatalf("rubble %vx%v want %vx%v", r.W, r.H, wantW, wantH)
	}
}

func TestYardDespawnDumpsLeaveNoRubble(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	g.blockers = append(g.blockers,
		threats.Dump(220, 232, 40, 24, DumpHP),
		threats.Dump(264, 232, 40, 24, DumpHP),
	)
	living := 0
	for _, b := range g.blockers {
		if b.Kind == threats.BlockerDump && b.Alive {
			living++
		}
	}
	if living < 2 {
		t.Fatalf("startRun living dumps=%d want >=2", living)
	}
	before := len(g.lot.Rubble)
	yard := g.lot.BuildingByID(lot.TargetYard)
	if yard == nil {
		t.Fatal("missing yard")
	}
	g.destroyBuilding(yard)
	for _, b := range g.blockers {
		if b.Kind == threats.BlockerDump && b.Alive {
			t.Fatalf("living dump at (%v,%v) after yard destroy", b.X, b.Y)
		}
	}
	for _, r := range g.lot.Rubble {
		if overlapsDumpCell(r, 220, 232, 40, 24) || overlapsDumpCell(r, 264, 232, 40, 24) {
			t.Fatalf("unexpected rubble at dump cell (%v,%v) size %vx%v", r.X, r.Y, r.W, r.H)
		}
	}
	if len(g.lot.Rubble) != before+1 {
		t.Fatalf("yard destroy rubble: before=%d after=%d want +1 (yard pile only)", before, len(g.lot.Rubble))
	}
}

func overlapsDumpCell(r lot.Rubble, dx, dy, dw, dh float64) bool {
	return r.X < dx+dw && r.X+r.W > dx && r.Y < dy+dh && r.Y+r.H > dy
}

func TestRubbleNeverCulled(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	g.blockers = append(g.blockers, threats.Jersey(176, 216, 48, 16, JerseyHP))
	jersey := &g.blockers[len(g.blockers)-1]
	g.killBlocker(jersey)
	g.blockers = append(g.blockers, threats.Dump(220, 232, 40, 24, DumpHP))
	g.killBlocker(&g.blockers[len(g.blockers)-1])
	if len(g.lot.Rubble) != 2 {
		t.Fatalf("rubble len=%d want 2", len(g.lot.Rubble))
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
	x, y, heading := countySpawn(t)
	g.dozer.X = x
	g.dozer.Y = y
	g.dozer.Heading = heading
	g.dozer.BladeDown = false
	g.dozer.IFrames = 0
	g.dozer.Speed = 0
	plates := g.dozer.Plates
	g.cruisers = []threats.Cruiser{
		{X: x + 16, Y: y, Alive: true},
	}
	g.stepCruisers()
	if g.dozer.Plates != plates {
		t.Fatalf("plates %d -> %d after side overlap with PIT off; want no peel", plates, g.dozer.Plates)
	}
}

// TestApplyConcreteLawCurrentWritersForceSolid locks the known spawn-vs-writer
// lie, not the spec. Wet spawn is Solid:false (threats.Concrete). Both writers
// currently force Solid:true and stamp HP. A later approved fix must change
// these assertions on purpose. Do not "fix" applyConcreteLaw to match this.
func TestApplyConcreteLawCurrentWritersForceSolid(t *testing.T) {
	set := threats.Concrete(0, 0, 16, 16)
	applyConcreteLaw(&set, true)
	if !set.Set || !set.Solid || set.HP != ConcreteSetHP {
		t.Fatalf("sets=true: Set=%v Solid=%v HP=%v want Set&&Solid&&HP==%v", set.Set, set.Solid, set.HP, ConcreteSetHP)
	}

	unset := threats.Concrete(0, 0, 16, 16)
	applyConcreteLaw(&unset, false)
	if unset.Set || !unset.Solid || unset.HP != ConcreteUnsetHP {
		t.Fatalf("sets=false: Set=%v Solid=%v HP=%v want Set==false&&Solid==true&&HP==%v", unset.Set, unset.Solid, unset.HP, ConcreteUnsetHP)
	}
}

// TestKillBlockerRubbleIsCollectSolid is show-off test 5's collision half:
// after killBlocker, the new rubble AABB is in collectSolids. Existing
// TestKillJerseyLeavesRubble / TestKillDumpLeavesRubble already assert
// lot.Rubble + inset + circle-vs-pile; they do not call collectSolids.
func TestKillBlockerRubbleIsCollectSolid(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	g.blockers = append(g.blockers, threats.Jersey(176, 216, 48, 16, JerseyHP))
	jersey := &g.blockers[len(g.blockers)-1]
	g.killBlocker(jersey)
	r := g.lot.Rubble[len(g.lot.Rubble)-1]
	found := false
	for _, s := range g.collectSolids() {
		if s.x == r.X && s.y == r.Y && s.w == r.W && s.h == r.H {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("rubble AABB %v,%v %vx%v missing from collectSolids", r.X, r.Y, r.W, r.H)
	}
}

func TestDozerPackBuildingRubbleBlocksMovement(t *testing.T) {
	g := New()
	g.audio = nil
	g.startRun()
	b := firstCountyMundane(g)
	if b == nil {
		t.Fatal("missing mundane county building")
	}
	g.destroyBuilding(b)
	r := g.lot.Rubble[len(g.lot.Rubble)-1]
	if r.Ramp {
		t.Fatalf("dozer rubble ramp=%v want false", r.Ramp)
	}
	g.dozer.X = r.X + r.W/2
	g.dozer.Y = r.Y + r.H + DozerBodyR - 1
	g.dozer.Heading = 0
	g.dozer.BladeDown = false
	for i := 0; i < TPS/2; i++ {
		g.stepPlay(Input{Throttle: 1})
		if g.run.Over {
			t.Fatalf("run ended while testing rubble block: %q", g.run.Death)
		}
	}
	if g.dozer.Y < r.Y+r.H+DozerBodyR-0.5 {
		t.Fatalf("dozer crossed solid rubble: Y=%v rubbleBottom=%v", g.dozer.Y, r.Y+r.H)
	}
}
