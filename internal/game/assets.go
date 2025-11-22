package game

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Global variables for game assets
var (
	ImgGround   *ebiten.Image
	ImgUnit     *ebiten.Image
	ImgBuilding *ebiten.Image
)

// LoadAssets loads all image assets into memory once at startup
func LoadAssets() {
	var err error

	// 1. Load Ground Texture
	ImgGround, _, err = ebitenutil.NewImageFromFile("assets/ground.png")
	if err != nil {
		log.Fatal("Failed to load ground.png:", err)
	}

	// 2. Load Unit Sprite
	ImgUnit, _, err = ebitenutil.NewImageFromFile("assets/unit.png")
	if err != nil {
		log.Fatal("Failed to load unit.png:", err)
	}

	// 3. Load Building Sprite
	ImgBuilding, _, err = ebitenutil.NewImageFromFile("assets/building.png")
	if err != nil {
		log.Fatal("Failed to load building.png:", err)
	}
}
