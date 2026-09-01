package game

import (
	"time"

	"permitdenied/internal/attach"
	"permitdenied/internal/capitol"
	"permitdenied/internal/city"
	"permitdenied/internal/lot"
	"permitdenied/internal/mapgen"
	"permitdenied/internal/meta"
	"permitdenied/internal/threats"
	"permitdenied/internal/town"
)

func (g *Game) isPlaying() bool {
	switch g.scene {
	case ScenePlay, SceneTown, SceneCity, SceneCapitol:
		return true
	default:
		return false
	}
}

func (g *Game) worldW() float64 {
	if g.mapW > 0 {
		return g.mapW
	}
	return LotW
}

func (g *Game) worldH() float64 {
	if g.mapH > 0 {
		return g.mapH
	}
	return LotH
}

func (g *Game) clockLimit() float64 {
	if g.clockSec > 0 {
		return g.clockSec
	}
	return RunSeconds
}

func (g *Game) buzzerTick() int {
	return int(g.clockLimit() * TPS)
}

func (g *Game) LoadProgress() {
	s, err := meta.Load(meta.DefaultPath())
	if err != nil {
		return
	}
	g.progress = s
}

func (g *Game) persistProgress() {
	path := g.metaPath
	if path == "" {
		path = meta.DefaultPath()
	}
	_ = meta.SaveFile(path, g.progress)
}

func (g *Game) applyStartKit() {
	g.kit = attach.Kit{}
	if g.progress.Has(meta.Ripper) {
		g.kit.Ripper = true
	}
	if g.progress.Has(meta.Wide) {
		g.kit.Wide = true
	}
	if g.progress.Has(meta.Ball) {
		g.kit.Ball = true
	}
	if g.progress.ArmorTier >= 1 {
		g.dozer.Plates += MetaArmorPlates
	}
}

func (g *Game) loadCounty() {
	g.scene = ScenePlay
	g.run.Tier = 0
	g.mapW, g.mapH = LotW, LotH
	g.clockSec = RunSeconds
	g.lot = lot.New(WreckHPPerTile)
	g.dozer.X, g.dozer.Y = SpawnX, SpawnY
	g.dozer.Heading = 0
	g.dozer.Speed = 0
	g.cruisers = nil
	g.blockers = initialBlockers(true)
	g.heavies = nil
	g.pickups = nil
	g.excavator = threats.Excavator{}
	g.chopper = threats.Chopper{X: 320, Y: 40, SpotR: ChopperSpotR}
	g.peds = spawnPeds()
	g.streets = nil
	g.procedural = false
	g.tower = capitol.Tower{}
	g.beat = struct {
		cruisers0, blockers, chopper, exAnn, exArr, concrete, twoFam bool
	}{}
	g.wave = struct{ t30, t80, t140 bool }{}
	g.resetMapClock()
}

func (g *Game) applyArena(a mapgen.Arena, scene Scene, tier int) {
	g.scene = scene
	g.run.Tier = tier
	g.mapW, g.mapH = a.W, a.H
	g.clockSec = a.Clock
	g.lot = lot.Lot{Buildings: mapgen.Buildings(a, WreckHPPerTile)}
	g.dozer.X, g.dozer.Y = a.SpawnX, a.SpawnY
	g.dozer.Heading = 0
	g.dozer.Speed = 0
	g.cruisers = nil
	g.blockers = nil
	g.heavies = nil
	g.excavator = threats.Excavator{}
	g.chopper = threats.Chopper{}
	g.peds = nil
	g.streets = a.Streets
	g.procedural = true
	g.tower = capitol.Tower{}
	g.beat = struct {
		cruisers0, blockers, chopper, exAnn, exArr, concrete, twoFam bool
	}{}
	g.wave = struct{ t30, t80, t140 bool }{}
	g.resetMapClock()
}

func (g *Game) resetMapClock() {
	g.run.Tick = 0
	g.run.PressureBonus = 0
	g.stallTicks = 0
	g.buryTicks = 0
	g.glanceLatch = false
	g.jerseyLatch = map[int]struct{}{}
}

func (g *Game) enterTown() {
	alt := g.progress.Has(meta.MapPool)
	a := town.New(g.run.Seed+11, alt)
	g.applyArena(a, SceneTown, 1)
	g.blockers = []threats.Blocker{
		threats.Jersey(560, 920, 80, 16, JerseyHP),
		threats.Jersey(640, 640, 64, 16, JerseyHP),
		threats.Jersey(600, 360, 80, 16, JerseyHP),
	}
	g.heavies = []threats.Heavy{
		threats.FireTruck(200, 640, HeavyFireW, HeavyFireH, HeavyFireHP),
	}
	g.spawnCruisersAt([][2]float64{{600, 1000}, {680, 1000}, {640, 800}})
	g.placePickups(attach.Ripper, attach.WideBlade)
	g.fx.SetBanner("PERMIT: TOWN ASSESSMENT", BannerLife)
	g.progress.NoteTier(2)
}

func (g *Game) enterCity() {
	alt := g.progress.Has(meta.MapPool)
	a := city.New(g.run.Seed+29, alt)
	g.applyArena(a, SceneCity, 2)
	g.blockers = []threats.Blocker{
		threats.Jersey(600, 960, 72, 16, JerseyHP),
		threats.Jersey(200, 720, 48, 16, JerseyHP),
		threats.Concrete(280, 480, 80, 24),
	}
	g.heavies = []threats.Heavy{
		threats.Wagon(240, 960, HeavyWagonW, HeavyWagonH, HeavyWagonHP),
		threats.Wagon(1040, 720, HeavyWagonW, HeavyWagonH, HeavyWagonHP),
		threats.FireTruck(640, 480, HeavyFireW, HeavyFireH, HeavyFireHP),
	}
	g.heavies[0].Parked = true
	g.spawnCruisersAt([][2]float64{{600, 1100}, {680, 1100}, {320, 800}, {900, 800}})
	g.chopper = threats.Chopper{X: 640, Y: 80, SpotR: ChopperSpotR, Active: true}
	g.placePickups(attach.WreckingBall, attach.PileDriver)
	g.fx.SetBanner("PERMIT: CITY DISTRICT", BannerLife)
	g.progress.NoteTier(3)
}

func (g *Game) enterCapitol() {
	a := capitol.Template()
	g.applyArena(a, SceneCapitol, 3)
	g.tower = capitol.Index(g.lot)
	g.blockers = []threats.Blocker{
		threats.Jersey(400, 1400, 80, 16, JerseyHP),
		threats.Concrete(400, 1280, 96, 24),
		threats.Dump(200, 1280, 40, 24, DumpHP),
	}
	g.heavies = []threats.Heavy{
		threats.FireTruck(200, 1100, HeavyFireW, HeavyFireH, HeavyFireHP),
		threats.Wagon(760, 1100, HeavyWagonW, HeavyWagonH, HeavyWagonHP),
	}
	g.spawnCruisersAt([][2]float64{{440, 1400}, {520, 1400}, {400, 1100}})
	g.chopper = threats.Chopper{X: 480, Y: 40, SpotR: ChopperSpotR, Active: true}
	g.placeExcavatorAt(480, 1000)
	g.placePickups(attach.PileDriver, attach.ExtraPlate)
	g.fx.SetBanner("PERMIT: CAPITOL CAMPUS", BannerLife)
	g.progress.NoteTier(4)
}

func (g *Game) spawnCruisersAt(spots [][2]float64) {
	for _, s := range spots {
		if len(g.cruisers) >= CruiserCap {
			return
		}
		g.cruisers = append(g.cruisers, threats.SpawnCruiser(s[0], s[1]))
	}
}

func (g *Game) placeExcavatorAt(x, y float64) {
	g.excavator = threats.Excavator{
		X: x, Y: y,
		Heading:   3.141592653589793,
		Announced: true,
		Arrived:   true,
		Alive:     true,
		HP:        ExcavatorHP,
		Swinging:  true,
	}
}

func (g *Game) placePickups(kinds ...attach.Kind) {
	g.pickups = nil
	spots := [][2]float64{
		{g.mapW * 0.25, g.mapH * 0.72},
		{g.mapW * 0.75, g.mapH * 0.55},
	}
	for i, k := range kinds {
		if i >= len(spots) {
			break
		}
		if g.kit.Has(k) && k != attach.ExtraPlate {
			continue
		}
		g.pickups = append(g.pickups, attach.Spawn(k, spots[i][0], spots[i][1], PickupSize))
	}
}

func (g *Game) mapCleared() bool {
	switch g.scene {
	case ScenePlay:
		return g.run.SheriffDown && g.run.YardDown && g.run.PlantDown
	case SceneTown, SceneCity:
		down, total := g.lot.NamedDown()
		return total > 0 && down == total
	case SceneCapitol:
		return g.tower.Cleared(g.lot)
	default:
		return false
	}
}

func (g *Game) advanceMap() {
	g.run.MapsCleared++
	if id := g.progress.GrantNext(); id != "" {
		g.newUnlocks = append(g.newUnlocks, id)
	}
	switch g.scene {
	case ScenePlay:
		g.enterTown()
	case SceneTown:
		g.enterCity()
	case SceneCity:
		g.enterCapitol()
	case SceneCapitol:
		g.endRun("cleared")
	}
}

func (g *Game) leaveResult() {
	total := g.run.Final(Mult(g.run.Targets))
	g.progress.NoteCash(total)
	if g.run.Death == "cleared" {
		g.progress.NoteTier(4)
		if id := g.progress.GrantNext(); id != "" {
			g.newUnlocks = append(g.newUnlocks, id)
		}
	}
	g.persistProgress()
	if len(g.newUnlocks) > 0 {
		g.scene = SceneMeta
		return
	}
	g.scene = SceneTitle
	g.audio.Stop()
}

func (g *Game) freshSeed() int64 {
	return time.Now().UnixNano()
}
