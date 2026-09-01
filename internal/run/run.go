package run

// Run is one campaign attempt. County booleans stay the lot boons.
type Run struct {
	Tick                             int
	Over                             bool
	Death                            string // "", "cooked", "track", "buzzer", "pinned", "buried", "cleared"
	StructCash                       int    // spec: Struct$
	VehicleCash                      int    // spec: Vehicle$
	TimeAlive                        float64
	Targets                          int // named wrecks this map / county 0..3
	SheriffDown, YardDown, PlantDown bool
	CruiserPIT                       bool // true until sheriff smashed
	DumpTrucks                       bool
	WallsBrittle                     bool
	ConcreteSets                     bool
	// PressureBonus is seconds added to Time() while the dozer lingers in the
	// chopper spotlight (0.35 * dt). Recommended §8.4 / §11 implementation:
	// t := float64(Tick)/TPS + PressureBonus. It accelerates the beat chart
	// without skipping an un-fired beat in a single step — comparisons are
	// monotonic and each event still has an once-flag.
	PressureBonus float64
	Tier          int // 0 county, 1 town, 2 city, 3 capitol
	Seed          int64
	MapsCleared   int
}

func New() Run {
	return Run{
		CruiserPIT:   true,
		DumpTrucks:   true,
		WallsBrittle: false,
		ConcreteSets: true,
	}
}

func (r *Run) Time() float64 {
	return float64(r.Tick)/60.0 + r.PressureBonus
}

func (r *Run) RecountTargets() {
	n := 0
	if r.SheriffDown {
		n++
	}
	if r.YardDown {
		n++
	}
	if r.PlantDown {
		n++
	}
	r.Targets = n
}

func (r *Run) Raw() int {
	return r.StructCash + r.VehicleCash + int(r.TimeAlive)
}

func (r *Run) Final(mult float64) int {
	return int(float64(r.Raw()) * mult)
}
