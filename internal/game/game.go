package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"permitdenied/internal/attach"
	"permitdenied/internal/audio"
	"permitdenied/internal/capitol"
	"permitdenied/internal/dozer"
	"permitdenied/internal/fx"
	"permitdenied/internal/lot"
	"permitdenied/internal/mapgen"
	"permitdenied/internal/meta"
	"permitdenied/internal/render"
	"permitdenied/internal/run"
	"permitdenied/internal/threats"
)

type Scene int

const (
	SceneTitle Scene = iota
	ScenePlay        // county
	SceneTown
	SceneCity
	SceneCapitol
	SceneTally
	SceneMeta
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
	heavies   []threats.Heavy
	pickups   []attach.Pickup
	kit       attach.Kit
	tower     capitol.Tower
	streets   []mapgen.Street
	progress  meta.Save
	metaPath  string
	progressLoadFailed bool

	mapW, mapH float64
	clockSec   float64
	procedural bool
	stallTicks int
	buryTicks  int
	newUnlocks []string

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
	wave struct {
		t30, t80, t140 bool
	}

	harness *harnessFrame

	glanceLatch bool
	jerseyLatch map[int]struct{}
}

func New() *Game {
	return &Game{
		scene:    SceneTitle,
		audio:    &audio.Audio{},
		mapW:     LotW,
		mapH:     LotH,
		clockSec: RunSeconds,
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.outsideW = outsideWidth
	g.outsideH = outsideHeight
	return ScreenW, ScreenH
}

func (g *Game) Update() error {
	in := g.readInput()
	if g.keyJust(ebiten.KeyF2) {
		g.debug = !g.debug
	}
	switch g.scene {
	case SceneTitle:
		if in.BladeToggle || in.Confirm {
			g.startRun()
			g.audio.StartChase()
		}
	case ScenePlay, SceneTown, SceneCity, SceneCapitol:
		if in.Back {
			g.persistProgress()
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
		if in.BladeToggle || in.Confirm {
			g.leaveResult()
			g.audio.Duck(false)
		}
	case SceneMeta:
		if in.BladeToggle || in.Confirm || in.Back {
			g.newUnlocks = nil
			g.scene = SceneTitle
			g.audio.Stop()
		}
	}
	return nil
}

func (g *Game) debugCheats() {
	if g.keyJust(ebiten.KeyF1) {
		g.peel("debug")
	}
	if g.keyJust(ebiten.KeyF3) {
		g.run.Tick += 15 * TPS
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.scene {
	case SceneTitle:
		render.DrawTitle(screen, g.progress.BestCash, g.progress.HighestTier)
	case SceneMeta:
		render.DrawMeta(screen, g.progress, g.newUnlocks)
	case ScenePlay, SceneTown, SceneCity, SceneCapitol, SceneTally:
		camX, camY := g.camera()
		sx, sy := g.fx.Offsets(g.run.Tick)
		down, total := g.lot.NamedDown()
		if g.scene == ScenePlay || (g.scene == SceneTally && g.run.Tier == 0) {
			down, total = g.run.Targets, 3
		}
		v := render.View{
			CamX: camX, CamY: camY, ShakeX: sx, ShakeY: sy,
			Tick:        g.run.Tick,
			Dozer:       g.dozer,
			Buildings:   g.lot.Buildings,
			Rubble:      g.lot.Rubble,
			Cruisers:    g.cruisers,
			Blockers:    g.blockers,
			Heavies:     g.heavies,
			Pickups:     g.pickups,
			Streets:     streetsAsRects(g.streets),
			Excavator:   g.excavator,
			Chopper:     g.chopper,
			Peds:        g.peds,
			Dollars:     g.fx.Dollars,
			Bursts:      g.fx.Bursts,
			Banner:      g.fx.Banner,
			BannerT:     g.fx.BannerT,
			RunTick:     g.run.Tick,
			StructCash:  g.run.StructCash,
			VehicleCash: g.run.VehicleCash,
			Targets:     g.run.Targets,
			NamedDown:   down,
			NamedTotal:  total,
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
			HideStance:  g.scene == SceneTally,
			MapW:        g.worldW(),
			MapH:        g.worldH(),
			Procedural:  g.procedural,
			Tier:        g.run.Tier,
			KitCount:    g.kit.Count(),
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

func streetsAsRects(in []mapgen.Street) []render.Rect {
	if len(in) == 0 {
		return nil
	}
	out := make([]render.Rect, len(in))
	for i, s := range in {
		out[i] = render.Rect{X: s.X, Y: s.Y, W: s.W, H: s.H}
	}
	return out
}

func (g *Game) startRun() {
	g.newUnlocks = nil
	g.run = run.New()
	g.run.Seed = g.freshSeed()
	g.dozer = dozer.Spawn(SpawnX, SpawnY)
	g.fx = fx.FX{}
	g.applyStartKit()
	g.loadCounty()
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
