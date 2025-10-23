// internal/game/unit.go
package game

import "image"

// UnitState defines what the unit is currently doing
type UnitState int

const (
	StateIdle            UnitState = iota // 0
	StateMoving                           // 1
	StateMovingToHarvest                  // 2
	StateHarvesting                       // 3
	StateReturning                        // 4
	StateAttacking                        // 5 (New)
)

// Unit struct is now more complex
type Unit struct {
	x, y             float64
	targetX, targetY float64
	speed            float64
	isSelected       bool

	// State & Resources
	state        UnitState
	targetNode   *ResourceNode
	cargo        int
	harvestTimer float64

	// --- New Fields for Combat ---
	team         int // e.g., 1 for player, 2 for enemy
	health       int
	maxHealth    int
	attackDamage int
	attackRange  float64
	attackSpeed  float64 // Attacks per second (e.g., 1.0)
	attackTimer  float64 // Cooldown timer
	targetEnemy  *Unit   // The enemy it's currently attacking
}

// Helper constants for the unit
const (
	unitSize        = 32.0
	unitHarvestTime = 3.0
	unitCargoSize   = 10
)

// NewUnit creates a basic "worker" unit
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
		attackRange:  40.0, // Slightly more than its own size
		attackSpeed:  1.0,  // 1 attack per second
		attackTimer:  0.0,
	}
}

// BoundingBox returns the unit's collision rectangle
func (u *Unit) BoundingBox() image.Rectangle {
	return image.Rect(int(u.x), int(u.y), int(u.x+unitSize), int(u.y+unitSize))
}
