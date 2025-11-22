// updateUnits runs the state machine for every unit
package game

import (
	"image" //

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) updateUnits(units, enemies []*Unit) {
	dt := 1.0 / float64(ebiten.TPS())

	for _, unit := range units {
		// Update attack cooldown
		if unit.attackTimer > 0 {
			unit.attackTimer -= dt
		}

		// ---  ---
		// Checks if a unit *would* collide at a future position
		isCollidingAt := func(u *Unit, nextX, nextY float64) bool {
			// Create the "next step" bounding box
			nextBox := image.Rect(int(nextX), int(nextY), int(nextX+unitSize), int(nextY+unitSize))

			// 1. Check against barracks
			// 1. Check against ALL buildings
			for _, b := range g.buildings {
				if nextBox.Overlaps(b.BoundingBox()) {
					return true
				}
			}

			// 2. Check against resource nodes
			for _, node := range g.resourceNodes {
				// IMPORTANT: Don't check for collision with our *own target*!
				if u.targetNode == node {
					continue
				}
				// 1. Check against ALL buildings
				for _, b := range g.buildings {
					if nextBox.Overlaps(b.BoundingBox()) {
						return true
					}
				}
			}

			// 3. TODO: Check against other units (complex, skip for now)

			return false
		}
		// --- E

		switch unit.state {

		case StateIdle:
			// Do nothing

		case StateMovingToBuild:
			// Distanz zum Gebäude prüfen
			if unit.targetBuilding == nil {
				unit.state = StateIdle
				continue
			}

			dist := distance(unit.x, unit.y, unit.targetBuilding.x, unit.targetBuilding.y)
			buildRange := 50.0 // Genug Abstand halten

			if dist > buildRange {
				// Hinlaufen (Copy-Paste von Moving Logic)
				dx := unit.targetBuilding.x - unit.x
				dy := unit.targetBuilding.y - unit.y
				nextX := unit.x + (dx/dist)*unit.speed
				nextY := unit.y + (dy/dist)*unit.speed
				// --- NEW: Collision Check ---
				if !isCollidingAt(unit, nextX, nextY) {
					unit.x = nextX
					unit.y = nextY
				} else {
					// Stop if we hit something
					unit.state = StateIdle
				}
				// --- END: Collision Check ---

				unit.x = nextX
				unit.y = nextY
			} else {
				// Angekommen -> Anfangen zu bauen
				unit.state = StateBuilding
			}

		case StateBuilding:
			if unit.targetBuilding == nil || unit.targetBuilding.buildProgress >= 1.0 {
				unit.state = StateIdle
				unit.targetBuilding = nil
				continue
			}
			// Baufortschritt erhöhen
			// Sagen wir, eine Einheit baut 20% pro Sekunde
			unit.targetBuilding.buildProgress += (0.2 * dt)

			if unit.targetBuilding.buildProgress >= 1.0 {
				unit.targetBuilding.buildProgress = 1.0
				unit.state = StateIdle // Fertig!
			}

		case StateMoving:
			dist := distance(unit.x, unit.y, unit.targetX, unit.targetY)
			touchingDistance := 5.0 // How close to get before stopping

			if dist > touchingDistance { // Check if we are "close enough"
				// Calculate next step
				dx := unit.targetX - unit.x
				dy := unit.targetY - unit.y
				nextX := unit.x + (dx/dist)*unit.speed
				nextY := unit.y + (dy/dist)*unit.speed

				// --- NEW: Collision Check ---
				if !isCollidingAt(unit, nextX, nextY) {
					unit.x = nextX
					unit.y = nextY
				} else {
					// Stop if we hit something
					unit.state = StateIdle
				}
				// --- END: Collision Check ---

			} else {
				unit.x = unit.targetX
				unit.y = unit.targetY
				unit.state = StateIdle // Arrived at destination
			}

		case StateMovingToHarvest:
			dist := distance(unit.x, unit.y, unit.targetNode.x, unit.targetNode.y)
			touchingDistance := 10.0 // Stop a bit further away to "work"

			if dist > touchingDistance {
				// Calculate next step
				dx := unit.targetNode.x - unit.x
				dy := unit.targetNode.y - unit.y
				nextX := unit.x + (dx/dist)*unit.speed
				nextY := unit.y + (dy/dist)*unit.speed

				// --- NEW: Collision Check ---
				if !isCollidingAt(unit, nextX, nextY) {
					unit.x = nextX
					unit.y = nextY
				} else {
					// Stop if we hit something on the way
					unit.state = StateIdle
				}
				// --- END: Collision Check ---
			} else {
				// Arrived at the node
				unit.state = StateHarvesting
				unit.harvestTimer = unitHarvestTime // Start the harvest timer
			}

		case StateHarvesting:
			// Wait for the timer to finish
			unit.harvestTimer -= dt
			if unit.harvestTimer <= 0 {
				if unit.targetNode.amount > 0 {
					// Collect resources
					collected := unitCargoSize
					if unit.targetNode.amount < collected {
						collected = unit.targetNode.amount
					}

					unit.targetNode.amount -= collected
					unit.cargo = collected

					// Set target to base and go return
					unit.state = StateReturning
					unit.targetX = float64(g.basePosition.X)
					unit.targetY = float64(g.basePosition.Y)

					// If node is empty, clear target
					if unit.targetNode.amount <= 0 {
						unit.targetNode = nil
						// Note: We should also remove the node from g.resourceNodes
						// But we'll leave that for a later refactor
					}

				} else {
					// Node is empty, go idle
					unit.state = StateIdle
					unit.targetNode = nil
				}
			}

		case StateReturning:
			dist := distance(unit.x, unit.y, unit.targetX, unit.targetY)
			baseTouchingDistance := 10.0

			if dist > baseTouchingDistance {
				// Calculate next step
				dx := unit.targetX - unit.x
				dy := unit.targetY - unit.y
				nextX := unit.x + (dx/dist)*unit.speed
				nextY := unit.y + (dy/dist)*unit.speed

				// --- NEW: Collision Check ---
				if !isCollidingAt(unit, nextX, nextY) {
					unit.x = nextX
					unit.y = nextY
				} else {
					// Stop if we hit something
					unit.state = StateIdle
				}
				// --- END: Collision Check ---
			} else {
				// Arrived at base
				unit.x = unit.targetX
				unit.y = unit.targetY
				g.playerResources += unit.cargo
				unit.cargo = 0
				if unit.targetNode != nil && unit.targetNode.amount > 0 {
					unit.state = StateMovingToHarvest
					unit.targetX = unit.targetNode.x
					unit.targetY = unit.targetNode.y
				} else {
					unit.state = StateIdle
				}
			}

		case StateAttacking:
			if unit.targetEnemy == nil || unit.targetEnemy.health <= 0 {
				unit.state = StateIdle
				unit.targetEnemy = nil
				continue
			}

			dist := distance(unit.x, unit.y, unit.targetEnemy.x, unit.targetEnemy.y)

			if dist <= unit.attackRange {
				// 1. In range: Attack (no change)
				if unit.attackTimer <= 0 {
					unit.targetEnemy.health -= unit.attackDamage
					unit.attackTimer = 1.0 / unit.attackSpeed
				}
			} else {
				// 2. Out of range: Chase the target
				// Calculate next step
				dx := unit.targetEnemy.x - unit.x
				dy := unit.targetEnemy.y - unit.y
				nextX := unit.x + (dx/dist)*unit.speed
				nextY := unit.y + (dy/dist)*unit.speed

				// --- NEW: Collision Check ---
				if !isCollidingAt(unit, nextX, nextY) {
					unit.x = nextX
					unit.y = nextY
				} else {
				}
			}
		}
	}
}

// New function to remove dead units
func (g *Game) cleanupDeadUnits() {
	// Create new slices for living units
	livingPlayerUnits := make([]*Unit, 0, len(g.units))
	for _, unit := range g.units {
		if unit.health > 0 {
			livingPlayerUnits = append(livingPlayerUnits, unit)
		} else {
			// If a selected unit dies, we need to handle that,
			// but for now, just remove it.
		}
	}
	g.units = livingPlayerUnits // Replace old slice

	livingEnemyUnits := make([]*Unit, 0, len(g.enemyUnits))
	for _, unit := range g.enemyUnits {
		if unit.health > 0 {
			livingEnemyUnits = append(livingEnemyUnits, unit)
		}
	}
	g.enemyUnits = livingEnemyUnits // Replace old slice
}

func (g *Game) updateBuildings() {
	dt := 1.0 / float64(ebiten.TPS())

	for _, b := range g.buildings {
		// Schritt 1: Ist das Gebäude selbst noch im Bau?
		// Wenn buildProgress unter 1.0 ist, ist es noch ein "Gerüst".
		if b.buildProgress < 1.0 {
			continue // Überspringe dieses Gebäude, es kann noch nichts tun.
		}

		// Schritt 2: Produziert das fertige Gebäude gerade eine Einheit?
		if b.isBuilding {
			// Wir nutzen jetzt die NEUE Variable productionProgress
			b.productionProgress += dt / unitBuildTime

			// Wenn die EINHEIT fertig ist (100%)
			if b.productionProgress >= 1.0 {
				b.isBuilding = false
				b.productionProgress = 0.0

				// Neue Einheit spawnen (Team 1)
				// Wir nutzen b.rallyPointX statt g.barracks...
				newUnit := NewUnit(b.rallyPointX, b.rallyPointY, 1)
				g.units = append(g.units, newUnit)
			}
		}
	}
}
