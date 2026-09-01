package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"permitdenied/internal/lot"
)

const WindowTitle = "PERMIT DENIED"

type Keys struct {
	Enter, Escape, F1, F2, F3 bool
}

type Snapshot struct {
	Title        string  `json:"title"`
	Scene        string  `json:"scene"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	Heading      float64 `json:"heading"`
	Speed        float64 `json:"speed"`
	BladeDown    bool    `json:"blade_down"`
	Stance       string  `json:"stance"`
	Plates       int     `json:"plates"`
	Heat         float64 `json:"heat"`
	Tick         int     `json:"tick"`
	Time         float64 `json:"time"`
	Clock        string  `json:"clock"`
	Death        string  `json:"death"`
	Over         bool    `json:"over"`
	Targets      int     `json:"targets"`
	StructCash   int     `json:"struct_cash"`
	VehicleCash  int     `json:"vehicle_cash"`
	SheriffHP    float64 `json:"sheriff_hp"`
	YardHP       float64 `json:"yard_hp"`
	PlantHP      float64 `json:"plant_hp"`
	Sheriff      string  `json:"sheriff"`
	Yard         string  `json:"yard"`
	Plant        string  `json:"plant"`
	Debug        bool    `json:"debug"`
	Banner       string  `json:"banner"`
	PIT          bool    `json:"pit"`
	Dumps        bool    `json:"dumps"`
	ConcreteSets bool    `json:"concrete_sets"`
}

type harnessFrame struct {
	in   Input
	keys Keys
}

func (g *Game) Silence() {
	g.audio = nil
}

func (g *Game) Drive(in Input, keys Keys) error {
	g.harness = &harnessFrame{in: in, keys: keys}
	err := g.Update()
	g.harness = nil
	return err
}

func (g *Game) Snapshot() Snapshot {
	s := Snapshot{
		Title: WindowTitle,
		Scene: g.sceneName(),
		Debug: g.debug,
	}
	if g.scene == SceneTitle {
		return s
	}
	s.X = g.dozer.X
	s.Y = g.dozer.Y
	s.Heading = g.dozer.Heading
	s.Speed = g.dozer.Speed
	s.BladeDown = g.dozer.BladeDown
	s.Stance = stanceLabel(g.dozer.BladeDown)
	s.Plates = g.dozer.Plates
	s.Heat = g.dozer.Heat
	s.Tick = g.run.Tick
	s.Time = g.run.Time()
	s.Clock = clockLabel(g.run.Tick)
	s.Death = g.run.Death
	s.Over = g.run.Over
	s.Targets = g.run.Targets
	s.StructCash = g.run.StructCash
	s.VehicleCash = g.run.VehicleCash
	s.Banner = g.fx.Banner
	s.PIT = g.run.CruiserPIT
	s.Dumps = g.run.DumpTrucks
	s.ConcreteSets = g.run.ConcreteSets
	if b := g.lot.BuildingByID(lot.TargetSheriff); b != nil {
		s.SheriffHP = b.HP
		s.Sheriff = buildingState(b.State)
	}
	if b := g.lot.BuildingByID(lot.TargetYard); b != nil {
		s.YardHP = b.HP
		s.Yard = buildingState(b.State)
	}
	if b := g.lot.BuildingByID(lot.TargetPlant); b != nil {
		s.PlantHP = b.HP
		s.Plant = buildingState(b.State)
	}
	return s
}

func (g *Game) sceneName() string {
	switch g.scene {
	case SceneTitle:
		return "title"
	case ScenePlay:
		return "play"
	case SceneTown:
		return "town"
	case SceneCity:
		return "city"
	case SceneCapitol:
		return "capitol"
	case SceneTally:
		return "tally"
	case SceneMeta:
		return "meta"
	default:
		return "unknown"
	}
}

func (g *Game) keyJust(k ebiten.Key) bool {
	if g.harness != nil {
		switch k {
		case ebiten.KeyEnter:
			return g.harness.keys.Enter
		case ebiten.KeyEscape:
			return g.harness.keys.Escape
		case ebiten.KeyF1:
			return g.harness.keys.F1
		case ebiten.KeyF2:
			return g.harness.keys.F2
		case ebiten.KeyF3:
			return g.harness.keys.F3
		case ebiten.KeySpace:
			return g.harness.in.BladeToggle
		}
		return false
	}
	return inpututil.IsKeyJustPressed(k)
}

func stanceLabel(down bool) string {
	if down {
		return "BLADE DOWN"
	}
	return "BLADE UP"
}

func clockLabel(tick int) string {
	elapsed := tick / TPS
	if elapsed < 0 {
		elapsed = 0
	}
	return fmt.Sprintf("%d:%02d", elapsed/60, elapsed%60)
}

func buildingState(s lot.BuildingState) string {
	switch s {
	case lot.Intact:
		return "intact"
	case lot.Cracked:
		return "cracked"
	case lot.InRubble:
		return "rubble"
	default:
		return "unknown"
	}
}
