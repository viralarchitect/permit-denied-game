package game

import (
	"testing"

	"permitdenied/internal/dozerpack"
	"permitdenied/internal/lot"
)

func countySpawn(t testing.TB) (x, y, heading float64) {
	t.Helper()
	return mustCountySpawn()
}

func mustCountySpawn() (x, y, heading float64) {
	scenario, err := dozerpack.LoadEmbedded()
	if err != nil {
		panic(err)
	}
	return scenario.Spawn()
}

func poseSouthOf(g *Game, b *lot.Building, bladeDown bool) {
	cx, _ := b.Center()
	g.dozer.X = cx
	g.dozer.Y = b.Y + b.H + DozerBodyR + 1
	g.dozer.Heading = 0
	g.dozer.BladeDown = bladeDown
	g.dozer.Speed = 0
	g.dozer.Heat = 0
}

func firstCountyMundane(g *Game) *lot.Building {
	for i := range g.lot.Buildings {
		if g.lot.Buildings[i].ID != lot.TargetNone {
			continue
		}
		return &g.lot.Buildings[i]
	}
	return nil
}
