package game

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// handleInput processes user mouse clicks
func (g *Game) handleInput(screenMouseX, screenMouseY int) {

	mouseX := float64(screenMouseX) + g.cameraX
	mouseY := float64(screenMouseY) + g.cameraY

	if g.isPlacingBuilding {
		// Position aktualisieren für den "Ghost"
		g.ghostX = mouseX
		g.ghostY = mouseY

		// Linksklick zum Platzieren
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if g.playerResources >= 100 { // Nehmen wir an, ein Gebäude kostet 100
				g.playerResources -= 100

				// Neues Gebäude erstellen (Fundament)
				newB := NewBuilding(g.ghostX, g.ghostY)
				newB.buildProgress = 0.0 // Muss noch gebaut werden
				g.buildings = append(g.buildings, newB)

				// Ausgewählte Einheiten zum Bauen schicken
				for _, unit := range g.units {
					if unit.isSelected {
						unit.state = StateMovingToBuild
						unit.targetBuilding = newB
						// Wir zielen auf die Mitte des Gebäudes
						unit.targetX = newB.x + newB.width/2
						unit.targetY = newB.y + newB.height/2
					}
				}
				g.isPlacingBuilding = false // Modus beenden
			}
		}
		// Rechtsklick zum Abbrechen
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			g.isPlacingBuilding = false
		}
		return // Wenn wir platzieren, machen wir keine andere Input-Verarbeitung
	}

	// --- NEU: Taste 'B' startet den Bau-Modus ---
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.isPlacingBuilding = true
		// Selektion aufheben, damit man besser sieht
		g.isDragging = false
	}
	// --- Right Click (Commands) ---
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		clickedNode := false
		clickedEnemy := false

		// 1. Check if we clicked an enemy
		for _, enemy := range g.enemyUnits {
			if mouseX >= enemy.x && mouseX <= enemy.x+unitSize &&
				mouseY >= enemy.y && mouseY <= enemy.y+unitSize {

				// Clicked on an enemy! Send all selected units to attack.
				for _, unit := range g.units {
					if unit.isSelected {
						unit.state = StateAttacking
						unit.targetEnemy = enemy
						unit.targetNode = nil // Clear resource target
					}
				}
				clickedEnemy = true
				break
			}
		}
		if clickedEnemy {
			return // Don't process other right-click actions
		}
		for _, node := range g.resourceNodes {
			if mouseX >= node.x && mouseX <= node.x+node.width &&
				mouseY >= node.y && mouseY <= node.y+node.height {

				// Clicked on a node! Send all selected units to harvest it.
				for _, unit := range g.units {
					if unit.isSelected {
						unit.state = StateMovingToHarvest
						unit.targetNode = node
						unit.targetX = node.x // Target the node's position
						unit.targetY = node.y
					}
				}
				clickedNode = true
				break
			}
		}

		// If we didn't click a node, it's a normal move command
		if !clickedNode {
			for _, unit := range g.units {
				if unit.isSelected {
					unit.state = StateMoving
					unit.targetNode = nil // Not targeting a node
					unit.targetX = mouseX
					unit.targetY = mouseY
				}
			}
		}
	}

	// --- Left Click (Selection) ---
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.isDragging = true
		g.dragStartX, g.dragStartY = int(mouseX), int(mouseY)

		unitClicked := false
		for _, unit := range g.units {
			if mouseX >= unit.x && mouseX <= unit.x+unitSize &&
				mouseY >= unit.y && mouseY <= unit.y+unitSize {
				unitClicked = true
				break
			}
		}

		// Deselect all units (unless holding Shift, but we'll add that later)
		for _, unit := range g.units {
			unit.isSelected = false
		}

		if unitClicked {
			for _, unit := range g.units {
				if mouseX >= unit.x && mouseX <= unit.x+unitSize &&
					mouseY >= unit.y && mouseY <= unit.y+unitSize {
					unit.isSelected = true
					break // Only select one
				}
			}
		}
	}

	// --- Drag Selection ---
	if g.isDragging {
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			g.isDragging = false
			selectionRect := image.Rect(g.dragStartX, g.dragStartY, int(mouseX), int(mouseY)).Canon()

			for _, unit := range g.units {
				unitRect := image.Rect(int(unit.x), int(unit.y), int(unit.x)+int(unitSize), int(unit.y)+int(unitSize))
				if selectionRect.Overlaps(unitRect) {
					unit.isSelected = true
				}
			}
		}
	}
}

// Update auch handleProductionInput, da g.barracks weg ist
// internal/game/input.go

func (g *Game) handleProductionInput() {
	// Wenn U gedrückt wird
	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		// Wir loopen durch ALLE Gebäude
		for _, b := range g.buildings {
			// Wir suchen ein Gebäude, das:
			// 1. Fertig gebaut ist (buildProgress >= 1.0)
			// 2. Gerade NICHTS produziert (!isBuilding)
			// 3. (Optional) Wir könnten noch prüfen, ob es der richtige Typ ist (z.B. Kaserne)

			if b.buildProgress >= 1.0 && !b.isBuilding {
				// Haben wir genug Geld?
				if g.playerResources >= unitCost {
					g.playerResources -= unitCost

					// Starte Produktion IN DIESEM Gebäude
					b.isBuilding = true
					b.productionProgress = 0.0 // Reset Fortschritt

					// Wir brechen nach dem ersten Gebäude ab, damit nicht alle gleichzeitig bauen
					// (Außer du willst das – dann entfern das 'break')
					break
				}
			}
		}
	}
}

func (g *Game) updateCameraInput() {
	// Kamera-Geschwindigkeit
	speed := 5.0

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
}
