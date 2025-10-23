// internal/game/unit.go
package game

// UnitState defines what the unit is currently doing
type UnitState int

const (
	StateIdle            UnitState = iota // 0
	StateMoving                           // 1
	StateMovingToHarvest                  // 2
	StateHarvesting                       // 3
	StateReturning                        // 4
)

// Unit struct is now more complex
type Unit struct {
	x          float64
	y          float64
	targetX    float64
	targetY    float64
	speed      float64
	isSelected bool

	// --- New Fields for State & Resources ---
	state        UnitState
	targetNode   *ResourceNode // The node it's currently targeting
	cargo        int           // How much resource the unit is carrying
	harvestTimer float64       // A countdown timer for harvesting
}

// Helper constants for the unit
const (
	unitSize        = 32.0 // Matched to your drawing
	unitHarvestTime = 3.0  // 3 seconds to harvest
	unitCargoSize   = 10   // Can carry 10 resources
)
