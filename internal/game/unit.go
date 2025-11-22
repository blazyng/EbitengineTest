package game

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// UnitState defines what the unit is currently doing
type UnitState int

const (
	StateIdle            UnitState = iota // 0
	StateMoving                           // 1
	StateMovingToHarvest                  // 2
	StateHarvesting                       // 3
	StateReturning                        // 4
	StateAttacking                        // 5
	StateMovingToBuild                    // 6
	StateBuilding                         // 7
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

// Draw draws JUST THIS UNIT
func (u *Unit) Draw(screen *ebiten.Image, camX, camY float64) {
	// Calculate screen position
	screenX := float32(u.x - camX)
	screenY := float32(u.y - camY)

	var unitColor color.RGBA

	// Set color based on team
	if u.team == 1 {
		unitColor = color.RGBA{0, 0, 255, 255} // Blue (Player)
	} else {
		unitColor = color.RGBA{150, 0, 150, 255} // Purple (Enemy)
	}

	// Yellow if carrying resources
	if u.cargo > 0 {
		unitColor = color.RGBA{255, 255, 0, 255}
	}

	// Draw the unit body
	vector.DrawFilledRect(screen, screenX, screenY, float32(unitSize), float32(unitSize), unitColor, false)

	// Draw selection border (only for player units)
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
