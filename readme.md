# 2D RTS Game

A small-scale 2D Real-Time Strategy game inspired by classics like **C&C: Red Alert 2** and **Age of Empires**, developed using the **Ebitengine** framework in Go. This project serves as a learning experience for game development and a showcase of Go programming skills.

---

## Project Goals

### Phase 1: The Core Foundation (Current Focus)

The primary goal of this phase is to build a minimum viable product (MVP) with the fundamental RTS mechanics. This includes:

* **Basic Unit Management:** Mouse-based selection and movement for a single unit.
* **Resource System:** A simple resource collection and management loop.
* **Building & Production:** A system to build structures and produce units from them.
* **Combat:** A basic combat loop where units can attack and destroy enemies.

### Phase 2: Factions & Progression

Once the core foundation is stable, this phase will introduce more depth and unique gameplay elements.

* **Multiple Factions:** Implementation of distinct factions with unique units and buildings.
* **Unit Promotion System:** A C&C Generals-inspired system where units gain experience points (XP) and are promoted, unlocking special abilities or stat bonuses.
* **Strategic Upgrades:** An Age of Empires IV-style upgrade system tied to in-game milestones.

### Phase 3: The Tournament

This is a long-term goal to add a structured single-player experience.

* **In-Game Tournament Mode:** A single-player tournament tree where the player competes against AI opponents.

---

## Technologies

* **Go:** The primary programming language.
* **Ebitengine:** A powerful and simple 2D game library for Go.

---

## How to Run

1.  Make sure Go is installed on your system.
2.  Clone the repository:
    ```sh
    git clone [your-repo-url]
    cd [your-repo-folder]
    ```
3.  Run the game:
    ```sh
    go run main.go
    ```

---

## Contribution

This is a personal learning project, but any feedback or suggestions are welcome!