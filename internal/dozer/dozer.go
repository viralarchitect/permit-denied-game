package dozer

// Dozer is the player machine. Fields match the implementation guide §6.
type Dozer struct {
	X, Y      float64
	Heading   float64
	Speed     float64 // signed: +forward along heading, −reverse
	BladeDown bool
	Plates    int     // 4 paint … 1 frame; 0 = dead this tick
	Heat      float64 // 0..100
	IFrames   int     // ticks after a peel
}

func Spawn(x, y float64) Dozer {
	return Dozer{
		X:       x,
		Y:       y,
		Heading: 0,
		Plates:  4,
	}
}
