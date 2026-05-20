package game

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	MapWidth  = 2000.0
	MapHeight = 2000.0
)

func (g *Game) updateCameraInput() {
	speed := 5.0

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.cameraX -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.cameraX += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.cameraY -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.cameraY += speed
	}

	// Clamp camera to map boundaries
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	if g.cameraY < 0 {
		g.cameraY = 0
	}
	if g.cameraX > MapWidth-320 {
		g.cameraX = MapWidth - 320
	}
	if g.cameraY > MapHeight-240 {
		g.cameraY = MapHeight - 240
	}
}

func (g *Game) handleInput(screenMouseX, screenMouseY int) {
	// Convert screen coordinates to world coordinates
	mouseX := float64(screenMouseX) + g.cameraX
	mouseY := float64(screenMouseY) + g.cameraY

	// --- Build Mode Logic ---
	if g.isPlacingBuilding {
		g.ghostX = mouseX
		g.ghostY = mouseY

		// 1. Check if placement is valid (collision check)
		canBuild := true
		ghostRect := image.Rect(int(g.ghostX), int(g.ghostY), int(g.ghostX)+64, int(g.ghostY)+64)

		for _, b := range g.buildings {
			if ghostRect.Overlaps(b.BoundingBox()) {
				canBuild = false
				break
			}
		}
		if canBuild {
			for _, res := range g.resourceNodes {
				if ghostRect.Overlaps(res.BoundingBox()) {
					canBuild = false
					break
				}
			}
		}
		g.canBuildHere = canBuild

		// 2. Left Click to Place
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if canBuild && g.playerResources >= 100 {
				g.playerResources -= 100

				newB := NewBuilding(g.ghostX, g.ghostY)
				newB.buildProgress = 0.0
				g.buildings = append(g.buildings, newB)

				// Order selected units to build
				for _, unit := range g.units {
					if unit.isSelected {
						unit.state = StateMovingToBuild
						unit.targetBuilding = newB
						unit.targetX = newB.x + newB.width/2
						unit.targetY = newB.y + newB.height/2
					}
				}
				g.isPlacingBuilding = false
			} else if !canBuild {
				fmt.Println("Cannot build here!")
			}
		}

		// Right Click to Cancel
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			g.isPlacingBuilding = false
		}
		return
	}

	// Toggle Build Mode with 'B'
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.isPlacingBuilding = true
		g.isDragging = false
	}

	// --- Right Click (Issue Commands) ---
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		clickedNode := false
		clickedEnemy := false

		// Calculate clamped target position to keep units inside the map
		targetX := mouseX
		targetY := mouseY
		if targetX < 0 {
			targetX = 0
		}
		if targetY < 0 {
			targetY = 0
		}
		if targetX > MapWidth-unitSize {
			targetX = MapWidth - unitSize
		}
		if targetY > MapHeight-unitSize {
			targetY = MapHeight - unitSize
		}

		// 1. Check if clicked an enemy (Attack)
		for _, enemy := range g.enemyUnits {
			if mouseX >= enemy.x && mouseX <= enemy.x+unitSize &&
				mouseY >= enemy.y && mouseY <= enemy.y+unitSize {

				for _, unit := range g.units {
					if unit.isSelected {
						unit.state = StateAttacking
						unit.targetEnemy = enemy
						unit.targetNode = nil
					}
				}
				clickedEnemy = true
				break
			}
		}
		if clickedEnemy {
			return
		}

		// 2. Check if clicked a resource node (Harvest)
		for _, node := range g.resourceNodes {
			if mouseX >= node.x && mouseX <= node.x+node.width &&
				mouseY >= node.y && mouseY <= node.y+node.height {

				for _, unit := range g.units {
					if unit.isSelected {
						unit.state = StateMovingToHarvest
						unit.targetNode = node
						unit.targetX = node.x
						unit.targetY = node.y
					}
				}
				clickedNode = true
				break
			}
		}

		// 3. Move Command
		if !clickedNode {
			for _, unit := range g.units {
				if unit.isSelected {
					unit.state = StateMoving
					unit.targetNode = nil
					unit.targetX = targetX
					unit.targetY = targetY
				}
			}
		}
	}

	// --- Left Click (Select Unit) ---
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		unitClicked := false
		var clickedUnit *Unit
		for _, unit := range g.units {
			if mouseX >= unit.x && mouseX <= unit.x+unitSize &&
				mouseY >= unit.y && mouseY <= unit.y+unitSize {
				unitClicked = true
				clickedUnit = unit
				break
			}
		}

		shiftPressed := ebiten.IsKeyPressed(ebiten.KeyShift)

		if unitClicked && clickedUnit != nil {
			if shiftPressed {
				// Toggle selection
				clickedUnit.isSelected = !clickedUnit.isSelected
			} else {
				// Deselect others and select this one
				for _, u := range g.units {
					u.isSelected = (u == clickedUnit)
				}
			}
		} else {
			// Clicked empty ground -> start drag selection
			g.isDragging = true
			g.dragStartX, g.dragStartY = int(mouseX), int(mouseY)
		}
	}

	// --- Drag Selection ---
	if g.isDragging {
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			g.isDragging = false
			selectionRect := image.Rect(g.dragStartX, g.dragStartY, int(mouseX), int(mouseY)).Canon()

			// Only perform box selection if the dragged area is large enough (to avoid overriding single click on release)
			if selectionRect.Dx() > 4 || selectionRect.Dy() > 4 {
				shiftPressed := ebiten.IsKeyPressed(ebiten.KeyShift)

				for _, unit := range g.units {
					unitRect := image.Rect(int(unit.x), int(unit.y), int(unit.x)+int(unitSize), int(unit.y)+int(unitSize))
					if selectionRect.Overlaps(unitRect) {
						unit.isSelected = true
					} else if !shiftPressed {
						unit.isSelected = false
					}
				}
			} else {
				// Simple click on empty ground -> deselect all unless Shift is held
				shiftPressed := ebiten.IsKeyPressed(ebiten.KeyShift)
				if !shiftPressed {
					for _, unit := range g.units {
						unit.isSelected = false
					}
				}
			}
		}
	}
}

// handleProductionInput handles hotkeys for training units
func (g *Game) handleProductionInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		for _, b := range g.buildings {
			// Find a completed building that is not currently busy
			if b.buildProgress >= 1.0 && !b.isBuilding {
				if g.playerResources >= unitCost {
					g.playerResources -= unitCost
					b.isBuilding = true
					b.productionProgress = 0.0
					break // Only build in one barracks per key press
				}
			}
		}
	}
}
