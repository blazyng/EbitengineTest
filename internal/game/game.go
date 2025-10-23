// internal/game/game.go
package game

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil" // For debug text
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Game struct holds all game state
// Game struct holds all game state
type Game struct {
	units           []*Unit
	resourceNodes   []*ResourceNode
	barracks        *Building // New: Our one and only barracks
	playerResources int
	basePosition    image.Point

	// Selection fields
	isDragging bool
	dragStartX int
	dragStartY int
}

// --- Constants ---
const (
	unitCost      = 50  // Cost for a new unit
	unitBuildTime = 5.0 // 5 seconds to build a unit
)

// NewGame initializes the game
func NewGame() (*Game, error) {
	g := &Game{
		basePosition:    image.Point{X: 10, Y: 10},
		playerResources: 100, // Start with 100 to test building
	}
	// Add starting units
	g.units = []*Unit{
		{x: 50, y: 50, targetX: 50, targetY: 50, speed: 2.0, state: StateIdle},
		{x: 60, y: 60, targetX: 60, targetY: 60, speed: 2.0, state: StateIdle},
	}

	// Add a resource node
	g.resourceNodes = []*ResourceNode{
		NewResourceNode(200, 200, 1000), // A "gold mine" with 1000 resources
	}

	// New: Add a barracks
	g.barracks = NewBuilding(10, 100) // Place it at (10, 100)

	return g, nil
}

// distance is a helper function
func distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()

	// --- 1. Handle Input ---
	g.handleInput(mouseX, mouseY)

	// New: Handle production input
	g.handleProductionInput() // We'll create this function

	// --- 2. Update Game State ---
	g.updateUnits()
	g.updateBuildings() // We'll create this function

	return nil
}

// New function to handle building units
func (g *Game) handleProductionInput() {
	// If we press 'U' and have enough resources and are not already building
	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		if g.playerResources >= unitCost && !g.barracks.isBuilding {
			// Start building
			g.playerResources -= unitCost
			g.barracks.isBuilding = true
			g.barracks.buildProgress = 0.0
		}
	}
}

// New function to update building logic
func (g *Game) updateBuildings() {
	if g.barracks.isBuilding {
		dt := 1.0 / float64(ebiten.TPS())
		g.barracks.buildProgress += dt / unitBuildTime // dt / 5.0 seconds

		if g.barracks.buildProgress >= 1.0 {
			// Finished building!
			g.barracks.isBuilding = false
			g.barracks.buildProgress = 0.0

			// Create a new unit at the rally point
			newUnit := &Unit{
				x:       g.barracks.rallyPointX,
				y:       g.barracks.rallyPointY,
				targetX: g.barracks.rallyPointX,
				targetY: g.barracks.rallyPointY,
				speed:   2.0,
				state:   StateIdle,
			}
			g.units = append(g.units, newUnit)
		}
	}
}

// handleInput processes user mouse clicks
func (g *Game) handleInput(mouseX, mouseY int) {
	// --- Right Click (Commands) ---
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		// Check if we clicked on a resource node
		clickedNode := false
		for _, node := range g.resourceNodes {
			if float64(mouseX) >= node.x && float64(mouseX) <= node.x+node.width &&
				float64(mouseY) >= node.y && float64(mouseY) <= node.y+node.height {

				// Clicked on a node! Send all selected units to harvest it.
				for _, unit := range g.units {
					if unit.isSelected {
						unit.state = StateMovingToHarvest
						unit.targetNode = node
						unit.targetX = node.x // Target the node's position
						unit.targetY = node.y
					}
				}
				clickedNode = true
				break
			}
		}

		// If we didn't click a node, it's a normal move command
		if !clickedNode {
			for _, unit := range g.units {
				if unit.isSelected {
					unit.state = StateMoving
					unit.targetNode = nil // Not targeting a node
					unit.targetX = float64(mouseX)
					unit.targetY = float64(mouseY)
				}
			}
		}
	}

	// --- Left Click (Selection) ---
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.isDragging = true
		g.dragStartX, g.dragStartY = mouseX, mouseY

		unitClicked := false
		for _, unit := range g.units {
			if float64(mouseX) >= unit.x && float64(mouseX) <= unit.x+unitSize &&
				float64(mouseY) >= unit.y && float64(mouseY) <= unit.y+unitSize {
				unitClicked = true
				break
			}
		}

		// Deselect all units (unless holding Shift, but we'll add that later)
		for _, unit := range g.units {
			unit.isSelected = false
		}

		if unitClicked {
			for _, unit := range g.units {
				if float64(mouseX) >= unit.x && float64(mouseX) <= unit.x+unitSize &&
					float64(mouseY) >= unit.y && float64(mouseY) <= unit.y+unitSize {
					unit.isSelected = true
					break // Only select one
				}
			}
		}
	}

	// --- Drag Selection ---
	if g.isDragging {
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			g.isDragging = false
			selectionRect := image.Rect(g.dragStartX, g.dragStartY, mouseX, mouseY).Canon()

			for _, unit := range g.units {
				unitRect := image.Rect(int(unit.x), int(unit.y), int(unit.x)+int(unitSize), int(unit.y)+int(unitSize))
				if selectionRect.Overlaps(unitRect) {
					unit.isSelected = true
				}
			}
		}
	}
}

// updateUnits runs the state machine for every unit
func (g *Game) updateUnits() {
	// Calculate delta time (time since last frame)
	// This makes movement and timers independent of frame rate
	dt := 1.0 / float64(ebiten.TPS())

	for _, unit := range g.units {
		switch unit.state {

		case StateIdle:
			// Do nothing

		case StateMoving:
			// This is your original movement logic
			dist := distance(unit.x, unit.y, unit.targetX, unit.targetY)
			if dist > unit.speed {
				dx := unit.targetX - unit.x
				dy := unit.targetY - unit.y
				unit.x += (dx / dist) * unit.speed
				unit.y += (dy / dist) * unit.speed
			} else {
				unit.x = unit.targetX
				unit.y = unit.targetY
				unit.state = StateIdle // Arrived at destination
			}

		case StateMovingToHarvest:
			// Move towards the target resource node
			dist := distance(unit.x, unit.y, unit.targetNode.x, unit.targetNode.y)
			if dist > unit.speed {
				dx := unit.targetNode.x - unit.x
				dy := unit.targetNode.y - unit.y
				unit.x += (dx / dist) * unit.speed
				unit.y += (dy / dist) * unit.speed
			} else {
				// Arrived at the node
				unit.x = unit.targetNode.x
				unit.y = unit.targetNode.y
				unit.state = StateHarvesting
				unit.harvestTimer = unitHarvestTime // Start the harvest timer
			}

		case StateHarvesting:
			// Wait for the timer to finish
			unit.harvestTimer -= dt
			if unit.harvestTimer <= 0 {
				if unit.targetNode.amount > 0 {
					// Collect resources
					collected := unitCargoSize
					if unit.targetNode.amount < collected {
						collected = unit.targetNode.amount
					}

					unit.targetNode.amount -= collected
					unit.cargo = collected

					// Set target to base and go return
					unit.state = StateReturning
					unit.targetX = float64(g.basePosition.X)
					unit.targetY = float64(g.basePosition.Y)

					// If node is empty, clear target
					if unit.targetNode.amount <= 0 {
						unit.targetNode = nil
						// Note: We should also remove the node from g.resourceNodes
						// But we'll leave that for a later refactor
					}

				} else {
					// Node is empty, go idle
					unit.state = StateIdle
					unit.targetNode = nil
				}
			}

		case StateReturning:
			// Move towards the base
			dist := distance(unit.x, unit.y, unit.targetX, unit.targetY)
			if dist > unit.speed {
				dx := unit.targetX - unit.x
				dy := unit.targetY - unit.y
				unit.x += (dx / dist) * unit.speed
				unit.y += (dy / dist) * unit.speed
			} else {
				// Arrived at base
				unit.x = unit.targetX
				unit.y = unit.targetY

				// Drop off resources
				g.playerResources += unit.cargo
				unit.cargo = 0

				// Check if we should go back to harvesting
				if unit.targetNode != nil && unit.targetNode.amount > 0 {
					unit.state = StateMovingToHarvest
					unit.targetX = unit.targetNode.x
					unit.targetY = unit.targetNode.y
				} else {
					unit.state = StateIdle // Nothing to do
				}
			}
		}
	}
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Draw the base (a simple white square)
	vector.DrawFilledRect(screen, float32(g.basePosition.X), float32(g.basePosition.Y), 16, 16, color.White, false)

	// Draw all resource nodes
	for _, node := range g.resourceNodes {
		node.Draw(screen)
	}

	// Draw all units
	for _, unit := range g.units {
		// Base unit color (blue)
		unitColor := color.RGBA{0, 0, 255, 255}

		// Change color based on cargo
		if unit.cargo > 0 {
			unitColor = color.RGBA{255, 255, 0, 255} // Yellow if carrying
		}

		vector.DrawFilledRect(screen, float32(unit.x), float32(unit.y), float32(unitSize), float32(unitSize), unitColor, false)

		// Draw selection border
		if unit.isSelected {
			vector.StrokeRect(screen, float32(unit.x), float32(unit.y), float32(unitSize), float32(unitSize), 2, color.RGBA{0, 255, 0, 255}, false)
		}
	}

	// Draw selection box
	if g.isDragging {
		mouseX, mouseY := ebiten.CursorPosition()
		x, y := float32(g.dragStartX), float32(g.dragStartY)
		w, h := float32(mouseX-g.dragStartX), float32(mouseY-g.dragStartY)
		fillColor := color.RGBA{0, 255, 0, 50}
		vector.DrawFilledRect(screen, x, y, w, h, fillColor, false)
		strokeColor := color.RGBA{0, 255, 0, 255}
		vector.StrokeRect(screen, x, y, w, h, 1, strokeColor, false)
	}

	// Draw resource counter
	resText := fmt.Sprintf("Resources: %d", g.playerResources)
	ebitenutil.DebugPrint(screen, resText)

	// New: Add a build instruction hint
	hintText := "Press [U] to build Unit (50)"
	ebitenutil.DebugPrintAt(screen, hintText, 0, 10) // Draw below resource count
}

// Layout is Ebitengine's layout function
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}
