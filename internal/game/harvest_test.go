package game

import (
	"fmt"
	"testing"
)

func TestHarvestAndReturn(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Make sure we have 2 units and 1 resource node
	g.units = []*Unit{
		NewUnit(50, 50, 1),
		NewUnit(80, 80, 1),
	}
	g.resourceNodes = []*ResourceNode{
		NewResourceNode(200, 200, 100),
	}
	// Keep barracks at default position to test real game behavior
	g.buildings = []*Building{
		NewBuilding(10, 100),
	}
	g.buildings[0].buildProgress = 1.0

	// Order both units to harvest
	for _, u := range g.units {
		u.isSelected = true
		u.state = StateMovingToHarvest
		u.targetNode = g.resourceNodes[0]
		u.targetX = g.resourceNodes[0].x
		u.targetY = g.resourceNodes[0].y
	}

	fmt.Printf("Initial units pos: %f, %f and %f, %f\n", g.units[0].x, g.units[0].y, g.units[1].x, g.units[1].y)

	// Run update loop
	for i := 0; i < 1000; i++ {
		err := g.Update()
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Print state transitions for both units
		if i%20 == 0 || g.units[0].state == StateHarvesting || g.units[1].state == StateHarvesting {
			fmt.Printf("Tick %d: Unit0 State: %d, Pos: (%f, %f), Cargo: %d; Unit1 State: %d, Pos: (%f, %f), Cargo: %d; Resources: %d\n",
				i, g.units[0].state, g.units[0].x, g.units[0].y, g.units[0].cargo,
				g.units[1].state, g.units[1].x, g.units[1].y, g.units[1].cargo,
				g.playerResources)
		}

		if g.playerResources > 1000 {
			fmt.Println("Success! Resources increased.")
			return
		}
		if (g.units[0].state == StateIdle && g.units[0].cargo == 0) || (g.units[1].state == StateIdle && g.units[1].cargo == 0) {
			t.Fatalf("A unit went Idle without delivering resources at tick %d!", i)
		}
	}
	t.Fatal("Unit failed to deliver resources within 1000 ticks!")
}
