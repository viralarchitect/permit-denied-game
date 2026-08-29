package game

const (
	TPS     = 60
	Dt      = 1.0 / 60.0
	ScreenW = 320
	ScreenH = 224
	Tile    = 16
	LotW    = 640
	LotH    = 1280

	RunSeconds = 210.0 // 3:30
	BuzzerTick = 210 * TPS

	// beat chart (seconds) — §11
	TBlockers    = 40
	TChopper     = 60
	TExAnnounce  = 80
	TExArrive    = 105
	TConcreteSet = 135
	TTwoFamilies = 165

	DozerBodyR = 14.0
	BladeW     = 32.0
	BladeHDown = 10.0
	BladeHUp   = 6.0
	BladeReach = 18.0 // center → blade center along forward

	PlatesMax       = 4
	HeatMax         = 100.0
	HeatCookPush    = 28.0 // /s blade-down into wall or deep rubble
	HeatCookStall   = 45.0 // /s wedged (speed < 8 and overlapping solid)
	HeatCoolAsphalt = 18.0 // /s blade-up, not overlapping solid
	HeatCoolIdle    = 8.0

	WreckHPPerTile = 2.0  // each 16×16 of a building
	WreckRateDown  = 14.0 // HP/s while blade-down overlapping
	WreckRateUp    = 1.5  // glance

	CruiserSpeed  = 95.0
	CruiserRadius = 8.0
	PedRadius     = 4.0

	HitStopTicks = 2
	ShakeDecay   = 8.0

	Mult0 = 1.00
	Mult1 = 1.25
	Mult2 = 1.60
	Mult3 = 2.00

	// §4.3 tank-steer
	TurnRateBladeUp   = 2.4 // rad/s
	TurnRateBladeDown = 1.1
	SpeedFwdUp        = 110.0 // px/s
	SpeedFwdDown      = 48.0
	SpeedRevUp        = 55.0
	SpeedRevDown      = 28.0
	AccelFwd          = 180.0 // px/s²
	AccelBrake        = 220.0
	AccelRev          = 120.0

	SpawnX = 320.0
	SpawnY = 1180.0

	StallSpeed = 8.0 // §8.7 / §9 stalled if |speed| < 8 and overlapping solid

	IFramesPeel = 45 // ticks after cruiser peel
	IFramesBoom = 60 // ticks after excavator boom

	ChopperSpeed = 70.0 // px/s lazy follow §12
	ChopperSpotR = 72.0
	PressureRate = 0.35 // §11 linger: PressureBonus += 0.35 * dt

	CruiserCap      = 6
	CruiserOffsetF  = 22.0 // seek dozer - Forward*22
	CruiserOffsetS  = 16.0 // + Right*side*16
	CruiserKillCash = 25
	FrontAlong      = 4.0 // §9 along > 4 is front

	ExcavatorHP       = 40.0
	ExcavatorKillCash = 200
	ExcavatorBodyW    = 20.0
	ExcavatorBodyH    = 36.0
	BoomLen           = 48.0
	BoomHalfW         = 5.0
	BoomPeriod        = 0.8
	BoomSweep         = 1.2 // radians peak-to-peak around heading

	PedMax         = 8
	PedCash        = 5
	FlattenSpeed   = 20.0
	PedWanderSpeed = 18.0

	JerseyHP        = 8.0
	JerseyBrittleHP = 1.0
	DumpHP          = 12.0
	DumpCash        = 15
	ConcreteSetHP   = 28.0
	ConcreteUnsetHP = 3.0 // plant boon: shoveable like a dump

	BannerLife  = 1.2
	DollarLife  = 0.7
	DollarRise  = 20.0
	TallyRoll   = 1.4
	ShakeWreck  = 3.5
	ShakePeel   = 5.0
	GlanceScale = -0.3 // blade-up body vs building
	RubbleInset = 2.0
	DeepRubbleA = 16.0 * 16.0

	TillerDeadzone = 8.0
	TapMoveMax     = 12.0
	TapDurMax      = 0.200 // seconds
	ThrottleTravel = 40.0
)
