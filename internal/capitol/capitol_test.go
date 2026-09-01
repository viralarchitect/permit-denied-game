package capitol

import (
	"testing"
)

func TestCoreZeroCollapsesPenthouse(t *testing.T) {
	l, tw := New(2.0)
	if tw.CoreIdx < 0 || tw.PenthouseIdx < 0 {
		t.Fatal("missing core or penthouse")
	}
	core := &l.Buildings[tw.CoreIdx]
	if !core.ApplyDamage(core.HP) {
		t.Fatal("core should fall")
	}
	if !tw.CoreDown(l) {
		t.Fatal("core not down")
	}
	hit := tw.CollapseFromCore(&l)
	if len(hit) == 0 {
		t.Fatal("collapse returned no buildings")
	}
	if !tw.PenthouseDown(l) {
		t.Fatal("penthouse still standing after core collapse")
	}
	if !tw.Cleared(l) {
		t.Fatal("core collapse should clear the tower")
	}
}

func TestPenthouseZeroWinsIfReached(t *testing.T) {
	l, tw := New(2.0)
	ph := &l.Buildings[tw.PenthouseIdx]
	cx, cy := ph.Center()
	tw.MarkReached(cx, cy, l)
	if !tw.Reached {
		t.Fatal("player on penthouse rect should count as reached")
	}
	if !ph.ApplyDamage(ph.HP) {
		t.Fatal("penthouse should fall")
	}
	if !tw.PenthouseDown(l) {
		t.Fatal("penthouse not down")
	}
	if !tw.Cleared(l) {
		t.Fatal("reached + penthouse rubble should win")
	}
}

func TestPenthouseZeroWithoutReachIsNotWin(t *testing.T) {
	l, tw := New(2.0)
	ph := &l.Buildings[tw.PenthouseIdx]
	ph.ApplyDamage(ph.HP)
	if tw.Cleared(l) {
		t.Fatal("penthouse rubble without reach should not win (core still up)")
	}
}
