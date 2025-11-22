package game

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Building struct {
	x, y               float64
	width, height      float64
	isBuilding         bool    // Is currently producing a unit?
	buildProgress      float64 // Construction progress (0.0 - 1.0)
	productionProgress float64 // Unit production progress (0.0 - 1.0)
	rallyPointX        float64
	rallyPointY        float64
}

func NewBuilding(x, y float64) *Building {
	return &Building{
		x:                  x,
		y:                  y,
		width:              64,
		height:             64,
		isBuilding:         false,
		buildProgress:      0,
		productionProgress: 0,
		rallyPointX:        x + 64 + 10,
		rallyPointY:        y + 32,
	}
}

func (b *Building) BoundingBox() image.Rectangle {
	return image.Rect(int(b.x), int(b.y), int(b.x+b.width), int(b.y+b.height))
}

func (b *Building) Draw(screen *ebiten.Image, camX, camY float64) {
	screenX := float32(b.x - camX)
	screenY := float32(b.y - camY)

	// Render Sprite
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(screenX), float64(screenY))

	// If building is under construction, render semi-transparent
	if b.buildProgress < 1.0 {
		op.ColorScale.Scale(1, 1, 1, 0.5)
	}
	screen.DrawImage(ImgBuilding, op)

	// Draw Unit Production Bar (Green)
	if b.isBuilding {
		barWidth := float32(b.width) * float32(b.productionProgress)
		barColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}
		vector.DrawFilledRect(screen, screenX, screenY+float32(b.height)+5, barWidth, 5, barColor, false)
	}

	// Draw Construction Progress Bar (Yellow, only if under construction)
	if b.buildProgress < 1.0 && b.buildProgress > 0 {
		barWidth := float32(b.width) * float32(b.buildProgress)
		barColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}
		vector.DrawFilledRect(screen, screenX, screenY-10, barWidth, 5, barColor, false)
	}
}
