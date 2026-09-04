package game

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Game struct holds the global game state
type Game struct {
	units           []*Unit
	enemyUnits      []*Unit
	resourceNodes   []*ResourceNode
	buildings       []*Building
	playerResources int
	basePosition    image.Point

	// Selection state
	isDragging       bool
	dragStartX       int
	dragStartY       int
	selectedBuilding *Building

	// Building placement state
	isBuildMenuOpen      bool
	isPlacingBuilding    bool
	placingBuildingType  BuildingType
	ghostX, ghostY       float64
	canBuildHere         bool

	// Camera state
	cameraX    float64
	cameraY    float64
	cameraZoom float64

	// Minimap dragging state
	isDraggingMinimap bool

	// Factions
	playerFaction FactionType
	enemyFaction  FactionType

	// Pathfinding
	basePathGrid *PathGrid

	// Combat & Projectiles
	projectiles []*Projectile
	particles   []*Particle
}

const (
	unitCost      = 50
	unitBuildTime = 5.0
)

// NewGame initializes the game world
func NewGame() (*Game, error) {
	g := &Game{
		basePosition:    image.Point{X: 10, Y: 10},
		playerResources: 1000,
		playerFaction:   FactionUSA,
		enemyFaction:    FactionChina,
		cameraZoom:      1.0,
	}

	// Initialize Player Units (Team 1, USA: 2 M.U.L.E. Drones + 1 Marine)
	g.units = []*Unit{
		NewFactionUnit(150, 50, 1, FactionUSA, UnitTypeWorker),
		NewFactionUnit(150, 90, 1, FactionUSA, UnitTypeWorker),
		NewFactionUnit(180, 70, 1, FactionUSA, UnitTypeInfantry),
	}

	// Initialize Enemy Units (Team 2, China: 2 Dozers + 1 Conscript + 1 Tank Buster)
	g.enemyUnits = []*Unit{
		NewFactionUnit(250, 100, 2, FactionChina, UnitTypeWorker),
		NewFactionUnit(250, 140, 2, FactionChina, UnitTypeWorker),
		NewFactionUnit(280, 120, 2, FactionChina, UnitTypeInfantry),
		NewFactionUnit(300, 140, 2, FactionChina, UnitTypeSpecialist),
	}

	// Initialize Resources
	g.resourceNodes = []*ResourceNode{
		NewResourceNode(200, 200, 1000),
	}

	// Initialize Buildings (Start with one pre-built Headquarters at base position)
	startHQ := NewFactionBuilding(10, 10, 1, FactionUSA)
	startHQ.buildProgress = 1.0
	g.buildings = []*Building{startHQ}

	return g, nil
}

// distance calculates the Euclidean distance between two points
func distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

// Update handles game logic updates
func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()

	// 1. Update Camera
	g.updateCameraInput()

	// 2. Handle Input
	g.handleInput(mouseX, mouseY)
	g.handleProductionInput()

	// 3. Update World State
	g.updateUnits(g.units, g.enemyUnits) // Player units
	g.updateUnits(g.enemyUnits, g.units) // Enemy units
	g.updateBuildings()

	// 4. Update Combat & FX
	dt := 1.0 / float64(ebiten.TPS())
	if dt <= 0 {
		dt = 1.0 / 60.0
	}
	g.updateProjectiles(dt)
	g.updateParticles(dt)

	// 5. Cleanup
	g.cleanupDeadUnits()
	g.cleanupDepletedResources()

	return nil
}

// Draw renders the game world
func (g *Game) Draw(screen *ebiten.Image) {
	// 1. Draw Background (Tiled with Zoom & Viewport Culling)
	w, h := ImgGround.Bounds().Dx(), ImgGround.Bounds().Dy()
	tileW := float64(w)
	tileH := float64(h)

	minTileX := int(math.Floor(g.cameraX/tileW)) * w
	minTileY := int(math.Floor(g.cameraY/tileH)) * h

	effectiveViewW := float64(ViewWidth) / g.cameraZoom
	effectiveViewH := float64(ViewHeight) / g.cameraZoom
	maxTileX := int(math.Ceil((g.cameraX+effectiveViewW)/tileW)) * w
	maxTileY := int(math.Ceil((g.cameraY+effectiveViewH)/tileH)) * h

	if maxTileX > MapWidth {
		maxTileX = MapWidth
	}
	if maxTileY > MapHeight {
		maxTileY = MapHeight
	}

	for y := minTileY; y < maxTileY; y += h {
		for x := minTileX; x < maxTileX; x += w {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(g.cameraZoom, g.cameraZoom)
			screenX := (float64(x) - g.cameraX) * g.cameraZoom
			screenY := (float64(y) - g.cameraY) * g.cameraZoom
			op.GeoM.Translate(screenX, screenY)
			screen.DrawImage(ImgGround, op)
		}
	}

	// 2. Draw Resource Nodes
	for _, node := range g.resourceNodes {
		node.Draw(screen, g.cameraX, g.cameraY, g.cameraZoom)
	}

	// 3. Draw Buildings
	for _, b := range g.buildings {
		b.Draw(screen, g.cameraX, g.cameraY, g.cameraZoom)
	}

	// 4. Draw Units (Player + Enemy)
	allUnits := append(g.units, g.enemyUnits...)
	for _, unit := range allUnits {
		unit.Draw(screen, g.cameraX, g.cameraY, g.cameraZoom)
	}

	// 5. Draw Projectiles and Visual Combat FX
	g.DrawCombat(screen, g.cameraX, g.cameraY, g.cameraZoom)

	// 6. Draw Ghost Building (Placement Mode)
	if g.isPlacingBuilding {
		g.DrawGhostBuilding(screen, g.placingBuildingType, g.ghostX, g.ghostY, g.canBuildHere, g.cameraX, g.cameraY, g.cameraZoom)
	}

	// 6. Draw Selection Box
	if g.isDragging {
		mouseX, mouseY := ebiten.CursorPosition()
		// Convert drag start to screen coordinates
		startXScreen := float32((float64(g.dragStartX) - g.cameraX) * g.cameraZoom)
		startYScreen := float32((float64(g.dragStartY) - g.cameraY) * g.cameraZoom)

		w := float32(mouseX) - startXScreen
		h := float32(mouseY) - startYScreen

		vector.StrokeRect(screen, startXScreen, startYScreen, w, h, 1, color.RGBA{0, 255, 0, 255}, false)
	}

	// 7. Render RTS UI / HUD
	g.DrawHUD(screen)
}

// getSelectedUnits returns a slice of currently selected player units
func (g *Game) getSelectedUnits() []*Unit {
	var selected []*Unit
	for _, u := range g.units {
		if u.isSelected {
			selected = append(selected, u)
		}
	}
	return selected
}

// TrainUnit orders a building to produce a unit of the specified type
func (g *Game) TrainUnit(b *Building, uType UnitType) bool {
	if b == nil {
		for _, building := range g.buildings {
			if building.buildProgress >= 1.0 && !building.isBuilding {
				b = building
				break
			}
		}
	}
	if b != nil && b.buildProgress >= 1.0 && !b.isBuilding {
		cfg := GetUnitConfig(b.faction, uType)
		if g.playerResources >= cfg.Cost {
			g.playerResources -= cfg.Cost
			b.isBuilding = true
			b.producingType = uType
			b.productionTime = cfg.BuildTime
			b.productionProgress = 0.0
			return true
		}
	}
	return false
}

// StopSelectedUnits resets selected units to idle state
func (g *Game) StopSelectedUnits() {
	for _, u := range g.units {
		if u.isSelected {
			u.state = StateIdle
			u.targetEnemy = nil
			u.targetNode = nil
			u.targetBuilding = nil
			u.targetX = u.x
			u.targetY = u.y
			u.path = nil
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}
