package game

func (g *Game) camera() (camX, camY float64) {
	maxX := g.worldW() - ScreenW
	maxY := g.worldH() - ScreenH
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	camX = clamp(g.dozer.X-ScreenW/2, 0, maxX)
	camY = clamp(g.dozer.Y-ScreenH/2, 0, maxY)
	return
}
