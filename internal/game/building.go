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
	producingType      UnitType
	productionTime     float64
	team               int
	faction            FactionType
	rallyPointX        float64
	rallyPointY        float64
	isSelected         bool
}

func NewFactionBuilding(x, y float64, team int, faction FactionType) *Building {
	return &Building{
		x:                  x,
		y:                  y,
		width:              64,
		height:             64,
		isBuilding:         false,
		buildProgress:      0,
		productionProgress: 0,
		producingType:      UnitTypeWorker,
		productionTime:     4.0,
		team:               team,
		faction:            faction,
		rallyPointX:        x + 64 + 10,
		rallyPointY:        y + 32,
	}
}

func NewBuilding(x, y float64) *Building {
	return NewFactionBuilding(x, y, 1, FactionUSA)
}

func (b *Building) BoundingBox() image.Rectangle {
	return image.Rect(int(b.x), int(b.y), int(b.x+b.width), int(b.y+b.height))
}

func (b *Building) Draw(screen *ebiten.Image, camX, camY, zoom float64) {
	screenX := float32((b.x - camX) * zoom)
	screenY := float32((b.y - camY) * zoom)
	bw := float32(b.width * zoom)
	bh := float32(b.height * zoom)

	// Viewport culling
	if screenX+bw < 0 || screenX > float32(ViewWidth) || screenY+bh < 0 || screenY > float32(ViewHeight) {
		return
	}

	// Render Sprite
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(zoom, zoom)
	op.GeoM.Translate(float64(screenX), float64(screenY))

	// If building is under construction, render semi-transparent
	if b.buildProgress < 1.0 {
		op.ColorScale.Scale(1, 1, 1, 0.5)
	}
	screen.DrawImage(ImgBuilding, op)

	// Draw Unit Production Bar (Green)
	if b.isBuilding {
		barWidth := bw * float32(b.productionProgress)
		barColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}
		vector.DrawFilledRect(screen, screenX, screenY+bh+3, barWidth, 4, barColor, false)
	}

	// Draw Construction Progress Bar (Yellow, only if under construction)
	if b.buildProgress < 1.0 && b.buildProgress > 0 {
		barWidth := bw * float32(b.buildProgress)
		barColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}
		vector.DrawFilledRect(screen, screenX, screenY-8, barWidth, 4, barColor, false)
	}

	// Draw selection border
	if b.isSelected {
		vector.StrokeRect(screen, screenX, screenY, bw, bh, 2, color.RGBA{0, 255, 0, 255}, false)
	}
}
