package threats

type Cruiser struct {
	X, Y, Heading float64
	Alive         bool
}

func SpawnCruiser(x, y float64) Cruiser {
	return Cruiser{X: x, Y: y, Alive: true}
}
