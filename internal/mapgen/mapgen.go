package mapgen

import (
	"math/rand"

	"permitdenied/internal/lot"
)

type Slot struct {
	X, Y, W, H float64
	Material   lot.Material
	Label      string
	Role       lot.Role
	Value      int
}

type Street struct {
	X, Y, W, H float64
}

type Arena struct {
	W, H           float64
	SpawnX, SpawnY float64
	Clock          float64
	Slots          []Slot
	Streets        []Street
}

func Shuffle(a Arena, seed int64, jitter float64) Arena {
	rng := rand.New(rand.NewSource(seed))
	civ := make([]int, 0, 8)
	for i, s := range a.Slots {
		if s.Role == lot.RoleTarget && s.Label != "" {
			civ = append(civ, i)
		}
	}
	if len(civ) > 1 {
		rng.Shuffle(len(civ), func(i, j int) {
			ii, jj := civ[i], civ[j]
			a.Slots[ii].X, a.Slots[jj].X = a.Slots[jj].X, a.Slots[ii].X
			a.Slots[ii].Y, a.Slots[jj].Y = a.Slots[jj].Y, a.Slots[ii].Y
		})
	}
	if jitter > 0 {
		step := 16.0
		for i := range a.Slots {
			if a.Slots[i].Role == lot.RoleCore || a.Slots[i].Role == lot.RolePenthouse || a.Slots[i].Role == lot.RoleFloor {
				continue
			}
			dx := float64(rng.Intn(3)-1) * step
			dy := float64(rng.Intn(3)-1) * step
			if dx > jitter {
				dx = jitter
			}
			if dy > jitter {
				dy = jitter
			}
			a.Slots[i].X += dx
			a.Slots[i].Y += dy
			if a.Slots[i].X < 16 {
				a.Slots[i].X = 16
			}
			if a.Slots[i].Y < 16 {
				a.Slots[i].Y = 16
			}
		}
	}
	return a
}

func Buildings(a Arena, hpPerTile float64) []lot.Building {
	out := make([]lot.Building, 0, len(a.Slots))
	for _, s := range a.Slots {
		id := lot.TargetNone
		switch s.Role {
		case lot.RoleTarget:
			id = lot.TargetCivic
		case lot.RoleCore:
			id = lot.TargetCore
		case lot.RolePenthouse:
			id = lot.TargetPenthouse
		case lot.RoleFloor:
			id = lot.TargetFloor
		}
		hp := lot.MaxHPMat(s.W, s.H, hpPerTile, s.Material)
		out = append(out, lot.Building{
			ID: id, Label: s.Label,
			X: s.X, Y: s.Y, W: s.W, H: s.H,
			HP: hp, MaxHP: hp,
			State: lot.Intact, Value: s.Value,
			Material: s.Material, Role: s.Role,
		})
	}
	return out
}
