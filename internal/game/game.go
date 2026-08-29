package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"permitdenied/internal/audio"
	"permitdenied/internal/dozer"
	"permitdenied/internal/fx"
	"permitdenied/internal/lot"
	"permitdenied/internal/render"
	"permitdenied/internal/run"
	"permitdenied/internal/threats"
)

type Scene int

const (
	SceneTitle Scene = iota
	ScenePlay
	SceneTally
)

type Game struct {
	scene Scene
	run   run.Run
	dozer dozer.Dozer
	lot   lot.Lot
	fx    fx.FX

	cruisers  []threats.Cruiser
	blockers  []threats.Blocker
	excavator threats.Excavator
	chopper   threats.Chopper
	peds      []threats.Ped

	outsideW, outsideH int
	debug              bool
	audio              *audio.Audio

	throttleID ebiten.TouchID
	throttleOn bool
	throttleY0 float64
	taps       map[ebiten.TouchID]tapInfo

	beat struct {
		cruisers0, blockers, chopper, exAnn, exArr, concrete, twoFam bool
	}
}

func New() *Game {
	return &Game{
		scene: SceneTitle,
		audio: &audio.Audio{},
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.outsideW = outsideWidth
	g.outsideH = outsideHeight
	return ScreenW, ScreenH
}

func (g *Game) Update() error {
	in := g.readInput()
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		g.debug = !g.debug
	}
	switch g.scene {
	case SceneTitle:
		if in.BladeToggle || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.startRun()
			g.audio.StartChase()
		}
	case ScenePlay:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.scene = SceneTitle
			g.audio.Stop()
			return nil
		}
		if g.fx.HitStop > 0 {
			g.fx.HitStop--
			g.run.Tick++
			return nil
		}
		g.debugCheats()
		g.stepPlay(in)
	case SceneTally:
		g.fx.TallyT += Dt
		g.audio.Duck(true)
		if in.BladeToggle || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.startRun()
			g.audio.StartChase()
			g.audio.Duck(false)
		}
	}
	return nil
}

func (g *Game) debugCheats() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.peel("debug")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		g.run.Tick += 15 * TPS
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.scene {
	case SceneTitle:
		render.DrawTitle(screen)
	case ScenePlay, SceneTally:
		camX, camY := g.camera()
		sx, sy := g.fx.Offsets(g.run.Tick)
		v := render.View{
			CamX: camX, CamY: camY, ShakeX: sx, ShakeY: sy,
			Tick:        g.run.Tick,
			Dozer:       g.dozer,
			Buildings:   g.lot.Buildings,
			Rubble:      g.lot.Rubble,
			Cruisers:    g.cruisers,
			Blockers:    g.blockers,
			Excavator:   g.excavator,
			Chopper:     g.chopper,
			Peds:        g.peds,
			Dollars:     g.fx.Dollars,
			Banner:      g.fx.Banner,
			BannerT:     g.fx.BannerT,
			RunTick:     g.run.Tick,
			StructCash:  g.run.StructCash,
			VehicleCash: g.run.VehicleCash,
			Targets:     g.run.Targets,
			PIT:         g.run.CruiserPIT,
			Dump:        g.run.DumpTrucks,
			Set:         g.run.ConcreteSets,
			YardDown:    g.run.YardDown,
			Heat:        g.dozer.Heat,
			Plates:      g.dozer.Plates,
			BladeDown:   g.dozer.BladeDown,
			Speed:       g.dozer.Speed,
			Debug:       g.debug,
			Time:        g.run.Time(),
		}
		render.DrawWorld(screen, v)
		render.DrawHUD(screen, v)
		if g.scene == SceneTally {
			render.DrawTally(screen, render.Tally{
				T:           g.fx.TallyT,
				Death:       g.run.Death,
				StructCash:  g.run.StructCash,
				VehicleCash: g.run.VehicleCash,
				Time:        g.run.TimeAlive,
				Targets:     g.run.Targets,
				Mult:        Mult(g.run.Targets),
				Total:       g.run.Final(Mult(g.run.Targets)),
			})
		}
	}
}

func (g *Game) startRun() {
	g.scene = ScenePlay
	g.run = run.New()
	g.dozer = dozer.Spawn(SpawnX, SpawnY)
	g.lot = lot.New(WreckHPPerTile)
	g.fx = fx.FX{}
	g.cruisers = nil
	g.blockers = initialBlockers(true)
	g.excavator = threats.Excavator{}
	g.chopper = threats.Chopper{X: 320, Y: 40, SpotR: ChopperSpotR}
	g.peds = spawnPeds()
	g.beat = struct {
		cruisers0, blockers, chopper, exAnn, exArr, concrete, twoFam bool
	}{}
}

func initialBlockers(dumpTrucks bool) []threats.Blocker {
	b := []threats.Blocker{
		threats.Jersey(400, 780, 48, 16, JerseyHP), // yard mouth choke
		threats.Concrete(256, 1100, 80, 24),
		threats.Concrete(280, 400, 80, 24),
		threats.Concrete(260, 200, 96, 32), // plant approach
	}
	if dumpTrucks {
		b = append(b,
			threats.Dump(456, 840, 40, 24, DumpHP),
			threats.Dump(520, 840, 40, 24, DumpHP),
		)
	}
	return b
}

func spawnPeds() []threats.Ped {
	spots := [][2]float64{
		{120, 1100}, {80, 980}, {160, 820}, {100, 600},
		{520, 1100}, {560, 920}, {540, 640}, {500, 400},
	}
	out := make([]threats.Ped, 0, PedMax)
	for i, s := range spots {
		if i >= PedMax {
			break
		}
		out = append(out, threats.Ped{
			X: s[0], Y: s[1], Alive: true,
			Heading: float64(i) * 0.7,
		})
	}
	return out
}
