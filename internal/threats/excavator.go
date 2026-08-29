package threats

type Excavator struct {
	X, Y, Heading             float64
	Announced, Arrived, Alive bool
	BoomPhase                 float64 // 0..1 swipe
	Corridor                  int
	HP                        float64
	Swinging                  bool
}

func (e *Excavator) Body() (x, y, w, h float64) {
	return e.X - 10, e.Y - 18, 20, 36
}
