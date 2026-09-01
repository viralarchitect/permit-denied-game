package attach

import "permitdenied/internal/lot"

type Kind int

const (
	Ripper Kind = iota
	WideBlade
	WreckingBall
	PileDriver
	ExtraPlate
	KindCount
)

type Kit struct {
	Ripper, Wide, Ball, Driver, Plate bool
}

type Pickup struct {
	Kind       Kind
	X, Y, W, H float64
	Alive      bool
}

func (k Kind) Label() string {
	switch k {
	case Ripper:
		return "RIPPER"
	case WideBlade:
		return "WIDE BLADE"
	case WreckingBall:
		return "WRECKING BALL"
	case PileDriver:
		return "PILE DRIVER"
	case ExtraPlate:
		return "EXTRA PLATE"
	default:
		return "ATTACHMENT"
	}
}

func (k *Kit) Grant(kind Kind) {
	switch kind {
	case Ripper:
		k.Ripper = true
	case WideBlade:
		k.Wide = true
	case WreckingBall:
		k.Ball = true
	case PileDriver:
		k.Driver = true
	case ExtraPlate:
		k.Plate = true
	}
}

func (k Kit) Has(kind Kind) bool {
	switch kind {
	case Ripper:
		return k.Ripper
	case WideBlade:
		return k.Wide
	case WreckingBall:
		return k.Ball
	case PileDriver:
		return k.Driver
	case ExtraPlate:
		return k.Plate
	default:
		return false
	}
}

func (k Kit) Count() int {
	n := 0
	if k.Ripper {
		n++
	}
	if k.Wide {
		n++
	}
	if k.Ball {
		n++
	}
	if k.Driver {
		n++
	}
	if k.Plate {
		n++
	}
	return n
}

func (k Kit) WreckMul(mat lot.Material) float64 {
	mul := 1.0
	switch mat {
	case lot.MatWood:
		mul = 1.00
	case lot.MatBrick:
		mul = 0.85
	case lot.MatConcrete:
		mul = 0.50
		if k.Ripper {
			mul *= 2.10
		}
		if k.Driver {
			mul *= 1.55
		}
	case lot.MatSteel:
		mul = 0.22
		if k.Ripper {
			mul *= 2.40
		}
		if k.Ball {
			mul *= 2.80
		}
		if k.Driver {
			mul *= 1.55
		}
	}
	return mul
}

func Spawn(kind Kind, x, y, size float64) Pickup {
	return Pickup{Kind: kind, X: x, Y: y, W: size, H: size, Alive: true}
}
