package game

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	MapWidth  = 2000.0
	MapHeight = 2000.0
)

func (g *Game) clampCamera() {
	if g.cameraZoom <= 0 {
		g.cameraZoom = 1.0
	}
	maxCamX := MapWidth - float64(ViewWidth)/g.cameraZoom
	if maxCamX < 0 {
		g.cameraX = 0
	} else if g.cameraX > maxCamX {
		g.cameraX = maxCamX
	}
	if g.cameraX < 0 {
		g.cameraX = 0
	}

	maxCamY := MapHeight - float64(ViewHeight)/g.cameraZoom
	if maxCamY < 0 {
		g.cameraY = 0
	} else if g.cameraY > maxCamY {
		g.cameraY = maxCamY
	}
	if g.cameraY < 0 {
		g.cameraY = 0
	}
}

func (g *Game) updateZoomInput() {
	_, wheelY := ebiten.Wheel()
	screenMouseX, screenMouseY := ebiten.CursorPosition()

	zoomDelta := 0.0
	// Only zoom with wheel if cursor is in game viewport (not in HUD)
	if screenMouseY < HudY && screenMouseY >= 0 && screenMouseX >= 0 && screenMouseX <= ViewWidth {
		if wheelY > 0 {
			zoomDelta = 0.15
		} else if wheelY < 0 {
			zoomDelta = -0.15
		}
	}

	// Keyboard zoom shortcuts
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		zoomDelta += 0.2
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		zoomDelta -= 0.2
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) || inpututil.IsKeyJustPressed(ebiten.Key0) {
		g.setZoom(1.0, float64(ViewWidth)/2.0, float64(ViewHeight)/2.0)
		return
	}

	if zoomDelta != 0 {
		anchorX := float64(screenMouseX)
		anchorY := float64(screenMouseY)
		if anchorX < 0 || anchorX > ViewWidth || anchorY < 0 || anchorY > ViewHeight {
			anchorX = float64(ViewWidth) / 2.0
			anchorY = float64(ViewHeight) / 2.0
		}
		g.setZoom(g.cameraZoom+zoomDelta, anchorX, anchorY)
	}
}

func (g *Game) setZoom(targetZoom, anchorX, anchorY float64) {
	if targetZoom < 0.5 {
		targetZoom = 0.5
	}
	if targetZoom > 2.5 {
		targetZoom = 2.5
	}
	if targetZoom == g.cameraZoom {
		return
	}

	// Keep the world point under the cursor unchanged
	worldAnchorX := g.cameraX + anchorX/g.cameraZoom
	worldAnchorY := g.cameraY + anchorY/g.cameraZoom

	g.cameraZoom = targetZoom

	g.cameraX = worldAnchorX - anchorX/g.cameraZoom
	g.cameraY = worldAnchorY - anchorY/g.cameraZoom

	g.clampCamera()
}

func (g *Game) updateCameraInput() {
	if g.cameraZoom <= 0 {
		g.cameraZoom = 1.0
	}

	// Update zoom
	g.updateZoomInput()

	speed := 5.0 / g.cameraZoom

	// 1. Keyboard Controls
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

	// 2. Mouse Edge-Scrolling Controls
	mouseX, mouseY := ebiten.CursorPosition()
	edgeThreshold := 8

	if mouseX >= 0 && mouseX <= 320 && mouseY >= 0 && mouseY <= 240 {
		if mouseX < edgeThreshold {
			g.cameraX -= speed
		}
		if mouseX > 320-edgeThreshold {
			g.cameraX += speed
		}
		if mouseY < edgeThreshold {
			g.cameraY -= speed
		}
		if mouseY > 240-edgeThreshold {
			g.cameraY += speed
		}
	}

	g.clampCamera()
}

func (g *Game) handleInput(screenMouseX, screenMouseY int) {
	// Fullscreen toggle with F11
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	// 1. Give UI (Minimap, HUD, Buttons) first priority on input
	if g.handleUIInput(screenMouseX, screenMouseY) {
		return
	}

	// 2. Global hotkeys
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.StopSelectedUnits()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.isPlacingBuilding {
			g.isPlacingBuilding = false
		} else if g.isBuildMenuOpen {
			g.isBuildMenuOpen = false
		} else {
			for _, u := range g.units {
				u.isSelected = false
			}
			if g.selectedBuilding != nil {
				g.selectedBuilding.isSelected = false
				g.selectedBuilding = nil
			}
		}
	}

	// Convert screen coordinates to world coordinates (using zoom)
	mouseX := g.cameraX + float64(screenMouseX)/g.cameraZoom
	mouseY := g.cameraY + float64(screenMouseY)/g.cameraZoom

	// --- Build Mode Logic ---
	if g.isPlacingBuilding {
		g.ghostX = mouseX
		g.ghostY = mouseY
		cfg := GetBuildingConfig(g.placingBuildingType)

		// 1. Check if placement is valid (collision check)
		canBuild := true

		// Cannot place building over HUD
		if float64(screenMouseY)+cfg.Height*g.cameraZoom > float64(HudY) {
			canBuild = false
		}

		ghostRect := image.Rect(int(g.ghostX), int(g.ghostY), int(g.ghostX+cfg.Width), int(g.ghostY+cfg.Height))

		if canBuild {
			for _, b := range g.buildings {
				if ghostRect.Overlaps(b.BoundingBox()) {
					canBuild = false
					break
				}
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
			if canBuild && g.playerResources >= cfg.Cost {
				g.playerResources -= cfg.Cost

				newB := NewSpecificBuilding(g.ghostX, g.ghostY, 1, g.playerFaction, g.placingBuildingType)
				newB.buildProgress = 0.0
				g.buildings = append(g.buildings, newB)
				g.InvalidatePathGrid()

				// Order selected builders to construct (or find idle builder)
				assigned := false
				for _, unit := range g.units {
					if unit.isSelected && unit.canBuild {
						unit.state = StateMovingToBuild
						unit.targetBuilding = newB
						unit.targetX = newB.x + newB.width/2
						unit.targetY = newB.y + newB.height/2
						unit.path = nil
						unit.stuckFrames = 0
						assigned = true
					}
				}
				if !assigned {
					for _, unit := range g.units {
						if unit.canBuild && unit.state == StateIdle {
							unit.state = StateMovingToBuild
							unit.targetBuilding = newB
							unit.targetX = newB.x + newB.width/2
							unit.targetY = newB.y + newB.height/2
							unit.path = nil
							unit.stuckFrames = 0
							break
						}
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

	// Toggle Build Menu with 'B'
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.isBuildMenuOpen = !g.isBuildMenuOpen
		g.isDragging = false
	}

	// Quick build hotkeys when Build Menu is open: 1: Barracks, 2: Turret, 3: Supply
	if g.isBuildMenuOpen {
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			g.StartPlacement(BuildingBarracks)
		} else if inpututil.IsKeyJustPressed(ebiten.Key2) {
			g.StartPlacement(BuildingTurret)
		} else if inpututil.IsKeyJustPressed(ebiten.Key3) {
			g.StartPlacement(BuildingSupply)
		}
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
						unit.path = nil
						unit.stuckFrames = 0
					}
				}
				clickedEnemy = true
				break
			}
		}
		if clickedEnemy {
			return
		}

		// 2. Check if clicked an unfinished building (Build)
		for _, b := range g.buildings {
			if b.buildProgress < 1.0 &&
				mouseX >= b.x && mouseX <= b.x+b.width &&
				mouseY >= b.y && mouseY <= b.y+b.height {

				for _, unit := range g.units {
					if unit.isSelected {
						if unit.canBuild {
							unit.state = StateMovingToBuild
							unit.targetBuilding = b
							unit.targetX = b.x + b.width/2
							unit.targetY = b.y + b.height/2
							unit.path = nil
							unit.stuckFrames = 0
						} else {
							// Guard the construction site
							unit.state = StateMoving
							unit.targetNode = nil
							unit.targetBuilding = nil
							unit.targetEnemy = nil
							unit.targetX = b.x + b.width/2
							unit.targetY = b.y + b.height/2
							unit.path = nil
							unit.stuckFrames = 0
						}
					}
				}
				return
			}
		}

		// 3. Check if clicked a resource node (Harvest)
		for _, node := range g.resourceNodes {
			if mouseX >= node.x && mouseX <= node.x+node.width &&
				mouseY >= node.y && mouseY <= node.y+node.height {

				for _, unit := range g.units {
					if unit.isSelected {
						if unit.canHarvest {
							unit.state = StateMovingToHarvest
							unit.targetNode = node
							unit.targetX = node.x
							unit.targetY = node.y
							unit.path = nil
							unit.stuckFrames = 0
						} else {
							// Combat units move to defend resource field instead of mining
							unit.state = StateMoving
							unit.targetNode = nil
							unit.targetBuilding = nil
							unit.targetEnemy = nil
							unit.targetX = node.x
							unit.targetY = node.y
							unit.path = nil
							unit.stuckFrames = 0
						}
					}
				}
				clickedNode = true
				break
			}
		}

		// 4. Move Command
		if !clickedNode {
			var selectedUnits []*Unit
			for _, unit := range g.units {
				if unit.isSelected {
					selectedUnits = append(selectedUnits, unit)
				}
			}
			count := len(selectedUnits)
			if count > 0 {
				cols := int(math.Ceil(math.Sqrt(float64(count))))
				spacing := unitSize + 8.0 // 40px formation spacing
				for i, unit := range selectedUnits {
					unit.state = StateMoving
					unit.targetNode = nil
					unit.targetBuilding = nil
					unit.targetEnemy = nil
					unit.stuckFrames = 0

					destX := targetX
					destY := targetY
					if count > 1 {
						col := i % cols
						row := i / cols
						offsetX := (float64(col) - float64(cols-1)/2.0) * spacing
						offsetY := (float64(row) - float64((count-1)/cols)/2.0) * spacing
						destX = math.Max(0, math.Min(MapWidth-unitSize, targetX+offsetX))
						destY = math.Max(0, math.Min(MapHeight-unitSize, targetY+offsetY))
					}
					unit.targetX = destX
					unit.targetY = destY
					unit.path = g.FindPath(unit.x, unit.y, destX, destY, nil, nil)
				}
			}
		}
	}

	// --- Left Click (Select Unit or Building) ---
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
			if g.selectedBuilding != nil {
				g.selectedBuilding.isSelected = false
				g.selectedBuilding = nil
			}
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
			// Check if a building was clicked
			buildingClicked := false
			var clickedBuilding *Building
			for _, b := range g.buildings {
				if mouseX >= b.x && mouseX <= b.x+b.width &&
					mouseY >= b.y && mouseY <= b.y+b.height {
					buildingClicked = true
					clickedBuilding = b
					break
				}
			}

			if buildingClicked && clickedBuilding != nil {
				for _, u := range g.units {
					u.isSelected = false
				}
				if g.selectedBuilding != nil {
					g.selectedBuilding.isSelected = false
				}
				clickedBuilding.isSelected = true
				g.selectedBuilding = clickedBuilding
			} else {
				// Clicked empty ground -> deselect building and start drag selection
				if g.selectedBuilding != nil {
					g.selectedBuilding.isSelected = false
					g.selectedBuilding = nil
				}
				g.isDragging = true
				g.dragStartX, g.dragStartY = int(mouseX), int(mouseY)
			}
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
				if g.selectedBuilding != nil {
					g.selectedBuilding.isSelected = false
					g.selectedBuilding = nil
				}
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
					if g.selectedBuilding != nil {
						g.selectedBuilding.isSelected = false
						g.selectedBuilding = nil
					}
				}
			}
		}
	}
}

// handleProductionInput handles hotkeys for training units
func (g *Game) handleProductionInput() {
	// [U]: Train Worker
	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		g.TrainUnit(g.selectedBuilding, UnitTypeWorker)
	}
	// [I]: Train Infantry (Marine / Conscript / Rebel)
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.TrainUnit(g.selectedBuilding, UnitTypeInfantry)
	}
	// [O]: Train Specialist (Javelin / Tank Buster / RPG)
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		g.TrainUnit(g.selectedBuilding, UnitTypeSpecialist)
	}
}

// StartPlacement initiates ghost building placement for a specific building type
func (g *Game) StartPlacement(bType BuildingType) {
	cfg := GetBuildingConfig(bType)
	if g.playerResources < cfg.Cost {
		return
	}
	g.placingBuildingType = bType
	g.isPlacingBuilding = true
	g.isBuildMenuOpen = false
	g.isDragging = false
}
