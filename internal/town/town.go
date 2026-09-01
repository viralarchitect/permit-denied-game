package town

import (
	"permitdenied/internal/lot"
	"permitdenied/internal/mapgen"
)

func Template(alt bool) mapgen.Arena {
	a := mapgen.Arena{
		W: 1280, H: 1280,
		SpawnX: 640, SpawnY: 1180,
		Clock: 270,
		Streets: []mapgen.Street{
			{X: 0, Y: 360, W: 1280, H: 48},
			{X: 0, Y: 640, W: 1280, H: 48},
			{X: 0, Y: 920, W: 1280, H: 48},
			{X: 616, Y: 0, W: 48, H: 1280},
		},
		Slots: []mapgen.Slot{
			{X: 160, Y: 200, W: 112, H: 80, Material: lot.MatConcrete, Label: "COURTHOUSE", Role: lot.RoleTarget, Value: 800},
			{X: 800, Y: 200, W: 128, H: 80, Material: lot.MatBrick, Label: "SCHOOL", Role: lot.RoleTarget, Value: 650},
			{X: 160, Y: 1000, W: 112, H: 64, Material: lot.MatConcrete, Label: "DEPOT", Role: lot.RoleTarget, Value: 500},
			{X: 800, Y: 1000, W: 96, H: 64, Material: lot.MatBrick, Label: "", Role: lot.RoleMundane, Value: 90},
			{X: 160, Y: 480, W: 80, H: 48, Material: lot.MatBrick, Label: "", Role: lot.RoleMundane, Value: 70},
			{X: 880, Y: 480, W: 80, H: 48, Material: lot.MatWood, Label: "", Role: lot.RoleMundane, Value: 50},
			{X: 880, Y: 760, W: 72, H: 48, Material: lot.MatWood, Label: "", Role: lot.RoleMundane, Value: 45},
			{X: 200, Y: 760, W: 64, H: 48, Material: lot.MatBrick, Label: "", Role: lot.RoleMundane, Value: 60},
		},
	}
	if alt {
		a.Slots[0].X, a.Slots[0].Y = 800, 1000
		a.Slots[1].X, a.Slots[1].Y = 160, 200
		a.Slots[2].X, a.Slots[2].Y = 800, 200
	}
	return a
}

func New(seed int64, alt bool) mapgen.Arena {
	return mapgen.Shuffle(Template(alt), seed, 32)
}
