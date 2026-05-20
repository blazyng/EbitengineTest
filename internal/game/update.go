package game

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) updateUnits(units, enemies []*Unit) {
	dt := 1.0 / float64(ebiten.TPS())

	for _, unit := range units {
		if unit.attackTimer > 0 {
			unit.attackTimer -= dt
		}

		// Helper: Checks if a future position causes collision
		isCollidingAt := func(u *Unit, nextX, nextY float64) bool {
			nextBox := image.Rect(int(nextX), int(nextY), int(nextX+unitSize), int(nextY+unitSize))

			// Check against Buildings
			for _, b := range g.buildings {
				if u.targetBuilding == b {
					continue
				}
				if nextBox.Overlaps(b.BoundingBox()) {
					return true
				}
			}

			// Check against Resources
			for _, node := range g.resourceNodes {
				if u.targetNode == node {
					continue
				}
				if nextBox.Overlaps(node.BoundingBox()) {
					return true
				}
			}
			return false
		}

		// Unit-to-unit soft separation to prevent overlapping
		allUnits := append(g.units, g.enemyUnits...)
		for _, other := range allUnits {
			if unit == other {
				continue
			}
			dist := distance(unit.x, unit.y, other.x, other.y)
			minDist := unitSize - 4.0
			if dist < minDist {
				if dist == 0 {
					unit.x -= 0.5
					continue
				}
				dx := (unit.x - other.x) / dist
				dy := (unit.y - other.y) / dist
				force := (minDist - dist) / minDist * 0.8
				nextX := unit.x + dx*force
				nextY := unit.y + dy*force
				if !isCollidingAt(unit, nextX, nextY) {
					unit.x = nextX
					unit.y = nextY
				}
			}
		}

		// Helper: Moves unit with sliding collision response
		moveWithSliding := func(u *Unit, nextX, nextY float64) bool {
			if !isCollidingAt(u, nextX, nextY) {
				u.x = nextX
				u.y = nextY
				return true
			}
			// Try to slide along the X-axis
			if !isCollidingAt(u, nextX, u.y) {
				u.x = nextX
				return true
			}
			// Try to slide along the Y-axis
			if !isCollidingAt(u, u.x, nextY) {
				u.y = nextY
				return true
			}
			return false
		}

		switch unit.state {

		case StateIdle:
			// Do nothing

		case StateMoving:
			dist := distance(unit.x, unit.y, unit.targetX, unit.targetY)
			touchingDistance := 5.0

			if dist > touchingDistance {
				dx := unit.targetX - unit.x
				dy := unit.targetY - unit.y
				nextX := unit.x + (dx/dist)*unit.speed
				nextY := unit.y + (dy/dist)*unit.speed

				if !moveWithSliding(unit, nextX, nextY) {
					unit.state = StateIdle
				}
			} else {
				unit.x = unit.targetX
				unit.y = unit.targetY
				unit.state = StateIdle
			}

		case StateMovingToBuild:
			if unit.targetBuilding == nil {
				unit.state = StateIdle
				continue
			}
			dist := distance(unit.x, unit.y, unit.targetBuilding.x, unit.targetBuilding.y)
			buildRange := 50.0

			if dist > buildRange {
				dx := unit.targetBuilding.x - unit.x
				dy := unit.targetBuilding.y - unit.y
				nextX := unit.x + (dx/dist)*unit.speed
				nextY := unit.y + (dy/dist)*unit.speed

				if !moveWithSliding(unit, nextX, nextY) {
					unit.state = StateIdle
				}
			} else {
				unit.state = StateBuilding
			}

		case StateBuilding:
			if unit.targetBuilding == nil || unit.targetBuilding.buildProgress >= 1.0 {
				unit.state = StateIdle
				unit.targetBuilding = nil
				continue
			}
			// Construction speed: 20% per second
			unit.targetBuilding.buildProgress += (0.2 * dt)

			if unit.targetBuilding.buildProgress >= 1.0 {
				unit.targetBuilding.buildProgress = 1.0
				unit.state = StateIdle
			}

		case StateMovingToHarvest:
			if unit.targetNode == nil || unit.targetNode.amount <= 0 {
				unit.state = StateIdle
				unit.targetNode = nil
				continue
			}
			dist := distance(unit.x, unit.y, unit.targetNode.x, unit.targetNode.y)
			touchingDistance := 10.0

			if dist > touchingDistance {
				dx := unit.targetNode.x - unit.x
				dy := unit.targetNode.y - unit.y
				nextX := unit.x + (dx/dist)*unit.speed
				nextY := unit.y + (dy/dist)*unit.speed

				if !moveWithSliding(unit, nextX, nextY) {
					unit.state = StateIdle
				}
			} else {
				unit.state = StateHarvesting
				unit.harvestTimer = unitHarvestTime
			}

		case StateHarvesting:
			if unit.targetNode == nil || unit.targetNode.amount <= 0 {
				unit.state = StateIdle
				unit.targetNode = nil
				continue
			}
			unit.harvestTimer -= dt
			if unit.harvestTimer <= 0 {
				if unit.targetNode.amount > 0 {
					collected := unitCargoSize
					if unit.targetNode.amount < collected {
						collected = unit.targetNode.amount
					}

					unit.targetNode.amount -= collected
					unit.cargo = collected

					unit.state = StateReturning
					dropOff := g.findClosestDropOffBuilding(unit.x, unit.y)
					if dropOff != nil {
						unit.targetBuilding = dropOff
						unit.targetX = dropOff.x + dropOff.width/2
						unit.targetY = dropOff.y + dropOff.height/2
					} else {
						unit.targetBuilding = nil
						unit.targetX = float64(g.basePosition.X)
						unit.targetY = float64(g.basePosition.Y)
					}

					if unit.targetNode.amount <= 0 {
						unit.targetNode = nil
					}

				} else {
					unit.state = StateIdle
					unit.targetNode = nil
				}
			}

		case StateReturning:
			dist := distance(unit.x, unit.y, unit.targetX, unit.targetY)
			touchingDistance := 10.0
			if unit.targetBuilding != nil {
				touchingDistance = (unit.targetBuilding.width / 2.0) + 12.0
			}

			if dist > touchingDistance {
				dx := unit.targetX - unit.x
				dy := unit.targetY - unit.y
				nextX := unit.x + (dx/dist)*unit.speed
				nextY := unit.y + (dy/dist)*unit.speed

				if !moveWithSliding(unit, nextX, nextY) {
					unit.state = StateIdle
				}
			} else {
				g.playerResources += unit.cargo
				unit.cargo = 0
				unit.targetBuilding = nil // Clear target building reference for collision

				if unit.targetNode != nil && unit.targetNode.amount > 0 {
					unit.state = StateMovingToHarvest
					unit.targetX = unit.targetNode.x
					unit.targetY = unit.targetNode.y
				} else {
					// Search for a new node if targetNode is nil or depleted
					unit.targetNode = g.findClosestResourceNode(unit.x, unit.y)
					if unit.targetNode != nil && unit.targetNode.amount > 0 {
						unit.state = StateMovingToHarvest
						unit.targetX = unit.targetNode.x
						unit.targetY = unit.targetNode.y
					} else {
						unit.state = StateIdle
					}
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
				if unit.attackTimer <= 0 {
					unit.targetEnemy.health -= unit.attackDamage
					unit.attackTimer = 1.0 / unit.attackSpeed
				}
			} else {
				dx := unit.targetEnemy.x - unit.x
				dy := unit.targetEnemy.y - unit.y
				nextX := unit.x + (dx/dist)*unit.speed
				nextY := unit.y + (dy/dist)*unit.speed

				moveWithSliding(unit, nextX, nextY)
			}
		}
	}
}

func (g *Game) cleanupDeadUnits() {
	livingPlayerUnits := make([]*Unit, 0, len(g.units))
	for _, unit := range g.units {
		if unit.health > 0 {
			livingPlayerUnits = append(livingPlayerUnits, unit)
		}
	}
	g.units = livingPlayerUnits

	livingEnemyUnits := make([]*Unit, 0, len(g.enemyUnits))
	for _, unit := range g.enemyUnits {
		if unit.health > 0 {
			livingEnemyUnits = append(livingEnemyUnits, unit)
		}
	}
	g.enemyUnits = livingEnemyUnits
}

func (g *Game) cleanupDepletedResources() {
	activeNodes := make([]*ResourceNode, 0, len(g.resourceNodes))
	for _, node := range g.resourceNodes {
		if node.amount > 0 {
			activeNodes = append(activeNodes, node)
		}
	}
	g.resourceNodes = activeNodes
}

func (g *Game) findClosestResourceNode(x, y float64) *ResourceNode {
	var closest *ResourceNode
	minDist := math.MaxFloat64
	for _, node := range g.resourceNodes {
		if node.amount > 0 {
			dist := distance(x, y, node.x, node.y)
			if dist < minDist {
				minDist = dist
				closest = node
			}
		}
	}
	return closest
}

func (g *Game) findClosestDropOffBuilding(x, y float64) *Building {
	var closest *Building
	minDist := math.MaxFloat64
	for _, b := range g.buildings {
		if b.buildProgress >= 1.0 {
			dist := distance(x, y, b.x, b.y)
			if dist < minDist {
				minDist = dist
				closest = b
			}
		}
	}
	return closest
}

func (g *Game) updateBuildings() {
	dt := 1.0 / float64(ebiten.TPS())

	for _, b := range g.buildings {
		if b.buildProgress < 1.0 {
			continue // Building is under construction
		}

		if b.isBuilding {
			b.productionProgress += dt / unitBuildTime

			if b.productionProgress >= 1.0 {
				b.isBuilding = false
				b.productionProgress = 0.0

				newUnit := NewUnit(b.rallyPointX, b.rallyPointY, 1)
				g.units = append(g.units, newUnit)
			}
		}
	}
}
