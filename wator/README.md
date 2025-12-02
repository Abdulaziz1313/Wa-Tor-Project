# Wa-Tor Ecological Simulation in Go

## Author

- **Name:** Abdulaziz Aloufi  
- **Student ID:** C00266252  
- **Course:** BSc (Hons) in Software Development – SETU Carlow  

---

This project is an implementation of the classic **Wa-Tor** predator–prey simulation, where  
**fish** and **sharks** live, move, reproduce, and die on a **toroidal (wrap-around) grid**.

The rules are based on:

> A.K. Dewdney, “Computer Recreations; Sharks and Fish wage an ecological war on the toroidal planet of Wa-Tor”,  
> *Scientific American*, pp. 14–22.

The program supports both:

- **Text mode** – coloured ASCII output in the terminal  
- **Graphics mode** – real-time animation using Ebiten (https://ebitengine.org/)

It is implemented in **Go**, runs on **Linux**, and is documented with **Doxygen**.

---

## 1. Features

- Configurable simulation parameters:
  - Number of sharks and fish
  - Breed times for fish and sharks
  - Shark starvation time
  - Grid size
  - Number of threads (goroutines) to use
- Two execution modes:
  - **Text mode** (default)
  - **Graphics mode** using Ebiten (`-graphics=true`)
- **Toroidal world** (edges wrap around)
- **Concurrency**: a parallel update mode using multiple goroutines
- Optional **CSV output** of population counts per step
- **Doxygen** documentation (generated into `docs/`)

---

## 2. Building and Running (Linux)

### 2.1 Requirements

- Go (Go toolchain installed – e.g. via `sudo apt install golang-go`)
- For **graphics mode** (Ebiten) on Linux, you need X11 / GL dev libraries, for example:

```bash
sudo apt update
sudo apt install -y \
  libx11-dev \
  libxrandr-dev \
  libxi-dev \
  libxinerama-dev \
  libxcursor-dev \
  libxxf86vm-dev \
  libgl1-mesa-dev
```

### 2.2 Run in text mode (terminal)

From the project folder:

```bash
go run . -gridSize=20 -numFish=80 -numShark=20 -steps=200 -printEvery=10
```

Example behaviour:

- Prints a summary of the parameters
- Every `printEvery` steps: coloured ASCII grid, plus fish and shark counts
- Final simulation time, for example:

```text
Simulation finished in 243.6341ms
```

### 2.3 Run with CSV stats (for graphs)

```bash
go run . \
  -gridSize=50 \
  -numFish=800 \
  -numShark=200 \
  -steps=1000 \
  -printEvery=0 \
  -csv=stats_1thread.csv \
  -threads=1
```

This creates `stats_1thread.csv` with columns:

```text
step,fish,sharks
0,800,200
1,...
...
```

### 2.4 Run in graphics mode (Ebiten window)

```bash
go run . \
  -gridSize=50 \
  -numFish=800 \
  -numShark=200 \
  -steps=1000 \
  -threads=4 \
  -printEvery=0 \
  -graphics=true
```

Graphics mode shows:

- A window titled **“Wa-Tor Simulation”**
- Dark blue background = water  
- Cyan squares = fish  
- Orange squares = sharks  
- HUD text at the top: current step, total steps, fish count, shark count  

---

## 3. Command-line Parameters

The program accepts the following parameters:

- `-numShark int`  
  Starting population of sharks.  
  **Default:** `100`

- `-numFish int`  
  Starting population of fish.  
  **Default:** `200`

- `-fishBreed int`  
  Number of chronons before a fish can reproduce.  
  Once the counter reaches this value, the fish reproduces when it moves: it leaves a new fish in its old cell and its own breed counter resets to `0`.  
  **Default:** `3`

- `-sharkBreed int`  
  Number of chronons before a shark can reproduce, in the same way as the fish.  
  **Default:** `5`

- `-starve int`  
  Number of chronons a shark can survive without food.  
  Sharks lose 1 energy each chronon; when energy reaches 0 they die.  
  Eating a fish resets energy back to this value.  
  **Default:** `3`

- `-gridSize int`  
  World dimensions (`N×N`). The grid is **toroidal** (wrap-around).  
  **Default:** `20`

- `-threads int`  
  Number of goroutines to use in the parallel update step.  
  - `1` = fully sequential (`World.Step()`)  
  - `>1` = parallel (`World.StepParallel(threads)`)  
  **Default:** `1`

- `-steps int`  
  Number of chronons (time steps) to run.  
  **Default:** `200`

- `-printEvery int`  
  In text mode, how often to print the grid.  
  - `0` = never print the grid  
  - `10` = print every 10 steps  
  **Default:** `20`

- `-csv string`  
  Optional CSV file path to write population counts.  
  If empty, no CSV is written.

- `-graphics` (boolean flag)  
  - `false` = text mode (terminal)  
  - `true` = graphics mode (Ebiten window)  

---

## 4. Simulation Rules (Implementation Summary)

The code follows the classic Wa-Tor rules.

### 4.1 Fish

At each chronon, a fish:

- Looks at the 4 neighbours (N, E, S, W) with toroidal wrapping.
- Moves to a random empty neighbour (if any).
- If there are no free neighbours, it stays in place.

**Reproduction:**

- Each fish has a `BreedCounter` (chronons since last reproduction).
- When `BreedCounter >= FishBreed` **and** it moves:
  - A new fish is left behind in its old cell.
  - Both the parent and the new fish start with `BreedCounter = 0`.

### 4.2 Sharks

At each chronon, a shark:

- First looks for adjacent fish.
  - If there are any, it moves to a random fish cell and eats the fish.
- If no fish are adjacent, it moves like a fish to a random empty cell (if any).
- If there are no fish and no empty cells, it stays in place.

**Energy / starvation:**

- Each shark has an `Energy` counter.
- Every chronon: `Energy--`.
- If `Energy <= 0`, the shark dies.
- If it eats a fish, `Energy` is reset to `Starve`.

**Reproduction:**

- Each shark has a `BreedCounter`.
- When `BreedCounter >= SharkBreed` and it actually **moves**:
  - A new shark with full energy is left behind in the original position.
  - The parent’s `BreedCounter` resets to `0`.

---

## 5. Concurrency and Speedup

Two main update functions:

- `World.Step()` – sequential update.
- `World.StepParallel(threads int)` – divides the grid rows into chunks;  
  each goroutine updates its own rows into a private grid, then the results are merged.

To measure speedup, you can use a larger problem size, for example:

```bash
go run . \
  -gridSize=200 \
  -numFish=8000 \
  -numShark=2000 \
  -steps=2000 \
  -printEvery=0 \
  -threads=1

go run . ... -threads=2
go run . ... -threads=4
go run . ... -threads=8
```

Record the `Simulation finished in ...` time for each run.  
A simple table of `Threads` vs `Time` and a graph of `Threads` vs `Speedup` can then be created (see `RESULT.md` for my measurements and graphs).

---

## 6. Documentation (Doxygen)

Doxygen configuration is stored in `Doxyfile`.

To generate the documentation:

```bash
doxygen Doxyfile
```

Then open:

- `docs/html/index.html` – main entry point for the generated docs.

The Go source files (`main.go`, `world.go`, `graphics.go`) contain comments suitable for Doxygen, documenting, for example:

- Types: `Params`, `World`, `Creature`, `Game`
- Core functions: `NewWorld`, `Step`, `StepParallel`, `RunSimulation`, `RunSimulationGraphics`, etc.

---

## 7. Project Structure

```text
wator/
├── Doxyfile
├── go.mod
├── go.sum
├── main.go        
├── world.go       
├── graphics.go    
├── README.md
├── RESULT.md      
├── docs/          
└── (runtime.png, speedup.png) 
```

---

## 8. Example Usage Summary

```bash
# Simple text simulation
go run . -gridSize=20 -numFish=80 -numShark=20 -steps=200 -printEvery=10

# Graphics mode, larger world
go run . -gridSize=50 -numFish=800 -numShark=200 -steps=1000 -threads=4 -graphics=true

# Text mode with CSV output for later plotting
go run . -gridSize=50 -numFish=800 -numShark=200 -steps=1000 -printEvery=0 -threads=1 -csv=stats_1thread.csv
```
