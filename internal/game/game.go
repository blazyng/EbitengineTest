package game

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
	isDragging bool
	dragStartX int
	dragStartY int

	// Building placement state
	isPlacingBuilding bool
	ghostX, ghostY    float64
	canBuildHere      bool

	// Camera state
	cameraX float64
	cameraY float64
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
	}

	// Initialize Player Units (Team 1)
	g.units = []*Unit{
		NewUnit(150, 50, 1),
		NewUnit(150, 90, 1),
	}

	// Initialize Enemy Units (Team 2)
	g.enemyUnits = []*Unit{
		NewUnit(250, 100, 2),
		NewUnit(250, 140, 2),
	}

	// Initialize Resources
	g.resourceNodes = []*ResourceNode{
		NewResourceNode(200, 200, 1000),
	}

	// Initialize Buildings (Start with one pre-built Headquarters at base position)
	startHQ := NewBuilding(10, 10)
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

	// 4. Cleanup
	g.cleanupDeadUnits()
	g.cleanupDepletedResources()

	return nil
}

// Draw renders the game world
func (g *Game) Draw(screen *ebiten.Image) {
	// 1. Draw Background (Tiled)
	w, h := ImgGround.Bounds().Dx(), ImgGround.Bounds().Dy()
	for x := 0; x < MapWidth; x += w {
		for y := 0; y < MapHeight; y += h {
			// Culling: Only draw visible tiles
			screenX := float64(x) - g.cameraX
			screenY := float64(y) - g.cameraY

			if screenX < -float64(w) || screenX > 320 || screenY < -float64(h) || screenY > 240 {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(screenX, screenY)
			screen.DrawImage(ImgGround, op)
		}
	}

	// 2. Draw Resources
	for _, node := range g.resourceNodes {
		node.Draw(screen, g.cameraX, g.cameraY)
	}

	// 3. Draw Buildings
	for _, b := range g.buildings {
		b.Draw(screen, g.cameraX, g.cameraY)
	}

	// 4. Draw Units (Player + Enemy)
	allUnits := append(g.units, g.enemyUnits...)
	for _, unit := range allUnits {
		unit.Draw(screen, g.cameraX, g.cameraY)
	}

	// 5. Draw Ghost Building (Placement Mode)
	if g.isPlacingBuilding {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.ghostX-g.cameraX, g.ghostY-g.cameraY)

		// Color feedback: Red if blocked, transparent white if valid
		if !g.canBuildHere {
			op.ColorScale.Scale(1, 0, 0, 0.7) // Red
		} else {
			op.ColorScale.Scale(1, 1, 1, 0.5) // Transparent
		}
		screen.DrawImage(ImgBuilding, op)
	}

	// 6. Draw Selection Box
	if g.isDragging {
		mouseX, mouseY := ebiten.CursorPosition()
		// Convert drag start to screen coordinates
		startXScreen := float32(float64(g.dragStartX) - g.cameraX)
		startYScreen := float32(float64(g.dragStartY) - g.cameraY)

		w := float32(mouseX) - startXScreen
		h := float32(mouseY) - startYScreen

		vector.StrokeRect(screen, startXScreen, startYScreen, w, h, 1, color.RGBA{0, 255, 0, 255}, false)
	}

	// 7. UI / Debug Text
	resText := fmt.Sprintf("Resources: %d", g.playerResources)
	ebitenutil.DebugPrint(screen, resText)

	hintText := "WASD: Camera | [B]: Build Mode | [U]: Train Unit"
	ebitenutil.DebugPrintAt(screen, hintText, 0, 15)

	selectedCount := 0
	for _, u := range g.units {
		if u.isSelected {
			selectedCount++
		}
	}
	if selectedCount > 0 {
		selText := fmt.Sprintf("Selected: %d", selectedCount)
		ebitenutil.DebugPrintAt(screen, selText, 0, 30)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}
