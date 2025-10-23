// internal/game/resource.go
package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ResourceNode represents a collectable resource on the map
type ResourceNode struct {
	x, y   float64
	width  float64 // Use width/height for clicking
	height float64
	amount int // How much resource is left
}

// NewResourceNode creates a new resource
func NewResourceNode(x, y float64, amount int) *ResourceNode {
	return &ResourceNode{
		x:      x,
		y:      y,
		width:  32, // Let's make it 32x32 for now
		height: 32,
		amount: amount,
	}
}

// Draw draws the resource node
func (r *ResourceNode) Draw(screen *ebiten.Image) {
	// A simple yellow square for "gold"
	fillColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	vector.DrawFilledRect(screen, float32(r.x), float32(r.y), float32(r.width), float32(r.height), fillColor, false)
}
