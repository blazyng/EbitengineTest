package game

import (
	"testing"
)

func TestBuildingConfigs(t *testing.T) {
	hq := GetBuildingConfig(BuildingHQ)
	if hq.Cost <= 0 || hq.MaxHealth != 1000 {
		t.Errorf("Unexpected HQ config: %+v", hq)
	}

	barracks := GetBuildingConfig(BuildingBarracks)
	if barracks.Cost != 100 || barracks.Width != 64 {
		t.Errorf("Unexpected Barracks config: %+v", barracks)
	}

	turret := GetBuildingConfig(BuildingTurret)
	if turret.Cost != 150 || turret.Width != 48 {
		t.Errorf("Unexpected Turret config: %+v", turret)
	}

	supply := GetBuildingConfig(BuildingSupply)
	if supply.Cost != 75 || supply.Width != 48 {
		t.Errorf("Unexpected Supply config: %+v", supply)
	}
}

func TestBuildMenuAndPlacement(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to initialize game: %v", err)
	}

	g.playerResources = 300
	g.isBuildMenuOpen = true

	btns := g.getCommandButtons()
	if len(btns) < 4 {
		t.Fatalf("Expected 4 build menu buttons, got %d", len(btns))
	}

	if btns[0].Label != "Barracks" || btns[1].Label != "Turret" || btns[2].Label != "Supply" {
		t.Errorf("Unexpected build menu buttons: %+v", btns)
	}

	// Click Turret button (index 1)
	btns[1].Action()

	if !g.isPlacingBuilding {
		t.Errorf("Expected isPlacingBuilding to be true")
	}
	if g.placingBuildingType != BuildingTurret {
		t.Errorf("Expected placingBuildingType to be BuildingTurret, got %v", g.placingBuildingType)
	}
	if g.isBuildMenuOpen {
		t.Errorf("Expected isBuildMenuOpen to be false")
	}
}

func TestTurretAutomatedAttack(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to initialize game: %v", err)
	}

	g.buildings = nil
	g.units = nil
	g.projectiles = nil
	g.particles = nil

	// Place player turret at (100, 100)
	turret := NewSpecificBuilding(100, 100, 1, FactionUSA, BuildingTurret)
	turret.buildProgress = 1.0 // Fully constructed
	g.buildings = []*Building{turret}

	// Place enemy unit within range at (160, 100) - distance ~60px, well within 150px range
	enemy := NewFactionUnit(160, 100, 2, FactionChina, UnitTypeInfantry)
	initHealth := enemy.health
	g.enemyUnits = []*Unit{enemy}

	// Run update loop for 1.5 seconds (90 ticks)
	for tick := 0; tick < 90; tick++ {
		err := g.Update()
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	}

	if enemy.health >= initHealth {
		t.Errorf("Expected turret to damage enemy! Initial HP: %d, Current HP: %d", initHealth, enemy.health)
	}
}

func TestSupplyDepotPassiveIncome(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to initialize game: %v", err)
	}

	g.buildings = nil
	g.units = nil
	g.enemyUnits = nil
	g.playerResources = 100

	supply := NewSpecificBuilding(100, 100, 1, FactionUSA, BuildingSupply)
	supply.buildProgress = 1.0
	g.buildings = []*Building{supply}

	// Run update for ~5 seconds (300 ticks) - should trigger at least one +10 gold income
	for i := 0; i < 300; i++ {
		err := g.Update()
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	}

	if g.playerResources <= 100 {
		t.Errorf("Expected passive income from Supply Depot, but resources are %d", g.playerResources)
	}
}
