package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ProjectileType distinguishes weapon projectile behavior and appearance
type ProjectileType int

const (
	ProjectileBullet ProjectileType = iota
	ProjectileRocket
)

// Projectile represents a traveling bullet or missile in flight
type Projectile struct {
	X, Y           float64
	StartX, StartY float64
	TargetX, TargetY float64
	TargetUnit     *Unit
	TargetBuilding *Building
	Damage         int
	Speed          float64
	ProjType       ProjectileType
	ShooterTeam    int
	DistanceMoved  float64
	TotalDistance  float64
	SmokeTimer     float64
	Color          color.RGBA
}

// ParticleType categorizes visual effects
type ParticleType int

const (
	ParticleMuzzleFlash ParticleType = iota
	ParticleSpark
	ParticleSmoke
	ParticleExplosion
	ParticleDebris
)

// Particle represents a short-lived visual effect
type Particle struct {
	X, Y       float64
	Vx, Vy     float64
	Life       float64
	MaxLife    float64
	Size       float64
	MaxSize    float64
	Color      color.RGBA
	Type       ParticleType
}

// FireUnitWeapon initiates a projectile attack with muzzle effects
func (g *Game) FireUnitWeapon(attacker *Unit, target *Unit) {
	if attacker == nil || target == nil {
		return
	}

	attacker.shootAnimTimer = 0.15

	srcX := attacker.x + unitSize/2.0
	srcY := attacker.y + unitSize/2.0
	dstX := target.x + unitSize/2.0
	dstY := target.y + unitSize/2.0

	dx := dstX - srcX
	dy := dstY - srcY
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		dist = 1
	}
	dirX := dx / dist
	dirY := dy / dist

	barrelDist := unitSize * 0.45
	muzzleX := srcX + dirX*barrelDist
	muzzleY := srcY + dirY*barrelDist

	var projType ProjectileType
	speed := 16.0
	projCol := color.RGBA{255, 230, 110, 255}

	if attacker.unitType == UnitTypeSpecialist {
		projType = ProjectileRocket
		speed = 7.0
		projCol = color.RGBA{255, 120, 30, 255}

		// Rocket launch flash and backblast smoke
		g.SpawnMuzzleFlash(muzzleX, muzzleY, dirX, dirY, color.RGBA{255, 190, 50, 255})
		g.SpawnSmokePuff(srcX-dirX*barrelDist, srcY-dirY*barrelDist, 6.0)
	} else {
		projType = ProjectileBullet
		speed = 15.0
		g.SpawnMuzzleFlash(muzzleX, muzzleY, dirX, dirY, color.RGBA{255, 255, 180, 255})
	}

	p := &Projectile{
		X:             muzzleX,
		Y:             muzzleY,
		StartX:        muzzleX,
		StartY:        muzzleY,
		TargetX:       dstX,
		TargetY:       dstY,
		TargetUnit:    target,
		Damage:        attacker.attackDamage,
		Speed:         speed,
		ProjType:      projType,
		ShooterTeam:   attacker.team,
		TotalDistance: dist,
		Color:         projCol,
	}
	g.projectiles = append(g.projectiles, p)
}

// FireTurretWeapon initiates high-velocity defense rounds from an automated defense turret
func (g *Game) FireTurretWeapon(turret *Building, target *Unit) {
	if turret == nil || target == nil {
		return
	}

	turret.shootAnimTimer = 0.12

	srcX := turret.x + turret.width/2.0
	srcY := turret.y + turret.height/2.0
	dstX := target.x + unitSize/2.0
	dstY := target.y + unitSize/2.0

	dx := dstX - srcX
	dy := dstY - srcY
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		dist = 1
	}
	dirX := dx / dist
	dirY := dy / dist

	turret.turretAimX = dirX
	turret.turretAimY = dirY

	muzzleDist := turret.width * 0.5
	muzzleX := srcX + dirX*muzzleDist
	muzzleY := srcY + dirY*muzzleDist

	// Double barrel muzzle flash
	g.SpawnMuzzleFlash(muzzleX, muzzleY, dirX, dirY, color.RGBA{255, 210, 60, 255})

	p := &Projectile{
		X:             muzzleX,
		Y:             muzzleY,
		StartX:        muzzleX,
		StartY:        muzzleY,
		TargetX:       dstX,
		TargetY:       dstY,
		TargetUnit:    target,
		Damage:        turret.attackDamage,
		Speed:         18.0, // High-velocity turret round
		ProjType:      ProjectileBullet,
		ShooterTeam:   turret.team,
		TotalDistance: dist,
		Color:         color.RGBA{255, 180, 50, 255},
	}
	g.projectiles = append(g.projectiles, p)
}

// SpawnMuzzleFlash creates a directional burst of light at weapon barrel
func (g *Game) SpawnMuzzleFlash(x, y, dirX, dirY float64, col color.RGBA) {
	p := &Particle{
		X:       x,
		Y:       y,
		Vx:      dirX * 0.5,
		Vy:      dirY * 0.5,
		Life:    0.08,
		MaxLife: 0.08,
		Size:    6.0,
		MaxSize: 9.0,
		Color:   col,
		Type:    ParticleMuzzleFlash,
	}
	g.particles = append(g.particles, p)
}

// SpawnBulletSparks spawns high-velocity ricochet sparks on impact
func (g *Game) SpawnBulletSparks(x, y float64, count int) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * 2 * math.Pi
		spd := 1.0 + rand.Float64()*2.5
		g.particles = append(g.particles, &Particle{
			X:       x,
			Y:       y,
			Vx:      math.Cos(angle) * spd,
			Vy:      math.Sin(angle) * spd,
			Life:    0.12 + rand.Float64()*0.1,
			MaxLife: 0.22,
			Size:    2.0,
			MaxSize: 3.0,
			Color:   color.RGBA{255, 240, 150, 255},
			Type:    ParticleSpark,
		})
	}
}

// SpawnSmokePuff creates a rising, expanding smoke puff
func (g *Game) SpawnSmokePuff(x, y float64, size float64) {
	g.particles = append(g.particles, &Particle{
		X:       x,
		Y:       y,
		Vx:      (rand.Float64() - 0.5) * 0.4,
		Vy:      -0.3 - rand.Float64()*0.3,
		Life:    0.35 + rand.Float64()*0.2,
		MaxLife: 0.55,
		Size:    size,
		MaxSize: size * 2.2,
		Color:   color.RGBA{180, 180, 180, 180},
		Type:    ParticleSmoke,
	})
}

// SpawnRocketExplosion creates a fiery blast with shockwave ring and debris
func (g *Game) SpawnRocketExplosion(x, y, radius float64) {
	// Core fireball
	g.particles = append(g.particles, &Particle{
		X:       x,
		Y:       y,
		Life:    0.28,
		MaxLife: 0.28,
		Size:    radius * 0.6,
		MaxSize: radius * 1.5,
		Color:   color.RGBA{255, 140, 20, 240},
		Type:    ParticleExplosion,
	})

	// Sparks
	g.SpawnBulletSparks(x, y, 10)

	// Smoke billows
	for i := 0; i < 4; i++ {
		angle := rand.Float64() * 2 * math.Pi
		dist := rand.Float64() * radius * 0.5
		g.SpawnSmokePuff(x+math.Cos(angle)*dist, y+math.Sin(angle)*dist, 7.0)
	}

	// Flying debris chunks
	for i := 0; i < 6; i++ {
		angle := rand.Float64() * 2 * math.Pi
		spd := 1.5 + rand.Float64()*2.0
		g.particles = append(g.particles, &Particle{
			X:       x,
			Y:       y,
			Vx:      math.Cos(angle) * spd,
			Vy:      math.Sin(angle) * spd,
			Life:    0.3 + rand.Float64()*0.2,
			MaxLife: 0.5,
			Size:    3.0,
			MaxSize: 3.0,
			Color:   color.RGBA{60, 60, 60, 255},
			Type:    ParticleDebris,
		})
	}
}

// SpawnUnitDestruction triggers a violent explosion with fire and smoke upon unit death
func (g *Game) SpawnUnitDestruction(x, y float64, team int) {
	g.SpawnRocketExplosion(x, y, 20.0)

	// Team-tinted fiery wreckage
	debrisCol := color.RGBA{80, 80, 80, 255}
	if team == 2 {
		debrisCol = color.RGBA{150, 40, 40, 255}
	} else {
		debrisCol = color.RGBA{40, 80, 160, 255}
	}

	for i := 0; i < 8; i++ {
		angle := rand.Float64() * 2 * math.Pi
		spd := 2.0 + rand.Float64()*3.0
		g.particles = append(g.particles, &Particle{
			X:       x,
			Y:       y,
			Vx:      math.Cos(angle) * spd,
			Vy:      math.Sin(angle) * spd,
			Life:    0.4 + rand.Float64()*0.3,
			MaxLife: 0.7,
			Size:    3.5,
			MaxSize: 3.5,
			Color:   debrisCol,
			Type:    ParticleDebris,
		})
	}
}

// updateProjectiles moves active bullets/rockets and applies damage on impact
func (g *Game) updateProjectiles(dt float64) {
	var activeProjectiles []*Projectile

	for _, p := range g.projectiles {
		// Update target coordinates if tracking living unit
		if p.TargetUnit != nil && p.TargetUnit.health > 0 {
			p.TargetX = p.TargetUnit.x + unitSize/2.0
			p.TargetY = p.TargetUnit.y + unitSize/2.0
		}

		dx := p.TargetX - p.X
		dy := p.TargetY - p.Y
		dist := math.Hypot(dx, dy)

		if dist < 0.001 {
			dist = 0.001
		}
		dirX := dx / dist
		dirY := dy / dist

		step := math.Min(dist, p.Speed)
		p.X += dirX * step
		p.Y += dirY * step
		p.DistanceMoved += step

		// Spawn trailing smoke for rockets
		if p.ProjType == ProjectileRocket {
			p.SmokeTimer += dt
			if p.SmokeTimer >= 0.025 {
				g.SpawnSmokePuff(p.X-dirX*4.0, p.Y-dirY*4.0, 3.5)
				p.SmokeTimer = 0
			}
		}

		// Check for impact
		hit := dist <= step || p.DistanceMoved >= p.TotalDistance
		if hit {
			// Apply damage
			if p.TargetUnit != nil && p.TargetUnit.health > 0 {
				p.TargetUnit.health -= p.Damage
			}

			// Spawn impact visual FX
			if p.ProjType == ProjectileRocket {
				g.SpawnRocketExplosion(p.X, p.Y, 16.0)
			} else {
				g.SpawnBulletSparks(p.X, p.Y, 5)
			}
		} else {
			activeProjectiles = append(activeProjectiles, p)
		}
	}

	g.projectiles = activeProjectiles
}

// updateParticles advances lifetime, position, and physics of visual effects
func (g *Game) updateParticles(dt float64) {
	var activeParticles []*Particle

	for _, p := range g.particles {
		p.Life -= dt
		if p.Life <= 0 {
			continue
		}

		p.X += p.Vx * dt * 60.0
		p.Y += p.Vy * dt * 60.0
		p.Vx *= 0.94
		p.Vy *= 0.94

		activeParticles = append(activeParticles, p)
	}

	g.particles = activeParticles
}

// DrawCombat renders all in-flight projectiles and visual particle effects
func (g *Game) DrawCombat(screen *ebiten.Image, camX, camY, zoom float64) {
	// 1. Draw Particles
	for _, p := range g.particles {
		screenX := float32((p.X - camX) * zoom)
		screenY := float32((p.Y - camY) * zoom)

		// Viewport culling (320x190)
		if screenX < -30 || screenX > 350 || screenY < -30 || screenY > 220 {
			continue
		}

		lifeRatio := float32(p.Life / p.MaxLife)
		if lifeRatio < 0 {
			lifeRatio = 0
		}

		switch p.Type {
		case ParticleMuzzleFlash:
			sz := float32(p.Size * zoom)
			vector.DrawFilledCircle(screen, screenX, screenY, sz, p.Color, false)
			vector.DrawFilledCircle(screen, screenX, screenY, sz*0.5, color.RGBA{255, 255, 255, 255}, false)

		case ParticleSpark:
			sz := float32(p.Size * zoom)
			if sz < 1.5 {
				sz = 1.5
			}
			col := p.Color
			col.A = uint8(float32(col.A) * lifeRatio)
			vector.DrawFilledRect(screen, screenX-sz/2, screenY-sz/2, sz, sz, col, false)

		case ParticleSmoke:
			currSize := float32((p.Size + (p.MaxSize-p.Size)*(1.0-float64(lifeRatio))) * zoom)
			alpha := uint8(float32(p.Color.A) * lifeRatio * 0.8)
			smokeCol := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
			vector.DrawFilledCircle(screen, screenX, screenY, currSize, smokeCol, false)

		case ParticleExplosion:
			currRadius := float32((p.Size + (p.MaxSize-p.Size)*(1.0-float64(lifeRatio))) * zoom)
			// Outer flame
			flameAlpha := uint8(float32(p.Color.A) * lifeRatio)
			flameCol := color.RGBA{p.Color.R, p.Color.G, p.Color.B, flameAlpha}
			vector.DrawFilledCircle(screen, screenX, screenY, currRadius, flameCol, false)
			// Inner yellow core
			if lifeRatio > 0.3 {
				coreRadius := currRadius * 0.55
				coreAlpha := uint8(255 * (lifeRatio - 0.3) / 0.7)
				vector.DrawFilledCircle(screen, screenX, screenY, coreRadius, color.RGBA{255, 250, 160, coreAlpha}, false)
			}
			// Shockwave outline
			vector.StrokeCircle(screen, screenX, screenY, currRadius*1.15, float32(1.5*zoom), color.RGBA{255, 220, 100, flameAlpha / 2}, false)

		case ParticleDebris:
			sz := float32(p.Size * zoom)
			if sz < 2 {
				sz = 2
			}
			col := p.Color
			col.A = uint8(float32(col.A) * lifeRatio)
			vector.DrawFilledRect(screen, screenX-sz/2, screenY-sz/2, sz, sz, col, false)
		}
	}

	// 2. Draw Projectiles
	for _, p := range g.projectiles {
		screenX := float32((p.X - camX) * zoom)
		screenY := float32((p.Y - camY) * zoom)

		if screenX < -30 || screenX > 350 || screenY < -30 || screenY > 220 {
			continue
		}

		dx := p.TargetX - p.X
		dy := p.TargetY - p.Y
		dist := math.Hypot(dx, dy)
		if dist == 0 {
			dist = 1
		}
		dirX := float32(dx / dist)
		dirY := float32(dy / dist)

		switch p.ProjType {
		case ProjectileBullet:
			// Tracer streak
			tailLen := float32(8.0 * zoom)
			tailX := screenX - dirX*tailLen
			tailY := screenY - dirY*tailLen
			vector.StrokeLine(screen, tailX, tailY, screenX, screenY, float32(1.8*zoom), p.Color, false)
			// Glowing head dot
			vector.DrawFilledCircle(screen, screenX, screenY, float32(2.0*zoom), color.RGBA{255, 255, 200, 255}, false)

		case ProjectileRocket:
			// Rocket missile body
			bodyLen := float32(10.0 * zoom)
			tailX := screenX - dirX*bodyLen
			tailY := screenY - dirY*bodyLen
			vector.StrokeLine(screen, tailX, tailY, screenX, screenY, float32(3.0*zoom), color.RGBA{100, 100, 110, 255}, false)
			// Rocket warhead tip
			vector.DrawFilledCircle(screen, screenX, screenY, float32(2.5*zoom), color.RGBA{220, 60, 40, 255}, false)
			// Engine exhaust flare
			flameLen := float32(4.0 * zoom)
			flameX := tailX - dirX*flameLen
			flameY := tailY - dirY*flameLen
			vector.StrokeLine(screen, tailX, tailY, flameX, flameY, float32(2.2*zoom), color.RGBA{255, 200, 50, 255}, false)
		}
	}
}
