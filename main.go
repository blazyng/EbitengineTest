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

	ebiten.SetWindowSize(1000, 1000)
	ebiten.SetWindowTitle("Mubi RTS")

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
