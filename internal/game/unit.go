package game

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type UnitState int

const (
	StateIdle UnitState = iota
	StateMoving
	StateMovingToHarvest
	StateHarvesting
	StateReturning
	StateAttacking
	StateMovingToBuild
	StateBuilding
)

type Unit struct {
	x, y             float64
	targetX, targetY float64
	speed            float64
	isSelected       bool

	state        UnitState
	targetNode   *ResourceNode
	cargo        int
	harvestTimer float64

	team          int
	faction       FactionType
	unitType      UnitType
	name          string
	role          string
	canHarvest    bool
	canBuild      bool
	cargoCapacity int

	health         int
	maxHealth      int
	attackDamage   int
	attackRange    float64
	attackSpeed    float64
	attackTimer    float64
	targetEnemy    *Unit
	targetBuilding *Building

	// Pathfinding
	path        []Point
	stuckFrames int

	// Combat Animation
	shootAnimTimer float64
	aimDirX        float64
	aimDirY        float64
}

const (
	unitSize        = 32.0
	unitHarvestTime = 3.0
	unitCargoSize   = 10
)

// NewFactionUnit creates a specialized unit belonging to a faction
func NewFactionUnit(x, y float64, team int, faction FactionType, uType UnitType) *Unit {
	cfg := GetUnitConfig(faction, uType)

	return &Unit{
		x:             x,
		y:             y,
		targetX:       x,
		targetY:       y,
		speed:         cfg.Speed,
		state:         StateIdle,
		team:          team,
		faction:       faction,
		unitType:      uType,
		name:          cfg.Name,
		role:          cfg.Role,
		canHarvest:    cfg.CanHarvest,
		canBuild:      cfg.CanBuild,
		cargoCapacity: cfg.CargoCapacity,
		health:        cfg.MaxHealth,
		maxHealth:     cfg.MaxHealth,
		attackDamage:  cfg.AttackDamage,
		attackRange:   cfg.AttackRange,
		attackSpeed:   cfg.AttackSpeed,
		attackTimer:   0.0,
	}
}

// NewUnit provides backward compatibility for tests and default unit creation
func NewUnit(x, y float64, team int) *Unit {
	faction := FactionUSA
	if team == 2 {
		faction = FactionChina
	}
	return NewFactionUnit(x, y, team, faction, UnitTypeWorker)
}

func (u *Unit) BoundingBox() image.Rectangle {
	return image.Rect(int(u.x), int(u.y), int(u.x+unitSize), int(u.y+unitSize))
}

// Draw renders the unit sprite relative to the camera with faction and role styling
func (u *Unit) Draw(screen *ebiten.Image, camX, camY, zoom float64) {
	screenX := float32((u.x - camX) * zoom)
	screenY := float32((u.y - camY) * zoom)
	uSize := float32(unitSize * zoom)

	// Viewport culling
	if screenX+uSize < 0 || screenX > float32(ViewWidth) || screenY+uSize < 0 || screenY > float32(ViewHeight) {
		return
	}

	// Render Sprite
	op := &ebiten.DrawImageOptions{}

	// Tinting logic based on Faction, Team, UnitType, and Cargo
	if u.cargo > 0 {
		// Carrying Gold -> Bright yellow
		op.ColorScale.Scale(1.1, 1.1, 0.2, 1)
	} else if u.team == 2 {
		// Enemy team (China / GBA)
		switch u.unitType {
		case UnitTypeWorker:
			op.ColorScale.Scale(1.1, 0.7, 0.4, 1) // Dozer / Work bronze
		case UnitTypeInfantry:
			op.ColorScale.Scale(1.3, 0.5, 0.5, 1) // Conscript / Crimson red
		case UnitTypeSpecialist:
			op.ColorScale.Scale(1.4, 0.4, 0.2, 1) // Tank Buster / Flame orange
		default:
			op.ColorScale.Scale(1, 0.5, 0.5, 1)
		}
	} else {
		// Player team (USA)
		switch u.unitType {
		case UnitTypeWorker:
			op.ColorScale.Scale(1.0, 1.0, 0.8, 1) // MULE Drone / Gold-amber
		case UnitTypeInfantry:
			op.ColorScale.Scale(0.6, 0.85, 1.3, 1) // Marine / Military blue
		case UnitTypeSpecialist:
			op.ColorScale.Scale(0.8, 0.6, 1.3, 1) // Javelin / Laser violet
		default:
			op.ColorScale.Scale(1, 1, 1, 1)
		}
	}

	recoilX := float32(0)
	recoilY := float32(0)
	if u.shootAnimTimer > 0 {
		recoilX = float32(-u.aimDirX * 2.5 * zoom)
		recoilY = float32(-u.aimDirY * 2.5 * zoom)
		op.ColorScale.Scale(1.3, 1.3, 1.3, 1) // Bright flash on firing unit body
	}

	op.GeoM.Scale(zoom, zoom)
	op.GeoM.Translate(float64(screenX+recoilX), float64(screenY+recoilY))
	screen.DrawImage(ImgUnit, op)

	// Muzzle flash on weapon barrel
	if u.shootAnimTimer > 0 && (u.aimDirX != 0 || u.aimDirY != 0) {
		barrelX := screenX + uSize/2.0 + float32(u.aimDirX*float64(uSize)*0.55)
		barrelY := screenY + uSize/2.0 + float32(u.aimDirY*float64(uSize)*0.55)
		vector.DrawFilledCircle(screen, barrelX, barrelY, float32(3.5*zoom), color.RGBA{255, 255, 180, 240}, false)
		vector.DrawFilledCircle(screen, barrelX, barrelY, float32(1.8*zoom), color.RGBA{255, 255, 255, 255}, false)
	}

	// Draw role badge pip at top right of sprite
	cfg := GetUnitConfig(u.faction, u.unitType)
	pipSize := float32(4 * zoom)
	if pipSize < 2 {
		pipSize = 2
	}
	vector.DrawFilledRect(screen, screenX+uSize-pipSize-1, screenY+1, pipSize, pipSize, cfg.BadgeColor, false)

	// Draw selection border (player only)
	if u.isSelected && u.team == 1 {
		vector.StrokeRect(screen, screenX, screenY, uSize, uSize, 2, color.RGBA{0, 255, 0, 255}, false)

		// Draw path waypoints line for selected player units
		if len(u.path) > 0 {
			prevX := float32((u.x + unitSize/2.0 - camX) * zoom)
			prevY := float32((u.y + unitSize/2.0 - camY) * zoom)
			lineCol := color.RGBA{0, 255, 180, 160}
			for _, pt := range u.path {
				currX := float32((pt.X + unitSize/2.0 - camX) * zoom)
				currY := float32((pt.Y + unitSize/2.0 - camY) * zoom)
				vector.StrokeLine(screen, prevX, prevY, currX, currY, float32(1.5*zoom), lineCol, false)
				vector.DrawFilledCircle(screen, currX, currY, float32(2.5*zoom), lineCol, false)
				prevX = currX
				prevY = currY
			}
		}
	}

	// Draw Health Bar
	if u.health < u.maxHealth {
		healthPercent := float32(u.health) / float32(u.maxHealth)
		barWidth := uSize

		// Red background
		vector.DrawFilledRect(screen, screenX, screenY-6, barWidth, 4, color.RGBA{255, 0, 0, 255}, false)
		// Green foreground
		vector.DrawFilledRect(screen, screenX, screenY-6, barWidth*healthPercent, 4, color.RGBA{0, 255, 0, 255}, false)
	}

	// Draw Harvesting Progress Bar (Yellow/Orange, under unit)
	if u.state == StateHarvesting && u.harvestTimer > 0 {
		progress := (unitHarvestTime - u.harvestTimer) / unitHarvestTime
		if progress < 0 {
			progress = 0
		}
		if progress > 1 {
			progress = 1
		}
		barWidth := uSize
		// Dark background
		vector.DrawFilledRect(screen, screenX, screenY+uSize+2, barWidth, 3, color.RGBA{50, 50, 50, 255}, false)
		// Yellow progress bar
		vector.DrawFilledRect(screen, screenX, screenY+uSize+2, barWidth*float32(progress), 3, color.RGBA{255, 200, 0, 255}, false)
	}
}
