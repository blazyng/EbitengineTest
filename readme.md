# 2D RTS Game

A small-scale 2D Real-Time Strategy game inspired by classics like **C&C: Red Alert 2** and **Age of Empires**, developed using the **Ebitengine** framework in Go. This project serves as a learning experience for game development and a showcase of Go programming skills.

![Current State](./game.png)


---

## Project Roadmap

### Phase 1: The Core Foundation (Completed ✔️)

This phase built the minimum viable product (MVP) with all fundamental RTS mechanics in place.

* `[x]` **Core Engine:** Setup Ebitengine, game loop, and window.
* `[x]` **Code Structure:** Refactored core logic from `main.go` into a clean `internal/game` package (e.g., `input.go`, `update.go`, `unit.go`).
* `[x]` **Unit Management:** Mouse-based selection (single click and drag-box) and movement (right-click).
* `[x]` **State Machine:** Implemented a robust state machine for units (`StateIdle`, `StateMoving`, `StateHarvesting`, etc.).
* `[x]` **Resource System:** A complete resource loop: units harvest from nodes, carry resources, and return to a drop-off point, updating a global resource counter.
* `[x]` **Building & Production:** A system to spend resources (via key-press) to produce new units from a building over time.
* `[x]` **Combat System:** Units have teams, HP, and stats. They can attack-move, engage enemies, deal damage, and are removed upon death. Includes health bars.
* `[x]` **Basic Collision:** Implemented "bounding box" collision detection to make buildings and resource nodes "solid" objects.

---

### Phase 2: The "Real" RTS Loop (Current Focus)

This phase focuses on upgrading the core systems from "functional" to "fun" and implementing proper game mechanics.

* `[x]` **Basic UI/HUD:** Interactive bottom HUD panel, top resource bar, tactical minimap with viewport & jump/move navigation, selection inspector (units & buildings), and clickable command buttons.
* `[x]` **Worker-Based Building:**
    * `[x]` Allow workers (current unit) to be commanded to build structures.
    * `[x]` Implement a building-placement mode (show a "ghost" building on the cursor).
    * `[x]` Create new unit states (`StateMovingToBuild`, `StateBuilding`).
* `[ ]` **Tiled Map System:** Replace the simple tiled texture with a multi-layered tilemap (e.g., using `tmx` files or varied terrain like water/cliffs).
* `[x]` **A* Pathfinding:** Robust 8-directional A* grid with C-space obstacle inflation, line-of-sight raycasting smoothing, visual waypoint lines for selected units, and multi-unit formation movement.
* `[x]` **Camera Controls:** Implement camera scrolling (WASD, arrow keys, and mouse edge-scrolling) clamped to map boundaries.
* `[x]` **Advanced Building System & Real Build Menu:** Categorized HUD build menu, automated Defense Turrets with targeting and firing, and Supply Depots with passive gold income.

---

### Phase 3: Factions & Polish

Once the core loop is solid, this phase will introduce variety and audiovisual feedback.

* `[x]` **Faction System:** Implement a structure for multiple factions (USA, China, GBA, EU, Anime, etc.) based on `factions.md`.
* `[x]` **Unique Units & Roles:** Distinct unit classes (Workers, Infantry, Anti-Armor Specialists) with unique stats, roles (harvesting vs combat), training times, and faction visual badges.
* `[x]` **Combat FX & Projectiles:** Muzzle flashes, recoil animations, fast bullet tracers, propelled rockets with smoke particle trails, hit sparks, and fiery unit destruction explosions.
* `[ ]` **Sprite & Animation System:** Replace the colored squares with actual 2D sprites.
    * `[ ]` Implement sprite sheets for animations (e.g., walking, attacking, harvesting).
* `[ ]` **Audio System:** Implement basic sound effects (clicks, attacks, "unit ready") and background music.
* `[ ]` **Basic Skirmish AI:** Create a simple AI opponent that can build a base, harvest resources, and send attack waves.

---

### Phase 4: Strategic Depth

This phase adds the "progression" elements inspired by C&C and AoE.

* `[ ]` **Unit Promotion System:** A C&C Generals-inspired system where units gain experience points (XP) and are promoted, unlocking special abilities or stat bonuses.
* `[ ]` **Strategic Upgrades:** An Age of Empires-style upgrade system (e.g., "Iron Broadswords," "Faster Engines") researched at buildings.
* `[ ]` **Campaign Mode:** A single-player tournament tree or a simple, scripted linear campaign.

---

### Phase 5: The Long-Term Vision

These are long-term goals if the project continues to be fun and successful.

* `[ ]` **Map Editor:** A simple in-game or external tool to create and save maps.
* `[ ]` **Multiplayer:** Implement network code for basic 1v1 multiplayer matches.

---

### Phase x: The crazy ideas

These are just funny ideas.

* `[ ]` **Steam Workshop integration for mods**
* `[ ]` **A co-op campaign and heroes for every faction (Pope, Greek gods, Hatsune Miku)**
* `[ ]` **The option for contributors to create weekly co-op campaigns (e.g., a Jackie Chan co-op campaign or an official Hatsune Miku co-op campaign)**
* `[ ]` **LAN-Mode**
* `[ ]` **Official fun modes like TD or Maul (Warcraft)**


---

## Technologies

* **Go:** The primary programming language.
* **Ebitengine:** A powerful and simple 2D game library for Go.

---

## How to Run

1.  Make sure Go is installed on your system.
2.  Clone the repository:
    ```sh
    git clone https://github.com/blazyng/EbitengineTest.git
    cd [your-cloned-folder]
    ```
3.  Run the game:
    ```sh
    go run ./
    ```
    *(Note: `go run ./` is often more reliable than `go run main.go` in projects with multiple packages)*

---

## Controls

* **Camera Movement:** `W`, `A`, `S`, `D` or Arrow keys, or moving cursor to screen edges.
* **Camera Zoom:** Mouse Wheel (zooms into/out of cursor position), `+` / `-` (`PageUp` / `PageDown`), `0` or `Home` (resets to 1.0x).
* **Window & Fullscreen:** Resize window freely by dragging edges or maximizing; toggle fullscreen with `F11`.
* **Selection:** Left-click on units/buildings or drag-box to select multiple units. `Esc` clears selection.
* **Orders:** Right-click to move, attack enemies, harvest resources (workers only), or build (workers only).
* **Minimap:** Left-click/drag minimap to jump camera; Right-click minimap to issue move orders to selected units.
* **Commands:**
  * `B` or HUD Button `[Build]`: Open Build Menu (`Esc` to close/cancel).
    * `1`: Place Barracks (100g - Trains Troops & Specialists)
    * `2`: Place Defense Turret (150g - Auto-defense turret with twin cannons)
    * `3`: Place Supply Depot (75g - Generates +10g every 4s)
  * `S` or HUD Button `[Stop]`: Stop selected units immediately.
  * `U`: Train Worker (`M.U.L.E. Drone` / `Construction Dozer`).
  * `I`: Train Infantry (`Marine` / `Conscript`).
  * `O`: Train Anti-Armor Specialist (`Javelin Soldier` / `Tank Buster`).

---

## Contribution

This is a personal learning project, but any feedback or suggestions are welcome!
