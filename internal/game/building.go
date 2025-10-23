// internal/game/building.go
package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Building represents a structure
type Building struct {
	x, y          float64
	width, height float64
	isBuilding    bool    // Is it currently producing something?
	buildProgress float64 // 0.0 to 1.0
	rallyPointX   float64 // Where units spawn
	rallyPointY   float64
}

// NewBuilding creates a new building
func NewBuilding(x, y float64) *Building {
	return &Building{
		x:             x,
		y:             y,
		width:         64, // Make it bigger than a unit
		height:        64,
		isBuilding:    false,
		buildProgress: 0,
		rallyPointX:   x + 64 + 10, // Spawn units to the right
		rallyPointY:   y + 32,
	}
}

// Draw draws the building
func (b *Building) Draw(screen *ebiten.Image) {
	// A simple red square for "Barracks"
	fillColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	vector.DrawFilledRect(screen, float32(b.x), float32(b.y), float32(b.width), float32(b.height), fillColor, false)

	// Draw build progress bar
	if b.isBuilding {
		barWidth := float32(b.width) * float32(b.buildProgress)
		barColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}
		vector.DrawFilledRect(screen, float32(b.x), float32(b.y+b.height+5), barWidth, 5, barColor, false)
	}
}
