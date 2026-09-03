package game

import (
	"testing"

	"permitdenied/internal/fx"
	"permitdenied/internal/lot"
)

func poseOnWestShack(g *Game) *lot.Building {
	b := firstCountyMundane(g)
	if b != nil {
		poseSouthOf(g, b, true)
	}
	g.dozer.Heat = 0
	return b
}

func poseOnSheriff(g *Game) *lot.Building {
	b := g.lot.BuildingByID(lot.TargetSheriff)
	if b != nil {
		poseSouthOf(g, b, true)
	}
	return b
}

func pulseWreck(g *Game, b *lot.Building, heatStop float64, maxTicks int) {
	g.dozer.BladeDown = true
	for i := 0; i < maxTicks; i++ {
		if b.State == lot.InRubble || g.dozer.Heat >= heatStop {
			return
		}
		g.stepPlay(Input{Throttle: 1})
	}
}

func coolOnAsphalt(g *Game, maxTicks int) {
	x, y, heading := mustCountySpawn()
	g.dozer.X = x
	g.dozer.Y = y
	g.dozer.Heading = heading
	g.dozer.BladeDown = false
	g.dozer.Speed = 0
	for i := 0; i < maxTicks; i++ {
		if g.dozer.Heat <= 20 {
			return
		}
		g.stepPlay(Input{})
	}
}

func burstKind(g *Game, kind fx.BurstKind) bool {
	for _, b := range g.fx.Bursts {
		if b.Kind == kind {
			return true
		}
	}
	return false
}

func TestWreckShackOnePulse(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	b := poseOnWestShack(g)
	if b == nil {
		t.Fatal("missing county mundane shack")
	}
	pulseWreck(g, b, HeatMax, 20*TPS)
	if b.State != lot.InRubble {
		t.Fatalf("shack state=%v HP=%v want InRubble", b.State, b.HP)
	}
	if g.dozer.Heat >= 90 {
		t.Fatalf("shack one-pulse heat=%v want < 90", g.dozer.Heat)
	}
}

func TestWreckSheriffTwoPulses(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	b := poseOnSheriff(g)
	if b == nil {
		t.Fatal("missing sheriff")
	}
	pulseWreck(g, b, 80, 20*TPS)
	if b.State != lot.Cracked && b.State != lot.InRubble {
		t.Fatalf("after pulse 1: state=%v HP=%v want Cracked or InRubble", b.State, b.HP)
	}
	if b.State != lot.InRubble {
		coolOnAsphalt(g, 90)
		poseOnSheriff(g)
		pulseWreck(g, b, 80, 20*TPS)
	}
	if b.State != lot.InRubble {
		t.Fatalf("after pulse 2: state=%v HP=%v want InRubble", b.State, b.HP)
	}

	g = New()
	g.Silence()
	g.startRun()
	b = poseOnSheriff(g)
	reached := false
	for i := 0; i < 20*TPS; i++ {
		g.stepPlay(Input{Throttle: 1})
		if g.dozer.Heat >= HeatMax || g.run.Death == "cooked" {
			reached = true
			break
		}
	}
	if !reached {
		t.Fatalf("planted through both pulses: heat=%v death=%q want HeatMax", g.dozer.Heat, g.run.Death)
	}
}

func TestWreckChaseDuck(t *testing.T) {
	g := New()
	g.startRun()
	g.audio.StartChase()
	if v := g.audio.ChaseVolume(); v != 0.35 {
		t.Fatalf("chase start volume=%v want 0.35", v)
	}
	poseOnSheriff(g)
	g.stepPlay(Input{Throttle: 1})
	if v := g.audio.ChaseVolume(); v >= 0.20 {
		t.Fatalf("eating ChaseVolume=%v want < 0.20", v)
	}
	g.dozer.BladeDown = false
	x, y, _ := countySpawn(t)
	g.dozer.X = x
	g.dozer.Y = y
	g.stepPlay(Input{})
	if v := g.audio.ChaseVolume(); v != 0.35 {
		t.Fatalf("after overlap ends ChaseVolume=%v want 0.35", v)
	}
	if g.scene != ScenePlay {
		t.Fatalf("scene=%v want play (not tally)", g.scene)
	}
}

func TestWreckChipThenCollapse(t *testing.T) {
	g := New()
	g.Silence()
	g.startRun()
	b := poseOnSheriff(g)
	if b == nil {
		t.Fatal("missing sheriff")
	}
	sawSpark := false
	sawShake := false
	var killFrame string
	for i := 0; i < 20*TPS; i++ {
		was := b.State
		g.stepPlay(Input{Throttle: 1})
		if was != lot.InRubble && b.State == lot.InRubble {
			for _, burst := range g.fx.Bursts {
				if burst.Kind == fx.BurstBoom {
					killFrame = burst.FrameName()
					break
				}
			}
			break
		}
		if burstKind(g, fx.BurstSpark) {
			sawSpark = true
		}
		if g.fx.Shake > 0 {
			sawShake = true
		}
	}
	if !sawSpark || !sawShake {
		t.Fatalf("chip before destroy: spark=%v shake=%v", sawSpark, sawShake)
	}
	if b.State != lot.InRubble {
		t.Fatalf("sheriff state=%v want InRubble", b.State)
	}
	if killFrame != "boom_00" {
		t.Fatalf("collapse frame=%q want boom_00", killFrame)
	}
	if !g.run.SheriffDown {
		t.Fatal("collapse skipped destroyBuilding (SheriffDown false)")
	}
}
