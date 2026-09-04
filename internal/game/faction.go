package game

import "image/color"

// FactionType identifies a playable or AI faction
type FactionType int

const (
	FactionUSA FactionType = iota
	FactionChina
	FactionGBA
	FactionEU
	FactionAnime
	FactionAliens
	FactionZealots
)

// UnitType identifies the class/role of a unit
type UnitType int

const (
	UnitTypeWorker UnitType = iota
	UnitTypeInfantry
	UnitTypeSpecialist
)

// UnitConfig holds stats and behaviors for a unit type
type UnitConfig struct {
	Type          UnitType
	Name          string
	Role          string
	Description   string
	Cost          int
	BuildTime     float64
	MaxHealth     int
	AttackDamage  int
	AttackRange   float64
	AttackSpeed   float64
	Speed         float64
	CanHarvest    bool
	CanBuild      bool
	CargoCapacity int
	BadgeColor    color.RGBA
}

// FactionInfo stores metadata, theme, and unit roster for a faction
type FactionInfo struct {
	Type        FactionType
	Name        string
	Theme       string
	Color       color.RGBA
	UnitConfigs map[UnitType]UnitConfig
}

// Factions registry containing all configured factions from factions.md
var Factions = map[FactionType]FactionInfo{
	FactionUSA: {
		Type:  FactionUSA,
		Name:  "USA",
		Theme: "Lasers, Drones & Air Superiority",
		Color: color.RGBA{50, 120, 240, 255},
		UnitConfigs: map[UnitType]UnitConfig{
			UnitTypeWorker: {
				Type:          UnitTypeWorker,
				Name:          "M.U.L.E. Drone",
				Role:          "Builder / Harvester",
				Description:   "Automated logistics drone. Gathers gold and constructs buildings.",
				Cost:          50,
				BuildTime:     4.0,
				MaxHealth:     80,
				AttackDamage:  5,
				AttackRange:   30.0,
				AttackSpeed:   1.0,
				Speed:         2.2,
				CanHarvest:    true,
				CanBuild:      true,
				CargoCapacity: 10,
				BadgeColor:    color.RGBA{255, 215, 0, 255}, // Gold
			},
			UnitTypeInfantry: {
				Type:          UnitTypeInfantry,
				Name:          "Marine",
				Role:          "Standard Infantry",
				Description:   "Combat soldier armed with assault rifle. Rapid fire.",
				Cost:          70,
				BuildTime:     5.0,
				MaxHealth:     120,
				AttackDamage:  15,
				AttackRange:   70.0,
				AttackSpeed:   1.5,
				Speed:         2.0,
				CanHarvest:    false,
				CanBuild:      false,
				CargoCapacity: 0,
				BadgeColor:    color.RGBA{60, 140, 255, 255}, // Blue
			},
			UnitTypeSpecialist: {
				Type:          UnitTypeSpecialist,
				Name:          "Javelin Soldier",
				Role:          "Anti-Armor Rocket",
				Description:   "Heavy soldier with guided missile launcher. High damage.",
				Cost:          110,
				BuildTime:     8.0,
				MaxHealth:     90,
				AttackDamage:  35,
				AttackRange:   110.0,
				AttackSpeed:   0.6,
				Speed:         1.6,
				CanHarvest:    false,
				CanBuild:      false,
				CargoCapacity: 0,
				BadgeColor:    color.RGBA{200, 80, 255, 255}, // Purple
			},
		},
	},
	FactionChina: {
		Type:  FactionChina,
		Name:  "China",
		Theme: "Napalm, Propaganda & Mass Infantry",
		Color: color.RGBA{220, 50, 40, 255},
		UnitConfigs: map[UnitType]UnitConfig{
			UnitTypeWorker: {
				Type:          UnitTypeWorker,
				Name:          "Construction Dozer",
				Role:          "Builder / Dozer",
				Description:   "Heavily armored builder dozer. High health and capacity.",
				Cost:          50,
				BuildTime:     4.5,
				MaxHealth:     120,
				AttackDamage:  5,
				AttackRange:   25.0,
				AttackSpeed:   0.8,
				Speed:         1.8,
				CanHarvest:    true,
				CanBuild:      true,
				CargoCapacity: 12,
				BadgeColor:    color.RGBA{255, 180, 0, 255}, // Amber
			},
			UnitTypeInfantry: {
				Type:          UnitTypeInfantry,
				Name:          "Conscript",
				Role:          "Mass Infantry",
				Description:   "Inexpensive front-line infantry. Fast to train.",
				Cost:          45,
				BuildTime:     3.5,
				MaxHealth:     85,
				AttackDamage:  12,
				AttackRange:   65.0,
				AttackSpeed:   1.3,
				Speed:         2.1,
				CanHarvest:    false,
				CanBuild:      false,
				CargoCapacity: 0,
				BadgeColor:    color.RGBA{255, 60, 60, 255}, // Red
			},
			UnitTypeSpecialist: {
				Type:          UnitTypeSpecialist,
				Name:          "Tank Buster",
				Role:          "Anti-Armor Bazooka",
				Description:   "Fires high-explosive rockets at enemy targets.",
				Cost:          100,
				BuildTime:     7.0,
				MaxHealth:     80,
				AttackDamage:  32,
				AttackRange:   100.0,
				AttackSpeed:   0.7,
				Speed:         1.7,
				CanHarvest:    false,
				CanBuild:      false,
				CargoCapacity: 0,
				BadgeColor:    color.RGBA{255, 120, 30, 255}, // Orange
			},
		},
	},
	FactionGBA: {
		Type:  FactionGBA,
		Name:  "GBA",
		Theme: "Toxins, Explosives & Salvage",
		Color: color.RGBA{60, 180, 60, 255},
		UnitConfigs: map[UnitType]UnitConfig{
			UnitTypeWorker: {
				Type:          UnitTypeWorker,
				Name:          "Worker",
				Role:          "Laborer",
				Description:   "Resourceful laborer. Fast and inexpensive.",
				Cost:          40,
				BuildTime:     3.0,
				MaxHealth:     70,
				AttackDamage:  4,
				AttackRange:   25.0,
				AttackSpeed:   1.0,
				Speed:         2.2,
				CanHarvest:    true,
				CanBuild:      true,
				CargoCapacity: 10,
				BadgeColor:    color.RGBA{180, 220, 50, 255},
			},
			UnitTypeInfantry: {
				Type:          UnitTypeInfantry,
				Name:          "Rebel",
				Role:          "Guerrilla Fighter",
				Description:   "Rapid-firing guerrilla soldier.",
				Cost:          45,
				BuildTime:     3.0,
				MaxHealth:     75,
				AttackDamage:  11,
				AttackRange:   60.0,
				AttackSpeed:   1.6,
				Speed:         2.3,
				CanHarvest:    false,
				CanBuild:      false,
				CargoCapacity: 0,
				BadgeColor:    color.RGBA{80, 200, 80, 255},
			},
			UnitTypeSpecialist: {
				Type:          UnitTypeSpecialist,
				Name:          "RPG Soldier",
				Role:          "Rocket Soldier",
				Description:   "Fires makeshift RPG rockets with decent range.",
				Cost:          90,
				BuildTime:     6.0,
				MaxHealth:     75,
				AttackDamage:  30,
				AttackRange:   95.0,
				AttackSpeed:   0.7,
				Speed:         1.9,
				CanHarvest:    false,
				CanBuild:      false,
				CargoCapacity: 0,
				BadgeColor:    color.RGBA{220, 160, 40, 255},
			},
		},
	},
	FactionEU: {
		Type:  FactionEU,
		Name:  "EU",
		Theme: "EMP, Sonic Tech & Ballistics",
		Color: color.RGBA{80, 160, 220, 255},
		UnitConfigs: map[UnitType]UnitConfig{
			UnitTypeWorker: {
				Type:          UnitTypeWorker,
				Name:          "Construction Pioneer",
				Role:          "Pioneer",
				Description:   "Advanced construction operative.",
				Cost:          50,
				BuildTime:     4.0,
				MaxHealth:     90,
				AttackDamage:  6,
				AttackRange:   30.0,
				AttackSpeed:   1.0,
				Speed:         2.0,
				CanHarvest:    true,
				CanBuild:      true,
				CargoCapacity: 10,
				BadgeColor:    color.RGBA{100, 200, 255, 255},
			},
			UnitTypeInfantry: {
				Type:          UnitTypeInfantry,
				Name:          "Grenadier",
				Role:          "Assault Infantry",
				Description:   "Armed with assault rifle and ballistic armor.",
				Cost:          75,
				BuildTime:     5.0,
				MaxHealth:     130,
				AttackDamage:  16,
				AttackRange:   65.0,
				AttackSpeed:   1.4,
				Speed:         1.9,
				CanHarvest:    false,
				CanBuild:      false,
				CargoCapacity: 0,
				BadgeColor:    color.RGBA{80, 180, 255, 255},
			},
			UnitTypeSpecialist: {
				Type:          UnitTypeSpecialist,
				Name:          "Tank Hunter",
				Role:          "Smart Missile",
				Description:   "Fires high-tech smart homing missiles.",
				Cost:          115,
				BuildTime:     8.0,
				MaxHealth:     95,
				AttackDamage:  36,
				AttackRange:   105.0,
				AttackSpeed:   0.6,
				Speed:         1.6,
				CanHarvest:    false,
				CanBuild:      false,
				CargoCapacity: 0,
				BadgeColor:    color.RGBA{140, 100, 255, 255},
			},
		},
	},
	FactionAnime: {
		Type:  FactionAnime,
		Name:  "Anime Republic",
		Theme: "Magic, Kawaii Power & Mechs",
		Color: color.RGBA{255, 130, 200, 255},
		UnitConfigs: map[UnitType]UnitConfig{
			UnitTypeWorker: {
				Type:          UnitTypeWorker,
				Name:          "Chibi Builder",
				Role:          "Cute Builder",
				Description:   "Energetic chibi builder powered by friendship.",
				Cost:          45,
				BuildTime:     3.5,
				MaxHealth:     75,
				AttackDamage:  5,
				AttackRange:   25.0,
				AttackSpeed:   1.2,
				Speed:         2.4,
				CanHarvest:    true,
				CanBuild:      true,
				CargoCapacity: 10,
				BadgeColor:    color.RGBA{255, 180, 220, 255},
			},
			UnitTypeInfantry: {
				Type:          UnitTypeInfantry,
				Name:          "Kitsune Soldier",
				Role:          "Fox-Girl Trooper",
				Description:   "Agile fox-girl infantry with laser assault rifle.",
				Cost:          65,
				BuildTime:     4.5,
				MaxHealth:     105,
				AttackDamage:  14,
				AttackRange:   65.0,
				AttackSpeed:   1.6,
				Speed:         2.3,
				CanHarvest:    false,
				CanBuild:      false,
				CargoCapacity: 0,
				BadgeColor:    color.RGBA{255, 105, 180, 255},
			},
			UnitTypeSpecialist: {
				Type:          UnitTypeSpecialist,
				Name:          "Tsundere Rocketeer",
				Role:          "Tsundere Rocket",
				Description:   "Fires devastating magical fireworks. Baka!",
				Cost:          105,
				BuildTime:     7.5,
				MaxHealth:     85,
				AttackDamage:  34,
				AttackRange:   100.0,
				AttackSpeed:   0.65,
				Speed:         1.8,
				CanHarvest:    false,
				CanBuild:      false,
				CargoCapacity: 0,
				BadgeColor:    color.RGBA{255, 80, 150, 255},
			},
		},
	},
}

// GetUnitConfig retrieves the unit configuration for a given faction and unit type
func GetUnitConfig(faction FactionType, uType UnitType) UnitConfig {
	fac, ok := Factions[faction]
	if !ok {
		fac = Factions[FactionUSA]
	}
	cfg, ok := fac.UnitConfigs[uType]
	if !ok {
		cfg = Factions[FactionUSA].UnitConfigs[UnitTypeWorker]
	}
	return cfg
}

// GetFactionInfo retrieves the metadata and roster for a given faction
func GetFactionInfo(faction FactionType) FactionInfo {
	fac, ok := Factions[faction]
	if !ok {
		return Factions[FactionUSA]
	}
	return fac
}
