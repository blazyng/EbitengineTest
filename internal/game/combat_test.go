package game

import (
	"testing"
)

func TestCombatProjectileAndDamage(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to initialize game: %v", err)
	}

	attacker := NewFactionUnit(50, 50, 1, FactionUSA, UnitTypeInfantry)
	target := NewFactionUnit(100, 50, 2, FactionChina, UnitTypeInfantry)

	initialHealth := target.health
	g.projectiles = nil
	g.particles = nil

	// Fire weapon
	g.FireUnitWeapon(attacker, target)

	if len(g.projectiles) != 1 {
		t.Fatalf("Expected 1 projectile, got %d", len(g.projectiles))
	}
	if len(g.particles) == 0 {
		t.Errorf("Expected muzzle flash particles, got 0")
	}

	p := g.projectiles[0]
	if p.Damage != attacker.attackDamage {
		t.Errorf("Expected projectile damage %d, got %d", attacker.attackDamage, p.Damage)
	}

	// Update projectiles for several frames until impact
	dt := 1.0 / 60.0
	for i := 0; i < 30; i++ {
		g.updateProjectiles(dt)
		if len(g.projectiles) == 0 {
			break
		}
	}

	if target.health >= initialHealth {
		t.Errorf("Target health did not decrease after projectile impact! Initial: %d, current: %d", initialHealth, target.health)
	}
}

func TestSpecialistRocketExplosion(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to initialize game: %v", err)
	}

	specialist := NewFactionUnit(50, 50, 1, FactionUSA, UnitTypeSpecialist)
	target := NewFactionUnit(150, 50, 2, FactionChina, UnitTypeWorker)

	g.projectiles = nil
	g.particles = nil

	g.FireUnitWeapon(specialist, target)

	if len(g.projectiles) != 1 {
		t.Fatalf("Expected rocket projectile")
	}
	if g.projectiles[0].ProjType != ProjectileRocket {
		t.Errorf("Expected ProjectileRocket, got %v", g.projectiles[0].ProjType)
	}

	dt := 1.0 / 60.0
	// Advance until rocket hits and triggers explosion particles
	for i := 0; i < 60; i++ {
		g.updateProjectiles(dt)
		g.updateParticles(dt)
		if len(g.projectiles) == 0 {
			break
		}
	}

	hasExplosion := false
	for _, part := range g.particles {
		if part.Type == ParticleExplosion {
			hasExplosion = true
			break
		}
	}
	if !hasExplosion {
		t.Errorf("Expected ParticleExplosion to be created upon rocket impact")
	}
}

func TestAutoGuardTargetAcquisition(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("Failed to initialize game: %v", err)
	}

	marine := NewFactionUnit(100, 100, 1, FactionUSA, UnitTypeInfantry)
	marine.state = StateIdle
	enemy := NewFactionUnit(130, 100, 2, FactionChina, UnitTypeInfantry)

	g.units = []*Unit{marine}
	g.enemyUnits = []*Unit{enemy}

	// Run 1 update tick
	err = g.Update()
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if marine.state != StateAttacking || marine.targetEnemy != enemy {
		t.Errorf("Expected marine to auto-acquire nearby enemy! State: %v, target: %v", marine.state, marine.targetEnemy)
	}
}
