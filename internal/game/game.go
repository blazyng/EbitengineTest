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
	units           []*Unit // Player's units
	enemyUnits      []*Unit // New: Enemy units
	resourceNodes   []*ResourceNode
	barracks        *Building
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
		playerResources: 100,
	}

	// Add starting units (Player, Team 1)
	g.units = []*Unit{
		NewUnit(50, 50, 1), // Use our new constructor
		NewUnit(60, 60, 1),
	}

	// New: Add enemy units (Enemy, Team 2)
	g.enemyUnits = []*Unit{
		NewUnit(250, 100, 2),
		NewUnit(250, 140, 2),
	}

	// Add a resource node
	g.resourceNodes = []*ResourceNode{
		NewResourceNode(200, 200, 1000),
	}

	// Add a barracks
	g.barracks = NewBuilding(10, 100)

	return g, nil
}

// distance is a helper function
func distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()

	g.handleInput(mouseX, mouseY)
	g.handleProductionInput()

	// Update all units (player and enemy)
	g.updateUnits(g.units, g.enemyUnits) // Pass enemy list
	g.updateUnits(g.enemyUnits, g.units) // Pass player list

	g.updateBuildings()

	// New: Clean up dead units
	g.cleanupDeadUnits()

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

// updateBuildings needs one small tweak:
func (g *Game) updateBuildings() {
	if g.barracks.isBuilding {
		dt := 1.0 / float64(ebiten.TPS())
		g.barracks.buildProgress += dt / unitBuildTime

		if g.barracks.buildProgress >= 1.0 {
			g.barracks.isBuilding = false
			g.barracks.buildProgress = 0.0

			// Create a new unit (Team 1)
			newUnit := NewUnit(g.barracks.rallyPointX, g.barracks.rallyPointY, 1) // Team 1
			g.units = append(g.units, newUnit)
		}
	}
}

// handleInput processes user mouse clicks
func (g *Game) handleInput(mouseX, mouseY int) {
	// --- Right Click (Commands) ---
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		clickedNode := false
		clickedEnemy := false

		// 1. Check if we clicked an enemy
		for _, enemy := range g.enemyUnits {
			if float64(mouseX) >= enemy.x && float64(mouseX) <= enemy.x+unitSize &&
				float64(mouseY) >= enemy.y && float64(mouseY) <= enemy.y+unitSize {

				// Clicked on an enemy! Send all selected units to attack.
				for _, unit := range g.units {
					if unit.isSelected {
						unit.state = StateAttacking
						unit.targetEnemy = enemy
						unit.targetNode = nil // Clear resource target
					}
				}
				clickedEnemy = true
				break
			}
		}
		if clickedEnemy {
			return // Don't process other right-click actions
		}
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
func (g *Game) updateUnits(units, enemies []*Unit) {
	dt := 1.0 / float64(ebiten.TPS())

	for _, unit := range units {
		// Update attack cooldown
		if unit.attackTimer > 0 {
			unit.attackTimer -= dt
		}

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
		case StateAttacking:
			if unit.targetEnemy == nil || unit.targetEnemy.health <= 0 {
				// Target is dead or gone
				unit.state = StateIdle
				unit.targetEnemy = nil
				continue
			}

			// Calculate distance to target
			dist := distance(unit.x, unit.y, unit.targetEnemy.x, unit.targetEnemy.y)

			if dist <= unit.attackRange {
				// 1. In range: Stop moving and attack
				// Stop moving

				if unit.attackTimer <= 0 {
					// Attack is ready
					unit.targetEnemy.health -= unit.attackDamage
					unit.attackTimer = 1.0 / unit.attackSpeed // Reset cooldown
				}
			} else {
				// 2. Out of range: Chase the target
				dx := unit.targetEnemy.x - unit.x
				dy := unit.targetEnemy.y - unit.y
				unit.x += (dx / dist) * unit.speed
				unit.y += (dy / dist) * unit.speed
			}

		}
	}
}

// New function to remove dead units
func (g *Game) cleanupDeadUnits() {
	// Create new slices for living units
	livingPlayerUnits := make([]*Unit, 0, len(g.units))
	for _, unit := range g.units {
		if unit.health > 0 {
			livingPlayerUnits = append(livingPlayerUnits, unit)
		} else {
			// If a selected unit dies, we need to handle that,
			// but for now, just remove it.
		}
	}
	g.units = livingPlayerUnits // Replace old slice

	livingEnemyUnits := make([]*Unit, 0, len(g.enemyUnits))
	for _, unit := range g.enemyUnits {
		if unit.health > 0 {
			livingEnemyUnits = append(livingEnemyUnits, unit)
		}
	}
	g.enemyUnits = livingEnemyUnits // Replace old slice
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
	g.barracks.Draw(screen)
	//draw barracks

	allUnits := append(g.units, g.enemyUnits...)

	for _, unit := range allUnits {
		var unitColor color.RGBA

		// Set color based on team
		if unit.team == 1 {
			unitColor = color.RGBA{0, 0, 255, 255} // Blue (Player)
		} else {
			unitColor = color.RGBA{150, 0, 150, 255} // Purple (Enemy)
		}

		// Yellow if carrying resources
		if unit.cargo > 0 {
			unitColor = color.RGBA{255, 255, 0, 255}
		}

		vector.DrawFilledRect(screen, float32(unit.x), float32(unit.y), float32(unitSize), float32(unitSize), unitColor, false)

		// Draw selection border (only for player units)
		if unit.isSelected && unit.team == 1 {
			vector.StrokeRect(screen, float32(unit.x), float32(unit.y), float32(unitSize), float32(unitSize), 2, color.RGBA{0, 255, 0, 255}, false)
		}

		// New: Draw Health Bar
		if unit.health < unit.maxHealth {
			healthPercent := float32(unit.health) / float32(unit.maxHealth)
			barWidth := float32(unitSize)

			// Red background
			vector.DrawFilledRect(screen, float32(unit.x), float32(unit.y-7), barWidth, 5, color.RGBA{255, 0, 0, 255}, false)
			// Green foreground
			vector.DrawFilledRect(screen, float32(unit.x), float32(unit.y-7), barWidth*healthPercent, 5, color.RGBA{0, 255, 0, 255}, false)
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
