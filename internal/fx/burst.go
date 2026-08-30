package fx

import "fmt"

type BurstKind int

const (
	BurstBoom BurstKind = iota
	BurstSpark
)

const (
	boomLife  = 10
	sparkLife = 6
)

type Burst struct {
	X, Y float64
	Age  int
	Kind BurstKind
}

func (f *FX) SpawnBoom(x, y float64) {
	f.Bursts = append(f.Bursts, Burst{X: x, Y: y, Kind: BurstBoom})
}

func (f *FX) SpawnSpark(x, y float64) {
	f.Bursts = append(f.Bursts, Burst{X: x, Y: y, Kind: BurstSpark})
}

func (b Burst) Dead() bool {
	if b.Kind == BurstSpark {
		return b.Age >= sparkLife
	}
	return b.Age >= boomLife
}

func (b Burst) FrameName() string {
	if b.Kind == BurstSpark {
		i := b.Age * 4 / sparkLife
		if i > 3 {
			i = 3
		}
		return fmt.Sprintf("spark_%02d", i)
	}
	i := b.Age * 6 / boomLife
	if i > 5 {
		i = 5
	}
	return fmt.Sprintf("boom_%02d", i)
}
