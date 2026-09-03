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
	TargetCivic
	TargetCore
	TargetPenthouse
	TargetFloor
)

type Material int

const (
	MatWood Material = iota
	MatBrick
	MatConcrete
	MatSteel
)

type Role int

const (
	RoleMundane Role = iota
	RoleTarget
	RoleFloor
	RoleCore
	RolePenthouse
)

type Building struct {
	ID         TargetID // TargetNone for mundane
	Label      string   // "SHERIFF", "YARD", "PLANT", or civic name
	X, Y, W, H float64
	HP, MaxHP  float64
	State      BuildingState
	Value      int // structure $ when fully rubble
	Material   Material
	Role       Role
	// Authored rubble settings come from the frozen pack/destructible contract.
	// Legacy literal maps leave these zeroed and keep their hardcoded fallback.
	AuthoredRubble bool
	SpawnsRubble   bool
	RubbleInset    float64
	RubbleRamp     bool
}

type Rubble struct { // axis-aligned, colliding unless Ramp
	X, Y, W, H float64
	Mass       float64
	Ramp       bool
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

func (b *Building) Named() bool {
	return b.Label != "" && b.Role != RoleMundane
}

func MaxHP(w, h, hpPerTile float64) float64 {
	return (w / 16) * (h / 16) * hpPerTile
}

func MaxHPMat(w, h, hpPerTile float64, mat Material) float64 {
	mul := 1.0
	switch mat {
	case MatWood:
		mul = 0.70
	case MatBrick:
		mul = 1.00
	case MatConcrete:
		mul = 1.35
	case MatSteel:
		mul = 1.85
	}
	return MaxHP(w, h, hpPerTile) * mul
}
