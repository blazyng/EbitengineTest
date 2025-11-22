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

type Game struct {
	units           []*Unit
	enemyUnits      []*Unit
	resourceNodes   []*ResourceNode
	buildings       []*Building
	playerResources int
	basePosition    image.Point

	isDragging bool
	dragStartX int
	dragStartY int

	isPlacingBuilding bool
	ghostX, ghostY    float64

	cameraX float64
	cameraY float64
}

const (
	unitCost      = 50
	unitBuildTime = 5.0
)

func NewGame() (*Game, error) {
	g := &Game{
		basePosition:    image.Point{X: 10, Y: 10},
		playerResources: 1000,
	}

	g.units = []*Unit{
		NewUnit(50, 50, 1),
		NewUnit(80, 80, 1),
	}

	g.enemyUnits = []*Unit{
		NewUnit(250, 100, 2),
		NewUnit(250, 140, 2),
	}

	g.resourceNodes = []*ResourceNode{
		NewResourceNode(200, 200, 1000),
	}

	startBarracks := NewBuilding(10, 100)
	startBarracks.buildProgress = 1.0
	g.buildings = []*Building{startBarracks}

	return g, nil
}

func distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()

	// 1. Camera Input
	g.updateCameraInput()

	// 2. Handle Input
	g.handleInput(mouseX, mouseY)
	g.handleProductionInput()

	// 3. Update World State
	g.updateUnits(g.units, g.enemyUnits)
	g.updateUnits(g.enemyUnits, g.units)
	g.updateBuildings()

	// 4. Cleanup
	g.cleanupDeadUnits()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 1. Base (offset by camera)
	vector.DrawFilledRect(screen, float32(g.basePosition.X-int(g.cameraX)), float32(g.basePosition.Y-int(g.cameraY)), 16, 16, color.White, false)

	// 2. Resources
	for _, node := range g.resourceNodes {
		node.Draw(screen, g.cameraX, g.cameraY)
	}

	// 3. Buildings
	for _, b := range g.buildings {
		b.Draw(screen, g.cameraX, g.cameraY)
	}

	// 4. Units
	allUnits := append(g.units, g.enemyUnits...)
	for _, unit := range allUnits {
		// THIS calls the function in unit.go for the specific unit
		unit.Draw(screen, g.cameraX, g.cameraY)
	}

	// 5. Ghost Building
	if g.isPlacingBuilding {
		vector.DrawFilledRect(screen, float32(g.ghostX-g.cameraX), float32(g.ghostY-g.cameraY), 64, 64, color.RGBA{255, 255, 255, 128}, false)
	}

	// 6. Selection Box
	if g.isDragging {
		mouseX, mouseY := ebiten.CursorPosition()
		// Convert DragStart back to Screen coordinates for drawing
		startXScreen := float32(float64(g.dragStartX) - g.cameraX)
		startYScreen := float32(float64(g.dragStartY) - g.cameraY)

		w := float32(mouseX) - startXScreen
		h := float32(mouseY) - startYScreen

		vector.StrokeRect(screen, startXScreen, startYScreen, w, h, 1, color.RGBA{0, 255, 0, 255}, false)
	}

	// 7. UI Text (Static on screen, no camera offset needed)
	resText := fmt.Sprintf("Resources: %d", g.playerResources)
	ebitenutil.DebugPrint(screen, resText)

	hintText := "Press [B] to build Building (100) | [U] to build Unit (50) | WASD for Camera"
	ebitenutil.DebugPrintAt(screen, hintText, 0, 15)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}
