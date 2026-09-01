package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"permitdenied/internal/game"
)

func main() {
	ebiten.SetWindowSize(1280, 896)
	ebiten.SetWindowTitle("PERMIT DENIED")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	g := game.New()
	g.LoadProgress()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
