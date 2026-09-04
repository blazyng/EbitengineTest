package main

import (
	"log"

	"github.com/blazyng/EbitengineTest/internal/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Load assets (images) before starting the game
	game.LoadAssets()

	// Initialize the game state
	g, err := game.NewGame()
	if err != nil {
		log.Fatalf("Could not initialize game: %v", err)
	}

	// Set window size to 1280x960 (4x retro scale) and allow user to resize/maximize freely
	ebiten.SetWindowSize(1280, 960)
	ebiten.SetWindowTitle("Mubi RTS")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
