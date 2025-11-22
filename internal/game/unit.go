package game

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type UnitState int

const (
	StateIdle UnitState = iota
	StateMoving
	StateMovingToHarvest
	StateHarvesting
	StateReturning
	StateAttacking
	StateMovingToBuild
	StateBuilding
)

type Unit struct {
	x, y             float64
	targetX, targetY float64
	speed            float64
	isSelected       bool

	state        UnitState
	targetNode   *ResourceNode
	cargo        int
	harvestTimer float64

	team           int
	health         int
	maxHealth      int
	attackDamage   int
	attackRange    float64
	attackSpeed    float64
	attackTimer    float64
	targetEnemy    *Unit
	targetBuilding *Building
}

const (
	unitSize        = 32.0
	unitHarvestTime = 3.0
	unitCargoSize   = 10
)

func NewUnit(x, y float64, team int) *Unit {
	return &Unit{
		x:            x,
		y:            y,
		targetX:      x,
		targetY:      y,
		speed:        2.0,
		state:        StateIdle,
		team:         team,
		health:       100,
		maxHealth:    100,
		attackDamage: 10,
		attackRange:  40.0,
		attackSpeed:  1.0,
		attackTimer:  0.0,
	}
}

func (u *Unit) BoundingBox() image.Rectangle {
	return image.Rect(int(u.x), int(u.y), int(u.x+unitSize), int(u.y+unitSize))
}

// Draw renders the unit sprite relative to the camera
func (u *Unit) Draw(screen *ebiten.Image, camX, camY float64) {
	screenX := float32(u.x - camX)
	screenY := float32(u.y - camY)

	// Render Sprite
	op := &ebiten.DrawImageOptions{}

	// Tinting logic using ColorScale
	if u.cargo > 0 {
		// Carrying Gold -> Yellow tint
		op.ColorScale.Scale(1, 1, 0, 1)
	} else if u.team == 2 {
		// Enemy -> Red tint
		op.ColorScale.Scale(1, 0.5, 0.5, 1)
	} else {
		// Player -> Normal
		op.ColorScale.Scale(1, 1, 1, 1)
	}

	op.GeoM.Translate(float64(screenX), float64(screenY))
	screen.DrawImage(ImgUnit, op)

	// Draw selection border (player only)
	if u.isSelected && u.team == 1 {
		vector.StrokeRect(screen, screenX, screenY, float32(unitSize), float32(unitSize), 2, color.RGBA{0, 255, 0, 255}, false)
	}

	// Draw Health Bar
	if u.health < u.maxHealth {
		healthPercent := float32(u.health) / float32(u.maxHealth)
		barWidth := float32(unitSize)

		// Red background
		vector.DrawFilledRect(screen, screenX, screenY-7, barWidth, 5, color.RGBA{255, 0, 0, 255}, false)
		// Green foreground
		vector.DrawFilledRect(screen, screenX, screenY-7, barWidth*healthPercent, 5, color.RGBA{0, 255, 0, 255}, false)
	}
}
