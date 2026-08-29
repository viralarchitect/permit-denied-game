package lot

// Lot holds smashable geometry. Layout is literals (implementation guide §7).
type Lot struct {
	Buildings []Building
	Rubble    []Rubble
}

func New(hpPerTile float64) Lot {
	mk := func(id TargetID, label string, x, y, w, h float64, value int) Building {
		hp := MaxHP(w, h, hpPerTile)
		return Building{
			ID: id, Label: label,
			X: x, Y: y, W: w, H: h,
			HP: hp, MaxHP: hp,
			State: Intact, Value: value,
		}
	}
	return Lot{
		Buildings: []Building{
			mk(TargetSheriff, "SHERIFF", 176, 640, 96, 80, 400),
			mk(TargetYard, "YARD", 448, 720, 128, 96, 500),
			mk(TargetPlant, "PLANT", 248, 72, 144, 112, 700),
			// mundane west shacks
			mk(TargetNone, "", 64, 900, 64, 48, 50),
			mk(TargetNone, "", 80, 500, 80, 48, 65),
			mk(TargetNone, "", 48, 300, 64, 64, 80),
			// mundane east shacks
			mk(TargetNone, "", 448, 1000, 72, 48, 55),
			mk(TargetNone, "", 480, 560, 64, 40, 40),
		},
	}
}

func (l *Lot) BuildingByID(id TargetID) *Building {
	for i := range l.Buildings {
		if l.Buildings[i].ID == id {
			return &l.Buildings[i]
		}
	}
	return nil
}

func (l *Lot) AddRubble(x, y, w, h, inset float64) {
	if w > inset*2 && h > inset*2 {
		x += inset
		y += inset
		w -= inset * 2
		h -= inset * 2
	}
	l.Rubble = append(l.Rubble, Rubble{X: x, Y: y, W: w, H: h})
}

func (l *Lot) RubbleBlocks(x, y, w, h float64) bool {
	for i := range l.Rubble {
		r := l.Rubble[i]
		if x < r.X+r.W && x+w > r.X && y < r.Y+r.H && y+h > r.Y {
			return true
		}
	}
	return false
}
