// internal/game/building.go
package game

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Building represents a structure
type Building struct {
	x, y               float64
	width, height      float64
	isBuilding         bool    // Is it currently producing something?
	buildProgress      float64 // 0.0 to 1.0
	rallyPointX        float64 // Where units spawn
	rallyPointY        float64
	productionProgress float64 //
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
// internal/game/building.go

func (b *Building) Draw(screen *ebiten.Image, camX, camY float64) {
	screenX := float32(b.x - camX)
	screenY := float32(b.y - camY)
	// 1. Das Gebäude selbst zeichnen
	fillColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	// Wenn das Gebäude noch im Bau ist (Ghost oder Worker baut noch), machen wir es grau/transparent
	if b.buildProgress < 1.0 {
		fillColor = color.RGBA{R: 100, G: 100, B: 100, A: 200} // Grau
	}

	vector.DrawFilledRect(screen, screenX, screenY, float32(b.width), float32(b.height), fillColor, false)

	// 2. Balken für EINHEITEN-Produktion (Der fehlende grüne Balken!)
	if b.isBuilding {
		// WICHTIG: Hier nutzen wir jetzt 'productionProgress'
		barWidth := float32(b.width) * float32(b.productionProgress)
		barColor := color.RGBA{R: 0, G: 255, B: 0, A: 255} // Grün

		// Zeichne Balken UNTER das Gebäude
		vector.DrawFilledRect(screen, screenX, screenY+float32(b.height)+5, barWidth, 5, barColor, false)
	}

	// 3. (Optional) Balken für GEBÄUDE-Baufortschritt (Wenn ein Worker es baut)
	if b.buildProgress < 1.0 && b.buildProgress > 0 {
		barWidth := float32(b.width) * float32(b.buildProgress)
		barColor := color.RGBA{R: 255, G: 255, B: 0, A: 255} // Gelb
		// Zeichne Balken ÜBER das Gebäude
		vector.DrawFilledRect(screen, float32(b.x), float32(b.y-10), barWidth, 5, barColor, false)
	}
}

// BoundingBox returns the building's collision rectangle
func (b *Building) BoundingBox() image.Rectangle {
	return image.Rect(int(b.x), int(b.y), int(b.x+b.width), int(b.y+b.height))
}
