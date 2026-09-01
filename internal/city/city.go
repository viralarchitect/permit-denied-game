package city

import (
	"permitdenied/internal/lot"
	"permitdenied/internal/mapgen"
)

func Template(alt bool) mapgen.Arena {
	a := mapgen.Arena{
		W: 1280, H: 1280,
		SpawnX: 640, SpawnY: 1180,
		Clock: 300,
		Streets: []mapgen.Street{
			{X: 0, Y: 240, W: 1280, H: 40},
			{X: 0, Y: 480, W: 1280, H: 40},
			{X: 0, Y: 720, W: 1280, H: 40},
			{X: 0, Y: 960, W: 1280, H: 40},
			{X: 200, Y: 0, W: 40, H: 1280},
			{X: 620, Y: 0, W: 40, H: 1280},
			{X: 1040, Y: 0, W: 40, H: 1280},
		},
		Slots: []mapgen.Slot{
			{X: 260, Y: 80, W: 80, H: 128, Material: lot.MatSteel, Label: "MID-RISE", Role: lot.RoleTarget, Value: 1100},
			{X: 700, Y: 80, W: 128, H: 96, Material: lot.MatConcrete, Label: "GARAGE", Role: lot.RoleTarget, Value: 900},
			{X: 260, Y: 1040, W: 144, H: 80, Material: lot.MatBrick, Label: "BUS BARN", Role: lot.RoleTarget, Value: 750},
			{X: 280, Y: 520, W: 160, H: 32, Material: lot.MatConcrete, Label: "OVERPASS", Role: lot.RoleTarget, Value: 600},
			{X: 700, Y: 280, W: 96, H: 80, Material: lot.MatSteel, Label: "", Role: lot.RoleMundane, Value: 180},
			{X: 1080, Y: 280, W: 80, H: 64, Material: lot.MatBrick, Label: "", Role: lot.RoleMundane, Value: 90},
			{X: 80, Y: 280, W: 80, H: 64, Material: lot.MatConcrete, Label: "", Role: lot.RoleMundane, Value: 110},
			{X: 700, Y: 1040, W: 96, H: 64, Material: lot.MatWood, Label: "", Role: lot.RoleMundane, Value: 55},
			{X: 1080, Y: 760, W: 72, H: 48, Material: lot.MatBrick, Label: "", Role: lot.RoleMundane, Value: 70},
		},
	}
	if alt {
		a.Slots[0].X, a.Slots[0].Y = 700, 80
		a.Slots[1].X, a.Slots[1].Y = 260, 80
		a.Slots[2].X, a.Slots[2].Y = 700, 1040
	}
	return a
}

func New(seed int64, alt bool) mapgen.Arena {
	return mapgen.Shuffle(Template(alt), seed, 16)
}
