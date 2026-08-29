package threats

type BlockerKind int

const (
	BlockerJersey BlockerKind = iota
	BlockerDump
	BlockerConcrete // sets at TConcreteSet unless plant dead
)

type Blocker struct {
	Kind       BlockerKind
	X, Y, W, H float64
	Set        bool // concrete only
	HP         float64
	Alive      bool
	Solid      bool
}

func Jersey(x, y, w, h, hp float64) Blocker {
	return Blocker{Kind: BlockerJersey, X: x, Y: y, W: w, H: h, HP: hp, Alive: true, Solid: true}
}

func Dump(x, y, w, h, hp float64) Blocker {
	return Blocker{Kind: BlockerDump, X: x, Y: y, W: w, H: h, HP: hp, Alive: true, Solid: true}
}

func Concrete(x, y, w, h float64) Blocker {
	return Blocker{Kind: BlockerConcrete, X: x, Y: y, W: w, H: h, Alive: true, Solid: false, Set: false}
}
