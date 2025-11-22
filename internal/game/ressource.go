package game

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ResourceNode represents a collectable resource on the map
type ResourceNode struct {
	x, y   float64
	width  float64
	height float64
	amount int
}

// NewResourceNode creates a new resource
func NewResourceNode(x, y float64, amount int) *ResourceNode {
	return &ResourceNode{
		x:      x,
		y:      y,
		width:  32,
		height: 32,
		amount: amount,
	}
}

// BoundingBox returns the resource's collision rectangle
func (r *ResourceNode) BoundingBox() image.Rectangle {
	return image.Rect(int(r.x), int(r.y), int(r.x+r.width), int(r.y+r.height))
}

// Draw draws the resource node taking camera position into account
func (r *ResourceNode) Draw(screen *ebiten.Image, camX, camY float64) {
	fillColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	// Calculate screen position based on camera
	screenX := float32(r.x - camX)
	screenY := float32(r.y - camY)

	vector.DrawFilledRect(screen, screenX, screenY, float32(r.width), float32(r.height), fillColor, false)
}
