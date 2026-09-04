package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	ViewWidth   = 320
	ViewHeight  = 190
	HudY        = 190
	HudHeight   = 50
	MinimapX    = 2
	MinimapY    = 192
	MinimapSize = 46
)

// UIButton defines a clickable UI button on the HUD
type UIButton struct {
	X, Y, W, H float32
	Label      string
	Subtext    string
	Action     func()
	Disabled   bool
	Danger     bool
}

// handleUIInput processes mouse interactions with the UI.
// Returns true if the mouse click or drag was consumed by the UI.
func (g *Game) handleUIInput(screenMouseX, screenMouseY int) bool {
	mouseX := float32(screenMouseX)
	mouseY := float32(screenMouseY)

	// 1. Check Minimap Dragging / Clicking
	if g.isDraggingMinimap {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.jumpCameraToMinimap(mouseX, mouseY)
			return true
		} else {
			g.isDraggingMinimap = false
		}
	}

	// 2. If mouse is in Top Bar (y < 14), block world clicks
	if mouseY < 14 {
		return true
	}

	// 3. If mouse is not in HUD (y < HudY), let world handle it
	if mouseY < HudY {
		return false
	}

	// 4. Mouse is inside bottom HUD:
	// Handle Minimap Left-Click
	if mouseX >= MinimapX && mouseX <= MinimapX+MinimapSize &&
		mouseY >= MinimapY && mouseY <= MinimapY+MinimapSize {

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.isDraggingMinimap = true
			g.jumpCameraToMinimap(mouseX, mouseY)
			return true
		}

		// Minimap Right-Click: Order selected units to move to location
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			relX := float64(mouseX - MinimapX)
			relY := float64(mouseY - MinimapY)
			targetX := (relX / float64(MinimapSize)) * MapWidth
			targetY := (relY / float64(MinimapSize)) * MapHeight

			var selectedUnits []*Unit
			for _, unit := range g.units {
				if unit.isSelected {
					selectedUnits = append(selectedUnits, unit)
				}
			}
			count := len(selectedUnits)
			if count > 0 {
				cols := int(math.Ceil(math.Sqrt(float64(count))))
				spacing := unitSize + 8.0
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
			return true
		}
		return true
	}

	// Handle Action Buttons Click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		buttons := g.getCommandButtons()
		for _, btn := range buttons {
			if !btn.Disabled && mouseX >= btn.X && mouseX <= btn.X+btn.W &&
				mouseY >= btn.Y && mouseY <= btn.Y+btn.H {
				if btn.Action != nil {
					btn.Action()
				}
				return true
			}
		}
	}

	// Any other click inside the HUD is consumed so it doesn't leak into the game world
	return true
}

func (g *Game) jumpCameraToMinimap(screenX, screenY float32) {
	relX := float64(screenX - MinimapX)
	relY := float64(screenY - MinimapY)

	if relX < 0 {
		relX = 0
	}
	if relX > MinimapSize {
		relX = MinimapSize
	}
	if relY < 0 {
		relY = 0
	}
	if relY > MinimapSize {
		relY = MinimapSize
	}

	worldX := (relX / float64(MinimapSize)) * MapWidth
	worldY := (relY / float64(MinimapSize)) * MapHeight

	zoom := g.cameraZoom
	if zoom <= 0 {
		zoom = 1.0
	}
	g.cameraX = worldX - (float64(ViewWidth)/zoom)/2.0
	g.cameraY = worldY - (float64(ViewHeight)/zoom)/2.0

	g.clampCamera()
}

// getCommandButtons constructs the list of active HUD buttons based on current selection & state
func (g *Game) getCommandButtons() []UIButton {
	var buttons []UIButton

	// State 1: In Building Placement Mode
	if g.isPlacingBuilding {
		buttons = append(buttons, UIButton{
			X:       220,
			Y:       193,
			W:       46,
			H:       22,
			Label:   "Cancel",
			Subtext: "Esc",
			Danger:  true,
			Action: func() {
				g.isPlacingBuilding = false
			},
		})
		buttons = append(buttons, UIButton{
			X:        270,
			Y:        193,
			W:        46,
			H:        22,
			Label:    "Place",
			Subtext:  "L-Click",
			Disabled: true,
		})
		return buttons
	}

	// State 1.5: In Build Menu (Choosing structure to place)
	if g.isBuildMenuOpen {
		barracksCfg := GetBuildingConfig(BuildingBarracks)
		turretCfg := GetBuildingConfig(BuildingTurret)
		supplyCfg := GetBuildingConfig(BuildingSupply)

		// Row 1: Barracks & Turret
		buttons = append(buttons, UIButton{
			X:        220,
			Y:        193,
			W:        46,
			H:        21,
			Label:    "Barracks",
			Subtext:  fmt.Sprintf("%dg [1]", barracksCfg.Cost),
			Disabled: g.playerResources < barracksCfg.Cost,
			Action: func() {
				g.StartPlacement(BuildingBarracks)
			},
		})

		buttons = append(buttons, UIButton{
			X:        270,
			Y:        193,
			W:        46,
			H:        21,
			Label:    "Turret",
			Subtext:  fmt.Sprintf("%dg [2]", turretCfg.Cost),
			Disabled: g.playerResources < turretCfg.Cost,
			Action: func() {
				g.StartPlacement(BuildingTurret)
			},
		})

		// Row 2: Supply & Back
		buttons = append(buttons, UIButton{
			X:        220,
			Y:        216,
			W:        46,
			H:        21,
			Label:    "Supply",
			Subtext:  fmt.Sprintf("%dg [3]", supplyCfg.Cost),
			Disabled: g.playerResources < supplyCfg.Cost,
			Action: func() {
				g.StartPlacement(BuildingSupply)
			},
		})

		buttons = append(buttons, UIButton{
			X:       270,
			Y:       216,
			W:       46,
			H:       21,
			Label:   "Back",
			Subtext: "Esc",
			Danger:  true,
			Action: func() {
				g.isBuildMenuOpen = false
			},
		})
		return buttons
	}

	// State 2: Building is Selected
	if g.selectedBuilding != nil {
		fac := g.selectedBuilding.faction
		wCfg := GetUnitConfig(fac, UnitTypeWorker)
		iCfg := GetUnitConfig(fac, UnitTypeInfantry)
		sCfg := GetUnitConfig(fac, UnitTypeSpecialist)

		canTrainW := g.selectedBuilding.buildProgress >= 1.0 && !g.selectedBuilding.isBuilding && g.playerResources >= wCfg.Cost
		canTrainI := g.selectedBuilding.buildProgress >= 1.0 && !g.selectedBuilding.isBuilding && g.playerResources >= iCfg.Cost
		canTrainS := g.selectedBuilding.buildProgress >= 1.0 && !g.selectedBuilding.isBuilding && g.playerResources >= sCfg.Cost

		// Row 1: Worker & Infantry
		buttons = append(buttons, UIButton{
			X:        220,
			Y:        193,
			W:        46,
			H:        21,
			Label:    "Worker",
			Subtext:  fmt.Sprintf("%dg [U]", wCfg.Cost),
			Disabled: !canTrainW,
			Action: func() {
				g.TrainUnit(g.selectedBuilding, UnitTypeWorker)
			},
		})

		buttons = append(buttons, UIButton{
			X:        270,
			Y:        193,
			W:        46,
			H:        21,
			Label:    "Troop",
			Subtext:  fmt.Sprintf("%dg [I]", iCfg.Cost),
			Disabled: !canTrainI,
			Action: func() {
				g.TrainUnit(g.selectedBuilding, UnitTypeInfantry)
			},
		})

		// Row 2: Specialist & Close
		buttons = append(buttons, UIButton{
			X:        220,
			Y:        216,
			W:        46,
			H:        21,
			Label:    "Rocket",
			Subtext:  fmt.Sprintf("%dg [O]", sCfg.Cost),
			Disabled: !canTrainS,
			Action: func() {
				g.TrainUnit(g.selectedBuilding, UnitTypeSpecialist)
			},
		})

		buttons = append(buttons, UIButton{
			X:       270,
			Y:       216,
			W:       46,
			H:       21,
			Label:   "Close",
			Subtext: "Esc",
			Action: func() {
				g.selectedBuilding.isSelected = false
				g.selectedBuilding = nil
			},
		})
		return buttons
	}

	// State 3: Unit(s) Selected
	selectedUnits := g.getSelectedUnits()
	if len(selectedUnits) > 0 {
		canBuild := g.playerResources >= 75
		buttons = append(buttons, UIButton{
			X:        220,
			Y:        193,
			W:        46,
			H:        22,
			Label:    "Build",
			Subtext:  "[B]",
			Disabled: !canBuild,
			Action: func() {
				g.isBuildMenuOpen = true
				g.isDragging = false
			},
		})

		buttons = append(buttons, UIButton{
			X:       270,
			Y:       193,
			W:       46,
			H:       22,
			Label:   "Stop",
			Subtext: "[S]",
			Action: func() {
				g.StopSelectedUnits()
			},
		})
		return buttons
	}

	// State 4: Default (No selection)
	canBuild := g.playerResources >= 75
	buttons = append(buttons, UIButton{
		X:        220,
		Y:        193,
		W:        46,
		H:        22,
		Label:    "Build",
		Subtext:  "[B]",
		Disabled: !canBuild,
		Action: func() {
			g.isBuildMenuOpen = true
			g.isDragging = false
		},
	})

	hasAvailableBuilding := false
	for _, b := range g.buildings {
		if b.buildProgress >= 1.0 && !b.isBuilding {
			hasAvailableBuilding = true
			break
		}
	}
	wCfg := GetUnitConfig(g.playerFaction, UnitTypeWorker)
	canTrain := hasAvailableBuilding && g.playerResources >= wCfg.Cost
	buttons = append(buttons, UIButton{
		X:        270,
		Y:        193,
		W:        46,
		H:        22,
		Label:    "Train",
		Subtext:  fmt.Sprintf("%dg [U]", wCfg.Cost),
		Disabled: !canTrain,
		Action: func() {
			g.TrainUnit(nil, UnitTypeWorker)
		},
	})

	return buttons
}

// DrawHUD renders the complete RTS UI
func (g *Game) DrawHUD(screen *ebiten.Image) {
	g.drawTopBar(screen)
	g.drawBottomHUD(screen)
}

// drawTopBar renders resource and status info at top
func (g *Game) drawTopBar(screen *ebiten.Image) {
	// Semi-transparent background banner
	vector.DrawFilledRect(screen, 0, 0, 320, 14, color.RGBA{18, 22, 30, 220}, false)
	vector.StrokeLine(screen, 0, 14, 320, 14, 1, color.RGBA{50, 60, 75, 255}, false)

	// Resource count with Faction Name
	facInfo := GetFactionInfo(g.playerFaction)
	goldText := fmt.Sprintf("[%s] Gold: %d", facInfo.Name, g.playerResources)
	ebitenutil.DebugPrintAt(screen, goldText, 4, 1)

	// Unit count
	unitText := fmt.Sprintf("Units: %d", len(g.units))
	ebitenutil.DebugPrintAt(screen, unitText, 94, 1)

	// Zoom indicator
	zoom := g.cameraZoom
	if zoom <= 0 {
		zoom = 1.0
	}
	zoomText := fmt.Sprintf("Zoom: %.1fx", zoom)
	ebitenutil.DebugPrintAt(screen, zoomText, 146, 1)

	// Quick hint on right
	hintText := "[Wheel] Zoom [F11] Full"
	ebitenutil.DebugPrintAt(screen, hintText, 208, 1)
}

// drawBottomHUD renders the main bottom dashboard
func (g *Game) drawBottomHUD(screen *ebiten.Image) {
	// Main panel background
	vector.DrawFilledRect(screen, 0, HudY, 320, HudHeight, color.RGBA{22, 26, 34, 250}, false)
	// Top border line
	vector.StrokeLine(screen, 0, HudY, 320, HudY, 1, color.RGBA{65, 75, 95, 255}, false)

	g.drawMinimap(screen)
	g.drawSelectionInfo(screen)
	g.drawCommandButtons(screen)
}

// drawMinimap renders the tactical overview map
func (g *Game) drawMinimap(screen *ebiten.Image) {
	// Background and border
	vector.DrawFilledRect(screen, MinimapX, MinimapY, MinimapSize, MinimapSize, color.RGBA{20, 32, 22, 255}, false)
	vector.StrokeRect(screen, MinimapX, MinimapY, MinimapSize, MinimapSize, 1, color.RGBA{70, 80, 100, 255}, false)

	scale := float64(MinimapSize) / MapWidth

	// 1. Draw Resource Nodes (Yellow dots)
	for _, node := range g.resourceNodes {
		nx := float32(float64(MinimapX) + node.x*scale)
		ny := float32(float64(MinimapY) + node.y*scale)
		vector.DrawFilledRect(screen, nx, ny, 2, 2, color.RGBA{255, 230, 0, 255}, false)
	}

	// 2. Draw Buildings (Blue for player)
	for _, b := range g.buildings {
		bx := float32(float64(MinimapX) + b.x*scale)
		by := float32(float64(MinimapY) + b.y*scale)
		bw := float32(b.width * scale)
		bh := float32(b.height * scale)
		if bw < 2 {
			bw = 2
		}
		if bh < 2 {
			bh = 2
		}
		vector.DrawFilledRect(screen, bx, by, bw, bh, color.RGBA{50, 120, 255, 255}, false)
	}

	// 3. Draw Player Units (Green dots)
	for _, u := range g.units {
		ux := float32(float64(MinimapX) + u.x*scale)
		uy := float32(float64(MinimapY) + u.y*scale)
		vector.DrawFilledRect(screen, ux, uy, 2, 2, color.RGBA{0, 255, 50, 255}, false)
	}

	// 4. Draw Enemy Units (Red dots)
	for _, u := range g.enemyUnits {
		ux := float32(float64(MinimapX) + u.x*scale)
		uy := float32(float64(MinimapY) + u.y*scale)
		vector.DrawFilledRect(screen, ux, uy, 2, 2, color.RGBA{255, 40, 40, 255}, false)
	}

	// 5. Draw Viewport (White rectangle scaled by zoom)
	zoom := g.cameraZoom
	if zoom <= 0 {
		zoom = 1.0
	}
	camBoxX := float32(float64(MinimapX) + g.cameraX*scale)
	camBoxY := float32(float64(MinimapY) + g.cameraY*scale)
	camBoxW := float32((float64(ViewWidth) / zoom) * scale)
	camBoxH := float32((float64(ViewHeight) / zoom) * scale)

	vector.StrokeRect(screen, camBoxX, camBoxY, camBoxW, camBoxH, 1, color.RGBA{255, 255, 255, 220}, false)
}

// drawSelectionInfo renders inspector data for selected unit or building
func (g *Game) drawSelectionInfo(screen *ebiten.Image) {
	panelX := float32(52)
	panelY := float32(192)
	panelW := float32(164)
	panelH := float32(46)

	// Inset box
	vector.DrawFilledRect(screen, panelX, panelY, panelW, panelH, color.RGBA{14, 17, 23, 255}, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1, color.RGBA{45, 55, 70, 255}, false)

	// Case 1: Building Selected
	if g.selectedBuilding != nil {
		b := g.selectedBuilding

		// Building Mini Icon
		iconX, iconY := float32(56), float32(195)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(24.0/64.0, 24.0/64.0)
		op.GeoM.Translate(float64(iconX), float64(iconY))
		screen.DrawImage(ImgBuilding, op)
		vector.StrokeRect(screen, iconX, iconY, 24, 24, 1, color.RGBA{0, 255, 0, 255}, false)

		facInfo := GetFactionInfo(b.faction)
		bCfg := GetBuildingConfig(b.buildingType)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("[%s] %s", facInfo.Name, bCfg.Name), 86, 194)

		if b.buildProgress < 1.0 {
			progPercent := float32(b.buildProgress)
			barW := float32(75)
			vector.DrawFilledRect(screen, 86, 207, barW, 5, color.RGBA{60, 60, 60, 255}, false)
			vector.DrawFilledRect(screen, 86, 207, barW*progPercent, 5, color.RGBA{255, 200, 0, 255}, false)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Build: %d%%", int(progPercent*100)), 86, 217)
		} else if b.isBuilding {
			prodCfg := GetUnitConfig(b.faction, b.producingType)
			progPercent := float32(b.productionProgress)
			barW := float32(75)
			vector.DrawFilledRect(screen, 86, 207, barW, 5, color.RGBA{60, 60, 60, 255}, false)
			vector.DrawFilledRect(screen, 86, 207, barW*progPercent, 5, color.RGBA{0, 255, 100, 255}, false)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Train %s %d%%", prodCfg.Name, int(progPercent*100)), 86, 217)
		} else if b.buildingType == BuildingTurret {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP:%d/%d Dmg:%d", b.health, b.maxHealth, b.attackDamage), 86, 207)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rng:%.0f Auto-Defense", b.attackRange), 86, 218)
		} else if b.buildingType == BuildingSupply {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP:%d/%d Storage", b.health, b.maxHealth), 86, 207)
			ebitenutil.DebugPrintAt(screen, "Income: +10g / 4s", 86, 218)
		} else {
			ebitenutil.DebugPrintAt(screen, "Status: Ready", 86, 207)
			ebitenutil.DebugPrintAt(screen, "Train: [U] [I] [O]", 86, 218)
		}
		return
	}

	// Case 2: Unit(s) Selected
	selectedUnits := g.getSelectedUnits()
	if len(selectedUnits) == 1 {
		u := selectedUnits[0]

		// Unit Mini Icon with Faction & UnitType Tint
		iconX, iconY := float32(56), float32(195)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(24.0/32.0, 24.0/32.0)
		if u.cargo > 0 {
			op.ColorScale.Scale(1.1, 1.1, 0.2, 1)
		} else if u.team == 2 {
			switch u.unitType {
			case UnitTypeWorker:
				op.ColorScale.Scale(1.1, 0.7, 0.4, 1)
			case UnitTypeInfantry:
				op.ColorScale.Scale(1.3, 0.5, 0.5, 1)
			case UnitTypeSpecialist:
				op.ColorScale.Scale(1.4, 0.4, 0.2, 1)
			}
		} else {
			switch u.unitType {
			case UnitTypeWorker:
				op.ColorScale.Scale(1.0, 1.0, 0.8, 1)
			case UnitTypeInfantry:
				op.ColorScale.Scale(0.6, 0.85, 1.3, 1)
			case UnitTypeSpecialist:
				op.ColorScale.Scale(0.8, 0.6, 1.3, 1)
			}
		}
		op.GeoM.Translate(float64(iconX), float64(iconY))
		screen.DrawImage(ImgUnit, op)
		vector.StrokeRect(screen, iconX, iconY, 24, 24, 1, color.RGBA{0, 255, 0, 255}, false)

		// Draw role badge pip on icon
		cfg := GetUnitConfig(u.faction, u.unitType)
		vector.DrawFilledRect(screen, iconX+20, iconY, 4, 4, cfg.BadgeColor, false)

		// Faction & Unit Name
		facInfo := GetFactionInfo(u.faction)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("[%s] %s", facInfo.Name, u.name), 86, 194)

		// Health Bar
		hpRatio := float32(u.health) / float32(u.maxHealth)
		barW := float32(65)
		vector.DrawFilledRect(screen, 86, 206, barW, 4, color.RGBA{180, 40, 40, 255}, false)
		vector.DrawFilledRect(screen, 86, 206, barW*hpRatio, 4, color.RGBA{40, 200, 40, 255}, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d/%d", u.health, u.maxHealth), 154, 203)

		// State string
		stateStr := "Idle"
		switch u.state {
		case StateMoving:
			stateStr = "Moving"
		case StateMovingToHarvest:
			stateStr = "To Gold"
		case StateHarvesting:
			stateStr = "Mining..."
		case StateReturning:
			stateStr = "Drop-off"
		case StateAttacking:
			stateStr = "Attacking"
		case StateMovingToBuild:
			stateStr = "To Site"
		case StateBuilding:
			stateStr = "Building"
		}

		if u.cargo > 0 {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s | Gold:%d", stateStr, u.cargo), 86, 214)
		} else {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s | %s", u.role, stateStr), 86, 214)
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Dmg:%d Rng:%.0f Spd:%.1f", u.attackDamage, u.attackRange, u.speed), 86, 226)
		return
	} else if len(selectedUnits) > 1 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Selected: %d Units", len(selectedUnits)), 60, 198)
		ebitenutil.DebugPrintAt(screen, "Right-Click: Move/Attack", 60, 212)
		ebitenutil.DebugPrintAt(screen, "[S]: Stop all units", 60, 224)
		return
	}

	// Case 3: No Selection
	ebitenutil.DebugPrintAt(screen, "No Selection", 60, 196)
	ebitenutil.DebugPrintAt(screen, "L-Click: Select unit/HQ", 60, 208)
	ebitenutil.DebugPrintAt(screen, "WASD: Pan | Minimap: Click", 60, 220)
}

// drawCommandButtons renders active HUD buttons
func (g *Game) drawCommandButtons(screen *ebiten.Image) {
	mouseX, mouseY := ebiten.CursorPosition()
	mx, my := float32(mouseX), float32(mouseY)

	buttons := g.getCommandButtons()
	for _, btn := range buttons {
		isHovered := !btn.Disabled && mx >= btn.X && mx <= btn.X+btn.W && my >= btn.Y && my <= btn.Y+btn.H
		isPressed := isHovered && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

		// Button background
		bgColor := color.RGBA{35, 42, 54, 255}
		if btn.Disabled {
			bgColor = color.RGBA{24, 28, 35, 255}
		} else if btn.Danger {
			bgColor = color.RGBA{140, 40, 40, 255}
			if isHovered {
				bgColor = color.RGBA{180, 50, 50, 255}
			}
		} else if isPressed {
			bgColor = color.RGBA{55, 80, 120, 255}
		} else if isHovered {
			bgColor = color.RGBA{48, 62, 82, 255}
		}

		vector.DrawFilledRect(screen, btn.X, btn.Y, btn.W, btn.H, bgColor, false)

		// Button border
		borderColor := color.RGBA{70, 82, 105, 255}
		if btn.Disabled {
			borderColor = color.RGBA{40, 45, 55, 255}
		} else if isHovered {
			borderColor = color.RGBA{120, 145, 180, 255}
		}
		vector.StrokeRect(screen, btn.X, btn.Y, btn.W, btn.H, 1, borderColor, false)

		// Text
		textY := int(btn.Y) + 2
		if btn.Subtext != "" {
			textY = int(btn.Y) + 1
		}
		ebitenutil.DebugPrintAt(screen, btn.Label, int(btn.X)+3, textY)
		if btn.Subtext != "" {
			ebitenutil.DebugPrintAt(screen, btn.Subtext, int(btn.X)+3, textY+9)
		}
	}

	// Bottom command info (only shown when 1 row of buttons)
	if g.selectedBuilding == nil {
		infoText := "[B] Build  [U] Train"
		if g.isPlacingBuilding {
			infoText = "L-Click: Place"
		} else if len(g.getSelectedUnits()) > 0 {
			infoText = "[B] Build  [S] Stop"
		}
		ebitenutil.DebugPrintAt(screen, infoText, 220, 226)
	}
}
