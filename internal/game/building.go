package game

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// BuildingType distinguishes functionality, cost, and visuals of structures
type BuildingType int

const (
	BuildingHQ BuildingType = iota
	BuildingBarracks
	BuildingTurret
	BuildingSupply
)

// BuildingConfig stores metadata, costs, and dimensions
type BuildingConfig struct {
	Type        BuildingType
	Name        string
	Cost        int
	Width       float64
	Height      float64
	MaxHealth   int
	Description string
	Hotkey      string
}

// GetBuildingConfig returns the configuration for a building type
func GetBuildingConfig(bType BuildingType) BuildingConfig {
	switch bType {
	case BuildingHQ:
		return BuildingConfig{
			Type:        BuildingHQ,
			Name:        "Headquarters",
			Cost:        400,
			Width:       64,
			Height:      64,
			MaxHealth:   1000,
			Description: "Command Center & Resource Drop-Off",
			Hotkey:      "H",
		}
	case BuildingBarracks:
		return BuildingConfig{
			Type:        BuildingBarracks,
			Name:        "Barracks",
			Cost:        100,
			Width:       64,
			Height:      64,
			MaxHealth:   500,
			Description: "Trains Infantry and Specialists",
			Hotkey:      "1",
		}
	case BuildingTurret:
		return BuildingConfig{
			Type:        BuildingTurret,
			Name:        "Defense Turret",
			Cost:        150,
			Width:       48,
			Height:      48,
			MaxHealth:   400,
			Description: "Automated perimeter defense",
			Hotkey:      "2",
		}
	case BuildingSupply:
		return BuildingConfig{
			Type:        BuildingSupply,
			Name:        "Supply Depot",
			Cost:        75,
			Width:       48,
			Height:      48,
			MaxHealth:   300,
			Description: "Generates +10 Gold periodically",
			Hotkey:      "3",
		}
	default:
		return BuildingConfig{
			Type:      BuildingBarracks,
			Name:      "Barracks",
			Cost:      100,
			Width:     64,
			Height:    64,
			MaxHealth: 500,
		}
	}
}

type Building struct {
	x, y               float64
	width, height      float64
	buildingType       BuildingType
	name               string
	health             int
	maxHealth          int
	isBuilding         bool    // Is currently producing a unit?
	buildProgress      float64 // Construction progress (0.0 - 1.0)
	productionProgress float64 // Unit production progress (0.0 - 1.0)
	producingType      UnitType
	productionTime     float64
	team               int
	faction            FactionType
	rallyPointX        float64
	rallyPointY        float64
	isSelected         bool

	// Passive income for Supply Depot
	supplyTimer float64

	// Defense Turret combat
	attackRange    float64
	attackDamage   int
	attackSpeed    float64
	attackTimer    float64
	targetEnemy    *Unit
	turretAimX     float64
	turretAimY     float64
	shootAnimTimer float64
}

// NewSpecificBuilding initializes a building of the given type
func NewSpecificBuilding(x, y float64, team int, faction FactionType, bType BuildingType) *Building {
	cfg := GetBuildingConfig(bType)
	b := &Building{
		x:                  x,
		y:                  y,
		width:              cfg.Width,
		height:             cfg.Height,
		buildingType:       bType,
		name:               cfg.Name,
		health:             cfg.MaxHealth,
		maxHealth:          cfg.MaxHealth,
		isBuilding:         false,
		buildProgress:      0,
		productionProgress: 0,
		producingType:      UnitTypeWorker,
		productionTime:     4.0,
		team:               team,
		faction:            faction,
		rallyPointX:        x + cfg.Width + 12,
		rallyPointY:        y + cfg.Height/2,
		turretAimX:         1.0,
		turretAimY:         0.0,
	}

	if bType == BuildingTurret {
		b.attackRange = 150.0
		b.attackDamage = 22
		b.attackSpeed = 1.4
	}

	return b
}

func NewFactionBuilding(x, y float64, team int, faction FactionType) *Building {
	return NewSpecificBuilding(x, y, team, faction, BuildingHQ)
}

func NewBuilding(x, y float64) *Building {
	return NewSpecificBuilding(x, y, 1, FactionUSA, BuildingBarracks)
}

func (b *Building) BoundingBox() image.Rectangle {
	return image.Rect(int(b.x), int(b.y), int(b.x+b.width), int(b.y+b.height))
}

func (b *Building) Draw(screen *ebiten.Image, camX, camY, zoom float64) {
	screenX := float32((b.x - camX) * zoom)
	screenY := float32((b.y - camY) * zoom)
	bw := float32(b.width * zoom)
	bh := float32(b.height * zoom)

	// Viewport culling
	if screenX+bw < 0 || screenX > float32(ViewWidth) || screenY+bh < 0 || screenY > float32(ViewHeight) {
		return
	}

	// 1. Render Base Structure Sprite
	op := &ebiten.DrawImageOptions{}
	// Scale to building dimensions
	imgW := float64(ImgBuilding.Bounds().Dx())
	imgH := float64(ImgBuilding.Bounds().Dy())
	scaleX := (b.width / imgW) * zoom
	scaleY := (b.height / imgH) * zoom
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(float64(screenX), float64(screenY))

	// Team / Faction tint
	if b.team == 2 {
		op.ColorScale.Scale(1.3, 0.6, 0.6, 1) // Enemy red tint
	} else {
		switch b.buildingType {
		case BuildingHQ:
			op.ColorScale.Scale(1.0, 1.0, 1.2, 1) // HQ steel-blue
		case BuildingBarracks:
			op.ColorScale.Scale(0.8, 1.1, 0.8, 1) // Barracks military-green
		case BuildingTurret:
			op.ColorScale.Scale(1.1, 0.9, 0.7, 1) // Turret hardened-grey/sand
		case BuildingSupply:
			op.ColorScale.Scale(1.2, 1.1, 0.7, 1) // Supply gold-industrial
		}
	}

	// Construction transparency
	if b.buildProgress < 1.0 {
		op.ColorScale.Scale(1, 1, 1, 0.5)
	}
	screen.DrawImage(ImgBuilding, op)

	// 2. Type-Specific Details
	centerX := screenX + bw/2.0
	centerY := screenY + bh/2.0

	switch b.buildingType {
	case BuildingHQ:
		// Radar antenna dish at top-right
		radarX := screenX + bw*0.75
		radarY := screenY + bh*0.25
		vector.DrawFilledCircle(screen, radarX, radarY, float32(4.0*zoom), color.RGBA{220, 220, 240, 255}, false)
		vector.StrokeLine(screen, radarX, radarY, radarX-float32(3*zoom), radarY-float32(5*zoom), float32(1.5*zoom), color.RGBA{180, 180, 200, 255}, false)
		// Blinking LED
		vector.DrawFilledCircle(screen, radarX-float32(3*zoom), radarY-float32(5*zoom), float32(1.5*zoom), color.RGBA{0, 255, 120, 255}, false)

	case BuildingBarracks:
		// Military hangar door stripe
		doorW := bw * 0.4
		doorH := bh * 0.35
		vector.DrawFilledRect(screen, centerX-doorW/2, screenY+bh-doorH, doorW, doorH, color.RGBA{40, 45, 50, 240}, false)
		vector.StrokeRect(screen, centerX-doorW/2, screenY+bh-doorH, doorW, doorH, float32(1.0*zoom), color.RGBA{255, 200, 50, 200}, false)

	case BuildingTurret:
		// Fortified Turret Bunker Base
		turretRad := float32(12.0 * zoom)
		vector.DrawFilledCircle(screen, centerX, centerY, turretRad, color.RGBA{70, 75, 80, 255}, false)
		vector.StrokeCircle(screen, centerX, centerY, turretRad, float32(1.5*zoom), color.RGBA{120, 125, 130, 255}, false)

		// Rotating Twin Barrels
		aimX := float32(b.turretAimX)
		aimY := float32(b.turretAimY)
		barrelLen := float32(14.0 * zoom)
		perpX := -aimY * float32(3.0*zoom)
		perpY := aimX * float32(3.0*zoom)

		// Left and Right barrels
		b1StartX := centerX + perpX
		b1StartY := centerY + perpY
		b1EndX := b1StartX + aimX*barrelLen
		b1EndY := b1StartY + aimY*barrelLen

		b2StartX := centerX - perpX
		b2StartY := centerY - perpY
		b2EndX := b2StartX + aimX*barrelLen
		b2EndY := b2StartY + aimY*barrelLen

		barrelWidth := float32(2.0 * zoom)
		if barrelWidth < 1.5 {
			barrelWidth = 1.5
		}
		vector.StrokeLine(screen, b1StartX, b1StartY, b1EndX, b1EndY, barrelWidth, color.RGBA{30, 30, 35, 255}, false)
		vector.StrokeLine(screen, b2StartX, b2StartY, b2EndX, b2EndY, barrelWidth, color.RGBA{30, 30, 35, 255}, false)

		// Turret central dome
		vector.DrawFilledCircle(screen, centerX, centerY, float32(6.0*zoom), color.RGBA{100, 105, 110, 255}, false)

		// Muzzle flash when firing
		if b.shootAnimTimer > 0 {
			vector.DrawFilledCircle(screen, b1EndX, b1EndY, float32(3.5*zoom), color.RGBA{255, 220, 100, 255}, false)
			vector.DrawFilledCircle(screen, b2EndX, b2EndY, float32(3.5*zoom), color.RGBA{255, 220, 100, 255}, false)
		}

	case BuildingSupply:
		// Silo crate containers
		crateW := bw * 0.28
		crateH := bh * 0.28
		vector.DrawFilledRect(screen, centerX-crateW-float32(2*zoom), centerY-crateH/2, crateW, crateH, color.RGBA{200, 160, 40, 255}, false)
		vector.DrawFilledRect(screen, centerX+float32(2*zoom), centerY-crateH/2, crateW, crateH, color.RGBA{180, 140, 30, 255}, false)
	}

	// 3. Draw Unit Production Bar (Green)
	if b.isBuilding {
		barWidth := bw * float32(b.productionProgress)
		vector.DrawFilledRect(screen, screenX, screenY+bh+2, bw, 3, color.RGBA{40, 40, 40, 220}, false)
		vector.DrawFilledRect(screen, screenX, screenY+bh+2, barWidth, 3, color.RGBA{0, 255, 100, 255}, false)
	}

	// 4. Draw Construction Progress Bar (Yellow, only if under construction)
	if b.buildProgress < 1.0 && b.buildProgress > 0 {
		barWidth := bw * float32(b.buildProgress)
		vector.DrawFilledRect(screen, screenX, screenY-7, bw, 3, color.RGBA{40, 40, 40, 220}, false)
		vector.DrawFilledRect(screen, screenX, screenY-7, barWidth, 3, color.RGBA{255, 210, 0, 255}, false)
	}

	// 5. Draw Health Bar if damaged
	if b.health < b.maxHealth && b.buildProgress >= 1.0 {
		hpRatio := float32(b.health) / float32(b.maxHealth)
		vector.DrawFilledRect(screen, screenX, screenY-6, bw, 3, color.RGBA{200, 40, 40, 220}, false)
		vector.DrawFilledRect(screen, screenX, screenY-6, bw*hpRatio, 3, color.RGBA{40, 220, 40, 255}, false)
	}

	// 6. Draw selection border & range indicator
	if b.isSelected {
		vector.StrokeRect(screen, screenX, screenY, bw, bh, 2, color.RGBA{0, 255, 0, 255}, false)

		// Turret range circle
		if b.buildingType == BuildingTurret {
			rangeRad := float32(b.attackRange * zoom)
			vector.StrokeCircle(screen, centerX, centerY, rangeRad, float32(1.0*zoom), color.RGBA{255, 200, 50, 100}, false)
		}
	}
}

// DrawGhostBuilding renders a placement ghost with size matching the building type and range preview
func (g *Game) DrawGhostBuilding(screen *ebiten.Image, bType BuildingType, gx, gy float64, valid bool, camX, camY, zoom float64) {
	cfg := GetBuildingConfig(bType)
	screenX := float32((gx - camX) * zoom)
	screenY := float32((gy - camY) * zoom)
	bw := float32(cfg.Width * zoom)
	bh := float32(cfg.Height * zoom)

	// Ghost background
	fillCol := color.RGBA{0, 255, 100, 80}
	strokeCol := color.RGBA{0, 255, 100, 220}
	if !valid {
		fillCol = color.RGBA{255, 50, 50, 80}
		strokeCol = color.RGBA{255, 50, 50, 220}
	}

	vector.DrawFilledRect(screen, screenX, screenY, bw, bh, fillCol, false)
	vector.StrokeRect(screen, screenX, screenY, bw, bh, float32(2.0*zoom), strokeCol, false)

	// Draw footprint grid cross
	vector.StrokeLine(screen, screenX, screenY, screenX+bw, screenY+bh, float32(1.0*zoom), strokeCol, false)
	vector.StrokeLine(screen, screenX+bw, screenY, screenX, screenY+bh, float32(1.0*zoom), strokeCol, false)

	// Range circle for turret ghost
	if bType == BuildingTurret {
		rangeRad := float32(150.0 * zoom)
		centerX := screenX + bw/2.0
		centerY := screenY + bh/2.0
		vector.StrokeCircle(screen, centerX, centerY, rangeRad, float32(1.0*zoom), color.RGBA{255, 220, 80, 120}, false)
	}

	// Label
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s (%dg)", cfg.Name, cfg.Cost), int(screenX), int(screenY-12))
}
