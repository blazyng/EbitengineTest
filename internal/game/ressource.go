package game

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type ResourceNode struct {
	x, y   float64
	width  float64
	height float64
	amount int
}

func NewResourceNode(x, y float64, amount int) *ResourceNode {
	return &ResourceNode{
		x:      x,
		y:      y,
		width:  32,
		height: 32,
		amount: amount,
	}
}

func (r *ResourceNode) BoundingBox() image.Rectangle {
	return image.Rect(int(r.x), int(r.y), int(r.x+r.width), int(r.y+r.height))
}

func (r *ResourceNode) Draw(screen *ebiten.Image, camX, camY, zoom float64) {
	screenX := float32((r.x - camX) * zoom)
	screenY := float32((r.y - camY) * zoom)
	rw := float32(r.width * zoom)
	rh := float32(r.height * zoom)

	// Viewport culling
	if screenX+rw < 0 || screenX > float32(ViewWidth) || screenY+rh < 0 || screenY > float32(ViewHeight) {
		return
	}

	fillColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	vector.DrawFilledRect(screen, screenX, screenY, rw, rh, fillColor, false)

	// Draw remaining amount text above the node
	amountText := fmt.Sprintf("%d", r.amount)
	ebitenutil.DebugPrintAt(screen, amountText, int(screenX), int(screenY)-12)
}
