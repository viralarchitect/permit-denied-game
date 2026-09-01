package capitol

import (
	"permitdenied/internal/lot"
	"permitdenied/internal/mapgen"
)

type Tower struct {
	CoreIdx       int
	PenthouseIdx  int
	FloorIdxs     []int
	Reached       bool
	CoreCollapsed bool
}

func Template() mapgen.Arena {
	return mapgen.Arena{
		W: 960, H: 1600,
		SpawnX: 480, SpawnY: 1500,
		Clock: 330,
		Streets: []mapgen.Street{
			{X: 400, Y: 0, W: 160, H: 1600},
			{X: 0, Y: 1280, W: 960, H: 40},
			{X: 0, Y: 1100, W: 960, H: 32},
		},
		Slots: []mapgen.Slot{
			{X: 80, Y: 1320, W: 80, H: 48, Material: lot.MatBrick, Label: "", Role: lot.RoleMundane, Value: 70},
			{X: 800, Y: 1320, W: 80, H: 48, Material: lot.MatBrick, Label: "", Role: lot.RoleMundane, Value: 70},
			{X: 64, Y: 1140, W: 96, H: 64, Material: lot.MatConcrete, Label: "ANNEX", Role: lot.RoleTarget, Value: 400},
			{X: 800, Y: 1140, W: 96, H: 64, Material: lot.MatConcrete, Label: "WING", Role: lot.RoleTarget, Value: 400},
			{X: 380, Y: 880, W: 200, H: 120, Material: lot.MatConcrete, Label: "FLOOR 1", Role: lot.RoleFloor, Value: 300},
			{X: 380, Y: 700, W: 200, H: 120, Material: lot.MatConcrete, Label: "FLOOR 2", Role: lot.RoleFloor, Value: 350},
			{X: 380, Y: 520, W: 200, H: 120, Material: lot.MatSteel, Label: "FLOOR 3", Role: lot.RoleFloor, Value: 450},
			{X: 440, Y: 360, W: 80, H: 100, Material: lot.MatSteel, Label: "CORE", Role: lot.RoleCore, Value: 900},
			{X: 400, Y: 80, W: 160, H: 120, Material: lot.MatSteel, Label: "PENTHOUSE", Role: lot.RolePenthouse, Value: 2000},
		},
	}
}

func New(hpPerTile float64) (lot.Lot, Tower) {
	a := Template()
	l := lot.Lot{Buildings: mapgen.Buildings(a, hpPerTile)}
	t := Index(l)
	return l, t
}

func Index(l lot.Lot) Tower {
	t := Tower{CoreIdx: -1, PenthouseIdx: -1}
	for i := range l.Buildings {
		switch l.Buildings[i].Role {
		case lot.RoleCore:
			t.CoreIdx = i
		case lot.RolePenthouse:
			t.PenthouseIdx = i
		case lot.RoleFloor:
			t.FloorIdxs = append(t.FloorIdxs, i)
		}
	}
	return t
}

func (t *Tower) MarkReached(x, y float64, l lot.Lot) {
	if t.Reached || t.PenthouseIdx < 0 || t.PenthouseIdx >= len(l.Buildings) {
		return
	}
	b := l.Buildings[t.PenthouseIdx]
	if x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H {
		t.Reached = true
	}
}

func (t Tower) CoreDown(l lot.Lot) bool {
	if t.CoreIdx < 0 || t.CoreIdx >= len(l.Buildings) {
		return false
	}
	return l.Buildings[t.CoreIdx].State == lot.InRubble || l.Buildings[t.CoreIdx].HP <= 0
}

func (t Tower) PenthouseDown(l lot.Lot) bool {
	if t.PenthouseIdx < 0 || t.PenthouseIdx >= len(l.Buildings) {
		return false
	}
	return l.Buildings[t.PenthouseIdx].State == lot.InRubble || l.Buildings[t.PenthouseIdx].HP <= 0
}

func (t Tower) Cleared(l lot.Lot) bool {
	if t.CoreDown(l) {
		return true
	}
	return t.PenthouseDown(l) && t.Reached
}

// CollapseFromCore wrecks every remaining floor and the penthouse.
func (t *Tower) CollapseFromCore(l *lot.Lot) []int {
	if t.CoreCollapsed {
		return nil
	}
	t.CoreCollapsed = true
	var hit []int
	collapse := func(i int) {
		if i < 0 || i >= len(l.Buildings) {
			return
		}
		b := &l.Buildings[i]
		if b.State == lot.InRubble {
			return
		}
		b.HP = 0
		b.State = lot.InRubble
		hit = append(hit, i)
	}
	for _, i := range t.FloorIdxs {
		collapse(i)
	}
	collapse(t.PenthouseIdx)
	if t.CoreIdx >= 0 && t.CoreIdx < len(l.Buildings) {
		b := &l.Buildings[t.CoreIdx]
		if b.State != lot.InRubble {
			b.HP = 0
			b.State = lot.InRubble
			hit = append(hit, t.CoreIdx)
		}
	}
	return hit
}

func IsRampRole(r lot.Role) bool {
	return r == lot.RoleFloor || r == lot.RolePenthouse
}
