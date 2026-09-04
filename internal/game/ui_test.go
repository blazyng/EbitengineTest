package game

import (
	"testing"
)

func TestUIInterception(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// 1. Click in viewport (e.g. x: 100, y: 100) -> should NOT be consumed by UI
	if g.handleUIInput(100, 100) {
		t.Errorf("Expected mouse in game view to not be consumed by UI")
	}

	// 2. Click in top resource bar (y < 14) -> should be consumed by UI
	if !g.handleUIInput(50, 5) {
		t.Errorf("Expected mouse in top bar to be consumed by UI")
	}

	// 3. Click in bottom HUD panel (y >= HudY) -> should be consumed by UI
	if !g.handleUIInput(160, 200) {
		t.Errorf("Expected mouse in bottom HUD to be consumed by UI")
	}
}

func TestMinimapCameraJump(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Jump to center of minimap
	centerX := float32(MinimapX + MinimapSize/2)
	centerY := float32(MinimapY + MinimapSize/2)
	g.jumpCameraToMinimap(centerX, centerY)

	expectedWorldCenterX := MapWidth / 2.0
	expectedWorldCenterY := MapHeight / 2.0
	expectedCamX := expectedWorldCenterX - float64(ViewWidth)/2.0
	expectedCamY := expectedWorldCenterY - float64(ViewHeight)/2.0

	if g.cameraX < expectedCamX-10 || g.cameraX > expectedCamX+10 {
		t.Errorf("Expected cameraX close to %f, got %f", expectedCamX, g.cameraX)
	}
	if g.cameraY < expectedCamY-10 || g.cameraY > expectedCamY+10 {
		t.Errorf("Expected cameraY close to %f, got %f", expectedCamY, g.cameraY)
	}

	// Jump to top-left (0,0) -> camera should clamp to (0,0)
	g.jumpCameraToMinimap(MinimapX, MinimapY)
	if g.cameraX != 0 || g.cameraY != 0 {
		t.Errorf("Expected camera clamped to (0,0), got (%f, %f)", g.cameraX, g.cameraY)
	}

	// Jump to bottom-right -> camera should clamp to (MapWidth - ViewWidth, MapHeight - ViewHeight)
	g.jumpCameraToMinimap(MinimapX+MinimapSize, MinimapY+MinimapSize)
	maxCamX := MapWidth - float64(ViewWidth)
	maxCamY := MapHeight - float64(ViewHeight)
	if g.cameraX != maxCamX || g.cameraY != maxCamY {
		t.Errorf("Expected camera clamped to (%f, %f), got (%f, %f)", maxCamX, maxCamY, g.cameraX, g.cameraY)
	}
}

func TestUICommandButtons(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// 1. Default buttons (No selection)
	btns := g.getCommandButtons()
	if len(btns) < 2 {
		t.Fatalf("Expected at least 2 default buttons, got %d", len(btns))
	}
	if btns[0].Label != "Build" || btns[1].Label != "Train" {
		t.Errorf("Unexpected default buttons: %s, %s", btns[0].Label, btns[1].Label)
	}

	// 2. Unit Selected -> "Build" and "Stop" buttons
	g.units[0].isSelected = true
	g.units[0].state = StateMoving
	btns = g.getCommandButtons()
	if len(btns) < 2 || btns[1].Label != "Stop" {
		t.Fatalf("Expected Stop button when unit is selected")
	}

	// Execute Stop action
	btns[1].Action()
	if g.units[0].state != StateIdle {
		t.Errorf("Expected unit to become StateIdle after Stop, got %v", g.units[0].state)
	}

	// 3. Building Selected -> "Worker", "Troop", "Rocket", and "Close" buttons
	g.units[0].isSelected = false
	hq := g.buildings[0]
	hq.isSelected = true
	g.selectedBuilding = hq

	btns = g.getCommandButtons()
	if len(btns) < 4 || btns[0].Label != "Worker" || btns[1].Label != "Troop" || btns[2].Label != "Rocket" || btns[3].Label != "Close" {
		t.Fatalf("Expected Worker, Troop, Rocket, and Close buttons for building, got %d buttons", len(btns))
	}

	// Train worker via action
	initRes := g.playerResources
	wCost := GetUnitConfig(hq.faction, UnitTypeWorker).Cost
	btns[0].Action()
	if !hq.isBuilding {
		t.Errorf("Expected building to start training unit")
	}
	if g.playerResources != initRes-wCost {
		t.Errorf("Expected resources to decrease by %d, got %d", wCost, g.playerResources)
	}

	// Close action
	btns[3].Action()
	if g.selectedBuilding != nil || hq.isSelected {
		t.Errorf("Expected building to be deselected")
	}

	// 4. Build Mode Active -> "Cancel" button
	g.isPlacingBuilding = true
	btns = g.getCommandButtons()
	if len(btns) < 1 || btns[0].Label != "Cancel" {
		t.Fatalf("Expected Cancel button during placement")
	}
	btns[0].Action()
	if g.isPlacingBuilding {
		t.Errorf("Expected placement mode to be cancelled")
	}
}

func TestFactionUnits(t *testing.T) {
	usaWorker := NewFactionUnit(0, 0, 1, FactionUSA, UnitTypeWorker)
	if !usaWorker.canHarvest || !usaWorker.canBuild {
		t.Errorf("Expected USA worker to be able to harvest and build")
	}
	if usaWorker.name != "M.U.L.E. Drone" {
		t.Errorf("Expected M.U.L.E. Drone, got %s", usaWorker.name)
	}

	usaMarine := NewFactionUnit(0, 0, 1, FactionUSA, UnitTypeInfantry)
	if usaMarine.canHarvest || usaMarine.canBuild {
		t.Errorf("Expected USA marine to NOT harvest or build")
	}
	if usaMarine.attackRange <= usaWorker.attackRange {
		t.Errorf("Expected marine to have longer attack range than worker")
	}

	chinaDozer := NewFactionUnit(0, 0, 2, FactionChina, UnitTypeWorker)
	if chinaDozer.cargoCapacity != 12 {
		t.Errorf("Expected China dozer to have cargo capacity 12, got %d", chinaDozer.cargoCapacity)
	}
}
