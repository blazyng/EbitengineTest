package game

import "github.com/hajimehoshi/ebiten/v2"

// updateUnits runs the state machine for every unit
func (g *Game) updateUnits(units, enemies []*Unit) {
	dt := 1.0 / float64(ebiten.TPS())

	for _, unit := range units {
		// Update attack cooldown
		if unit.attackTimer > 0 {
			unit.attackTimer -= dt
		}

		switch unit.state {

		case StateIdle:
			// Do nothing

		case StateMoving:
			// This is your original movement logic
			dist := distance(unit.x, unit.y, unit.targetX, unit.targetY)
			if dist > unit.speed {
				dx := unit.targetX - unit.x
				dy := unit.targetY - unit.y
				unit.x += (dx / dist) * unit.speed
				unit.y += (dy / dist) * unit.speed
			} else {
				unit.x = unit.targetX
				unit.y = unit.targetY
				unit.state = StateIdle // Arrived at destination
			}

		case StateMovingToHarvest:
			// Move towards the target resource node
			dist := distance(unit.x, unit.y, unit.targetNode.x, unit.targetNode.y)
			if dist > unit.speed {
				dx := unit.targetNode.x - unit.x
				dy := unit.targetNode.y - unit.y
				unit.x += (dx / dist) * unit.speed
				unit.y += (dy / dist) * unit.speed
			} else {
				// Arrived at the node
				unit.x = unit.targetNode.x
				unit.y = unit.targetNode.y
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
			// Move towards the base
			dist := distance(unit.x, unit.y, unit.targetX, unit.targetY)
			if dist > unit.speed {
				dx := unit.targetX - unit.x
				dy := unit.targetY - unit.y
				unit.x += (dx / dist) * unit.speed
				unit.y += (dy / dist) * unit.speed
			} else {
				// Arrived at base
				unit.x = unit.targetX
				unit.y = unit.targetY

				// Drop off resources
				g.playerResources += unit.cargo
				unit.cargo = 0

				// Check if we should go back to harvesting
				if unit.targetNode != nil && unit.targetNode.amount > 0 {
					unit.state = StateMovingToHarvest
					unit.targetX = unit.targetNode.x
					unit.targetY = unit.targetNode.y
				} else {
					unit.state = StateIdle // Nothing to do
				}
			}
		case StateAttacking:
			if unit.targetEnemy == nil || unit.targetEnemy.health <= 0 {
				// Target is dead or gone
				unit.state = StateIdle
				unit.targetEnemy = nil
				continue
			}

			// Calculate distance to target
			dist := distance(unit.x, unit.y, unit.targetEnemy.x, unit.targetEnemy.y)

			if dist <= unit.attackRange {
				// 1. In range: Stop moving and attack
				// Stop moving

				if unit.attackTimer <= 0 {
					// Attack is ready
					unit.targetEnemy.health -= unit.attackDamage
					unit.attackTimer = 1.0 / unit.attackSpeed // Reset cooldown
				}
			} else {
				// 2. Out of range: Chase the target
				dx := unit.targetEnemy.x - unit.x
				dy := unit.targetEnemy.y - unit.y
				unit.x += (dx / dist) * unit.speed
				unit.y += (dy / dist) * unit.speed
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

// updateBuildings needs one small tweak:
func (g *Game) updateBuildings() {
	if g.barracks.isBuilding {
		dt := 1.0 / float64(ebiten.TPS())
		g.barracks.buildProgress += dt / unitBuildTime

		if g.barracks.buildProgress >= 1.0 {
			g.barracks.isBuilding = false
			g.barracks.buildProgress = 0.0

			// Create a new unit (Team 1)
			newUnit := NewUnit(g.barracks.rallyPointX, g.barracks.rallyPointY, 1) // Team 1
			g.units = append(g.units, newUnit)
		}
	}
}
