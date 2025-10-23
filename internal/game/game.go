// internal/game/game.go
package game

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil" // For debug text
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

	// 1. Handle Input
	g.handleInput(mouseX, mouseY) //  in input.go
	g.handleProductionInput()     // in input.go

	// 2. Update World State
	g.updateUnits(g.units, g.enemyUnits) // in update.go
	g.updateUnits(g.enemyUnits, g.units) // in update.go
	g.updateBuildings()                  // in update.go

	// 3. Cleanup
	g.cleanupDeadUnits() // in update.go

	return nil
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
