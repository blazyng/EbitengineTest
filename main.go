package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	units []*Unit
}

type Unit struct {
	x          float64
	y          float64
	targetX    float64
	targetY    float64
	speed      float64
	isSelected bool // New: State to check if the unit is selected
}

func (g *Game) Update() error {
	// Check if the left mouse button is pressed
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()

		// Flag to check if a unit was clicked
		unitClicked := false
		for _, unit := range g.units {
			// Check if the click is within the unit's rectangle
			if mouseX >= int(unit.x) && mouseX <= int(unit.x)+32 &&
				mouseY >= int(unit.y) && mouseY <= int(unit.y)+32 {
				// If a unit was clicked, set the flag.
				unitClicked = true
			}
		}

		// If a unit was clicked, select it and deselect all others.
		if unitClicked {
			for _, unit := range g.units {
				if mouseX >= int(unit.x) && mouseX <= int(unit.x)+32 &&
					mouseY >= int(unit.y) && mouseY <= int(unit.y)+32 {
					// Select only the clicked unit.
					unit.isSelected = true
				} else {
					// Deselect all other units.
					unit.isSelected = false
				}
			}
		} else {
			// If no unit was clicked, but one is already selected, set its target to the cursor position.
			for _, unit := range g.units {
				if unit.isSelected {
					unit.targetX = float64(mouseX)
					unit.targetY = float64(mouseY)
				}
			}
		}
	}

	// PHASE 3: Execute movement for all units (regardless of clicks).
	for _, unit := range g.units {
		dx := unit.targetX - unit.x
		dy := unit.targetY - unit.y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance > unit.speed {
			unit.x += (dx / distance) * unit.speed
			unit.y += (dy / distance) * unit.speed
		} else {
			unit.x = unit.targetX
			unit.y = unit.targetY
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, unit := range g.units {
		// Draw the unit
		vector.DrawFilledRect(screen, float32(unit.x), float32(unit.y), 32, 32, color.RGBA{0, 0, 255, 255}, false)

		// If the unit is selected, draw a green border
		if unit.isSelected {
			vector.StrokeRect(screen, float32(unit.x), float32(unit.y), 32, 32, 2, color.RGBA{0, 255, 0, 255}, false)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	game := &Game{
		units: []*Unit{
			{x: 100, y: 100, targetX: 100, targetY: 100, speed: 2.0},
			{x: 200, y: 200, targetX: 200, targetY: 200, speed: 2.0},
		},
	}
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Mubi")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
