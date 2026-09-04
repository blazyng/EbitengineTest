package game

import (
	"image"
	"testing"
)

func TestPathfindingDirectLine(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to initialize game: %v", err)
	}

	// Empty map with no obstacles between (100, 100) and (300, 100)
	g.buildings = nil
	g.resourceNodes = nil
	g.InvalidatePathGrid()

	path := g.FindPath(100, 100, 300, 100, nil, nil)
	if len(path) == 0 {
		t.Fatalf("Expected a path, got empty")
	}

	// Line-of-sight smoothing should reduce clear straight path to a single endpoint waypoint
	last := path[len(path)-1]
	if distance(last.X, last.Y, 300, 100) > 4.0 {
		t.Errorf("Path destination (%f, %f) expected near (300, 100)", last.X, last.Y)
	}
}

func TestPathfindingAroundBuilding(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to initialize game: %v", err)
	}

	// Place an obstacle building directly between start (50, 100) and goal (250, 100)
	// Building at (120, 80), width 64, height 64 (spans x: 120-184, y: 80-144)
	obstacle := NewBuilding(120, 80)
	obstacle.buildProgress = 1.0
	g.buildings = []*Building{obstacle}
	g.resourceNodes = nil
	g.InvalidatePathGrid()

	path := g.FindPath(50, 100, 250, 100, nil, nil)
	if len(path) == 0 {
		t.Fatalf("Expected path around building, got empty")
	}

	// Ensure no waypoint places the unit inside the building's collision box
	for i, pt := range path {
		unitBox := image.Rect(int(pt.X), int(pt.Y), int(pt.X+unitSize), int(pt.Y+unitSize))
		if unitBox.Overlaps(obstacle.BoundingBox()) {
			t.Errorf("Waypoint %d (%f, %f) collides with obstacle bounding box %v", i, pt.X, pt.Y, obstacle.BoundingBox())
		}
	}

	// Check that the last waypoint reaches the goal
	last := path[len(path)-1]
	if distance(last.X, last.Y, 250, 100) > 16.0 {
		t.Errorf("Path destination (%f, %f) expected near goal (250, 100)", last.X, last.Y)
	}
}

func TestPathfindingTargetPassThrough(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to initialize game: %v", err)
	}

	targetBuilding := NewBuilding(200, 200)
	targetBuilding.buildProgress = 1.0
	g.buildings = []*Building{targetBuilding}
	g.resourceNodes = nil
	g.InvalidatePathGrid()

	// Target is inside the building itself (e.g. drop-off or builder target)
	targetX := targetBuilding.x + targetBuilding.width/2
	targetY := targetBuilding.y + targetBuilding.height/2

	// When targetBuilding is ignored, path should be found directly to it
	path := g.FindPath(50, 200, targetX, targetY, targetBuilding, nil)
	if len(path) == 0 {
		t.Fatalf("Expected path to target building with pass-through, got empty")
	}

	last := path[len(path)-1]
	if distance(last.X, last.Y, targetX, targetY) > 8.0 {
		t.Errorf("Expected path to reach target building center, got last (%f, %f)", last.X, last.Y)
	}
}

func TestUnitNavigatesAroundObstacleInGame(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to initialize game: %v", err)
	}

	// Place obstacle building in middle
	obstacle := NewBuilding(100, 50)
	obstacle.buildProgress = 1.0
	g.buildings = []*Building{obstacle}
	g.resourceNodes = nil
	g.InvalidatePathGrid()

	// Place unit at (40, 66) and command to move to (200, 66)
	u := NewUnit(40, 66, 1)
	u.speed = 4.0
	g.units = []*Unit{u}

	u.isSelected = true
	u.state = StateMoving
	u.targetX = 200
	u.targetY = 66
	u.path = g.FindPath(u.x, u.y, u.targetX, u.targetY, nil, nil)

	// Simulate game ticks
	reached := false
	for tick := 0; tick < 200; tick++ {
		err := g.Update()
		if err != nil {
			t.Fatalf("Update error at tick %d: %v", tick, err)
		}

		// Check collision at all times
		unitBox := image.Rect(int(u.x), int(u.y), int(u.x+unitSize), int(u.y+unitSize))
		if unitBox.Overlaps(obstacle.BoundingBox()) {
			t.Fatalf("Unit collided with building at tick %d! Pos: (%f, %f)", tick, u.x, u.y)
		}

		if distance(u.x, u.y, 200, 66) < 6.0 {
			reached = true
			break
		}
	}

	if !reached {
		t.Fatalf("Unit did not reach destination in 200 ticks! Pos: (%f, %f), State: %v", u.x, u.y, u.state)
	}
}
