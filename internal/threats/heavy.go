package threats

type HeavyKind int

const (
	HeavyFire HeavyKind = iota
	HeavyWagon
)

type Heavy struct {
	Kind         HeavyKind
	X, Y         float64
	W, H         float64
	Heading      float64
	HP           float64
	Alive, Solid bool
	Parked       bool
}

func FireTruck(x, y, w, h, hp float64) Heavy {
	return Heavy{Kind: HeavyFire, X: x, Y: y, W: w, H: h, HP: hp, Alive: true, Solid: true}
}

func Wagon(x, y, w, h, hp float64) Heavy {
	return Heavy{Kind: HeavyWagon, X: x, Y: y, W: w, H: h, HP: hp, Alive: true, Solid: true}
}

func (h Heavy) Body() (x, y, w, hh float64) {
	return h.X - h.W/2, h.Y - h.H/2, h.W, h.H
}
