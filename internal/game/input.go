package game

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

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
