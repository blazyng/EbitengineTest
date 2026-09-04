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
		if unit.shootAnimTimer > 0 {
			unit.shootAnimTimer -= dt
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

		// Helper: Moves unit along its path or toward destination with sliding collision
		stepAlongPath := func(u *Unit, targetX, targetY float64, ignoreBuilding *Building, ignoreNode *ResourceNode) bool {
			// If path is empty, find path to target
			if len(u.path) == 0 {
				u.path = g.FindPath(u.x, u.y, targetX, targetY, ignoreBuilding, ignoreNode)
			}

			// If still empty (e.g. adjacent or in direct reach)
			if len(u.path) == 0 {
				dist := distance(u.x, u.y, targetX, targetY)
				if dist <= math.Max(4.0, u.speed) {
					u.x = targetX
					u.y = targetY
					return true
				}
				dx := targetX - u.x
				dy := targetY - u.y
				nextX := u.x + (dx/dist)*u.speed
				nextY := u.y + (dy/dist)*u.speed
				return moveWithSliding(u, nextX, nextY)
			}

			// Follow next waypoint
			wp := u.path[0]
			wpDist := distance(u.x, u.y, wp.X, wp.Y)
			reachDist := math.Max(4.0, u.speed)

			if wpDist <= reachDist {
				u.x = wp.X
				u.y = wp.Y
				u.path = u.path[1:]
				return true
			}

			dx := wp.X - u.x
			dy := wp.Y - u.y
			step := math.Min(wpDist, u.speed)
			nextX := u.x + (dx/wpDist)*step
			nextY := u.y + (dy/wpDist)*step

			if !moveWithSliding(u, nextX, nextY) {
				// Blocked by dynamic collision (e.g. another unit or newly placed building)
				// Clear path to allow repath
				u.path = nil
				return false
			}
			return true
		}

		switch unit.state {

		case StateIdle:
			// Combat units auto-acquire targets within guard range
			if unit.attackDamage > 0 {
				guardRange := unit.attackRange + 40.0
				var closestEnemy *Unit
				minDist := guardRange
				for _, enemy := range enemies {
					if enemy.health <= 0 {
						continue
					}
					d := distance(unit.x, unit.y, enemy.x, enemy.y)
					if d < minDist {
						minDist = d
						closestEnemy = enemy
					}
				}
				if closestEnemy != nil {
					unit.state = StateAttacking
					unit.targetEnemy = closestEnemy
					unit.path = nil
					unit.stuckFrames = 0
				}
			}

		case StateMoving:
			dist := distance(unit.x, unit.y, unit.targetX, unit.targetY)
			touchingDistance := 5.0

			if dist > touchingDistance {
				if !stepAlongPath(unit, unit.targetX, unit.targetY, nil, nil) {
					unit.stuckFrames++
					if unit.stuckFrames > 30 {
						unit.state = StateIdle
						unit.path = nil
						unit.stuckFrames = 0
					}
				} else {
					unit.stuckFrames = 0
				}
			} else {
				unit.x = unit.targetX
				unit.y = unit.targetY
				unit.state = StateIdle
				unit.path = nil
				unit.stuckFrames = 0
			}

		case StateMovingToBuild:
			if unit.targetBuilding == nil {
				unit.state = StateIdle
				unit.path = nil
				continue
			}
			dist := distance(unit.x, unit.y, unit.targetBuilding.x, unit.targetBuilding.y)
			buildRange := 50.0

			if dist > buildRange {
				targetX := unit.targetBuilding.x + unit.targetBuilding.width/2.0
				targetY := unit.targetBuilding.y + unit.targetBuilding.height/2.0
				if !stepAlongPath(unit, targetX, targetY, unit.targetBuilding, nil) {
					unit.stuckFrames++
					if unit.stuckFrames > 45 {
						unit.state = StateIdle
						unit.path = nil
						unit.stuckFrames = 0
					}
				} else {
					unit.stuckFrames = 0
				}
			} else {
				unit.state = StateBuilding
				unit.path = nil
				unit.stuckFrames = 0
			}

		case StateBuilding:
			if unit.targetBuilding == nil || unit.targetBuilding.buildProgress >= 1.0 {
				unit.state = StateIdle
				unit.targetBuilding = nil
				unit.path = nil
				unit.stuckFrames = 0
				continue
			}
			// Construction speed: 20% per second
			unit.targetBuilding.buildProgress += (0.2 * dt)

			if unit.targetBuilding.buildProgress >= 1.0 {
				unit.targetBuilding.buildProgress = 1.0
				unit.state = StateIdle
				unit.path = nil
				unit.stuckFrames = 0
			}

		case StateMovingToHarvest:
			if unit.targetNode == nil || unit.targetNode.amount <= 0 {
				unit.state = StateIdle
				unit.targetNode = nil
				unit.path = nil
				unit.stuckFrames = 0
				continue
			}
			dist := distance(unit.x, unit.y, unit.targetNode.x, unit.targetNode.y)
			touchingDistance := 10.0

			if dist > touchingDistance {
				if !stepAlongPath(unit, unit.targetNode.x, unit.targetNode.y, nil, unit.targetNode) {
					unit.stuckFrames++
					if unit.stuckFrames > 45 {
						unit.state = StateIdle
						unit.path = nil
						unit.stuckFrames = 0
					}
				} else {
					unit.stuckFrames = 0
				}
			} else {
				unit.state = StateHarvesting
				unit.harvestTimer = unitHarvestTime
				unit.path = nil
				unit.stuckFrames = 0
			}

		case StateHarvesting:
			if unit.targetNode == nil || unit.targetNode.amount <= 0 {
				unit.state = StateIdle
				unit.targetNode = nil
				unit.path = nil
				unit.stuckFrames = 0
				continue
			}
			unit.harvestTimer -= dt
			if unit.harvestTimer <= 0 {
				if unit.targetNode.amount > 0 {
					capacity := unit.cargoCapacity
					if capacity <= 0 {
						capacity = unitCargoSize
					}
					collected := capacity
					if unit.targetNode.amount < collected {
						collected = unit.targetNode.amount
					}

					unit.targetNode.amount -= collected
					unit.cargo = collected

					unit.state = StateReturning
					unit.path = nil
					unit.stuckFrames = 0
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
					unit.path = nil
					unit.stuckFrames = 0
				}
			}

		case StateReturning:
			dist := distance(unit.x, unit.y, unit.targetX, unit.targetY)
			touchingDistance := 10.0
			if unit.targetBuilding != nil {
				touchingDistance = (unit.targetBuilding.width / 2.0) + 12.0
			}

			if dist > touchingDistance {
				if !stepAlongPath(unit, unit.targetX, unit.targetY, unit.targetBuilding, nil) {
					unit.stuckFrames++
					if unit.stuckFrames > 45 {
						unit.state = StateIdle
						unit.path = nil
						unit.stuckFrames = 0
					}
				} else {
					unit.stuckFrames = 0
				}
			} else {
				g.playerResources += unit.cargo
				unit.cargo = 0
				unit.targetBuilding = nil // Clear target building reference for collision
				unit.path = nil
				unit.stuckFrames = 0

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
				unit.path = nil
				unit.stuckFrames = 0
				continue
			}

			dist := distance(unit.x, unit.y, unit.targetEnemy.x, unit.targetEnemy.y)

			if dist <= unit.attackRange {
				unit.path = nil
				unit.stuckFrames = 0

				// Aim at target center
				dx := (unit.targetEnemy.x + unitSize/2.0) - (unit.x + unitSize/2.0)
				dy := (unit.targetEnemy.y + unitSize/2.0) - (unit.y + unitSize/2.0)
				d := math.Hypot(dx, dy)
				if d > 0 {
					unit.aimDirX = dx / d
					unit.aimDirY = dy / d
				}

				if unit.attackTimer <= 0 {
					g.FireUnitWeapon(unit, unit.targetEnemy)
					unit.attackTimer = 1.0 / unit.attackSpeed
				}
			} else {
				// Enemy might move, so if last waypoint is far from enemy, repath
				if len(unit.path) > 0 {
					lastPt := unit.path[len(unit.path)-1]
					if distance(lastPt.X, lastPt.Y, unit.targetEnemy.x, unit.targetEnemy.y) > unit.attackRange {
						unit.path = nil
					}
				}
				stepAlongPath(unit, unit.targetEnemy.x, unit.targetEnemy.y, nil, nil)
			}
		}
	}
}

func (g *Game) cleanupDeadUnits() {
	livingPlayerUnits := make([]*Unit, 0, len(g.units))
	for _, unit := range g.units {
		if unit.health > 0 {
			livingPlayerUnits = append(livingPlayerUnits, unit)
		} else {
			g.SpawnUnitDestruction(unit.x+unitSize/2.0, unit.y+unitSize/2.0, unit.team)
		}
	}
	g.units = livingPlayerUnits

	livingEnemyUnits := make([]*Unit, 0, len(g.enemyUnits))
	for _, unit := range g.enemyUnits {
		if unit.health > 0 {
			livingEnemyUnits = append(livingEnemyUnits, unit)
		} else {
			g.SpawnUnitDestruction(unit.x+unitSize/2.0, unit.y+unitSize/2.0, unit.team)
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
	if len(activeNodes) != len(g.resourceNodes) {
		g.InvalidatePathGrid()
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

		// 1. Turret Automated Defense
		if b.buildingType == BuildingTurret {
			if b.attackTimer > 0 {
				b.attackTimer -= dt
			}
			if b.shootAnimTimer > 0 {
				b.shootAnimTimer -= dt
			}

			var potentialTargets []*Unit
			if b.team == 1 {
				potentialTargets = g.enemyUnits
			} else {
				potentialTargets = g.units
			}

			bCenterX := b.x + b.width/2.0
			bCenterY := b.y + b.height/2.0
			var closestTarget *Unit
			minDist := b.attackRange

			for _, target := range potentialTargets {
				if target.health <= 0 {
					continue
				}
				d := distance(bCenterX, bCenterY, target.x+unitSize/2.0, target.y+unitSize/2.0)
				if d < minDist {
					minDist = d
					closestTarget = target
				}
			}

			if closestTarget != nil {
				dx := (closestTarget.x + unitSize/2.0) - bCenterX
				dy := (closestTarget.y + unitSize/2.0) - bCenterY
				d := math.Hypot(dx, dy)
				if d > 0 {
					b.turretAimX = dx / d
					b.turretAimY = dy / d
				}

				if b.attackTimer <= 0 {
					g.FireTurretWeapon(b, closestTarget)
					b.attackTimer = 1.0 / b.attackSpeed
				}
			}
		}

		// 2. Supply Depot Passive Gold Generation
		if b.buildingType == BuildingSupply && b.team == 1 {
			b.supplyTimer += dt
			if b.supplyTimer >= 4.0 { // +10 gold every 4 seconds
				g.playerResources += 10
				b.supplyTimer = 0
			}
		}

		// 3. Unit Production
		if b.isBuilding {
			buildTime := b.productionTime
			if buildTime <= 0 {
				buildTime = unitBuildTime
			}
			b.productionProgress += dt / buildTime

			if b.productionProgress >= 1.0 {
				b.isBuilding = false
				b.productionProgress = 0.0

				team := b.team
				if team == 0 {
					team = 1
				}
				faction := b.faction
				newUnit := NewFactionUnit(b.rallyPointX, b.rallyPointY, team, faction, b.producingType)
				if team == 1 {
					g.units = append(g.units, newUnit)
				} else {
					g.enemyUnits = append(g.enemyUnits, newUnit)
				}
			}
		}
	}
}
