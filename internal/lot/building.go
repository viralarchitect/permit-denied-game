package lot

type BuildingState int

const (
	Intact BuildingState = iota
	Cracked
	InRubble // spec name Rubble; colliding pile, never culled
)

type TargetID int

const (
	TargetNone TargetID = iota
	TargetSheriff
	TargetYard
	TargetPlant
)

type Building struct {
	ID         TargetID // TargetNone for mundane
	Label      string   // "SHERIFF", "YARD", "PLANT", or ""
	X, Y, W, H float64
	HP, MaxHP  float64
	State      BuildingState
	Value      int // structure $ when fully rubble
}

type Rubble struct { // axis-aligned, colliding
	X, Y, W, H float64
}

func (b *Building) ApplyDamage(amount float64) (destroyed bool) {
	if b.State == InRubble {
		return false
	}
	b.HP -= amount
	if b.HP <= 0 {
		b.HP = 0
		b.State = InRubble
		return true
	}
	if b.HP <= b.MaxHP*0.5 {
		b.State = Cracked
	}
	return false
}

func (b *Building) Center() (float64, float64) {
	return b.X + b.W/2, b.Y + b.H/2
}

func MaxHP(w, h, hpPerTile float64) float64 {
	return (w / 16) * (h / 16) * hpPerTile
}
