package main

import (
	"log"

	"github.com/blazyng/EbitengineTest/internal/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// "Konstructor"-Function
	g, err := game.NewGame()
	if err != nil {
		log.Fatalf("Konnte das Spiel nicht initialisieren: %v", err)
	}

	ebiten.SetWindowSize(1000, 1000)
	ebiten.SetWindowTitle("Mubi")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
