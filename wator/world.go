//--------------------------------
//Author: Abdulaziz Hameed Aloufi
//Student ID: C00266252
//--------------------------------

package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

// CellType represents the contents of a grid cell: empty, fish or shark.
type CellType int

const (
	// Empty means there is no creature in the cell.
	Empty CellType = iota
	// FishCell means the cell is occupied by a fish.
	FishCell
	// SharkCell means the cell is occupied by a shark.
	SharkCell
)

// Creature represents either a fish or a shark living in the grid.
// It carries enough state to support breeding and, for sharks, energy.
type Creature struct {
	Kind         CellType // FishCell or SharkCell
	BreedCounter int      // chronons since last reproduction
	Energy       int      // used only for sharks; 0 for fish
}

// World holds the simulation grid and the parameters used to evolve it.
type World struct {
	Size   int
	Grid   [][]*Creature
	Params Params
}

// CellAt returns the CellType at coordinates (x, y). If the grid cell is
// nil, the cell is considered empty.
func (w *World) CellAt(x, y int) CellType {
	c := w.Grid[y][x]
	if c == nil {
		return Empty
	}
	return c.Kind
}

// NewWorld creates a new toroidal Wa-Tor world with randomly placed
// fish and sharks according to the given parameters.
func NewWorld(p Params) *World {
	rand.Seed(time.Now().UnixNano())

	w := &World{
		Size:   p.GridSize,
		Grid:   make([][]*Creature, p.GridSize),
		Params: p,
	}

	for y := 0; y < p.GridSize; y++ {
		w.Grid[y] = make([]*Creature, p.GridSize)
	}

	totalCells := p.GridSize * p.GridSize
	if p.NumFish+p.NumShark > totalCells {
		fmt.Println("Error: more creatures than cells in the grid")
		os.Exit(1)
	}

	// Create a random permutation of all cell indices.
	positions := rand.Perm(totalCells)
	idx := 0

	// Place fish.
	for i := 0; i < p.NumFish; i++ {
		pos := positions[idx]
		idx++
		x := pos % p.GridSize
		y := pos / p.GridSize

		w.Grid[y][x] = &Creature{
			Kind:         FishCell,
			BreedCounter: 0,
			Energy:       0,
		}
	}

	// Place sharks.
	for i := 0; i < p.NumShark; i++ {
		pos := positions[idx]
		idx++
		x := pos % p.GridSize
		y := pos / p.GridSize

		w.Grid[y][x] = &Creature{
			Kind:         SharkCell,
			BreedCounter: 0,
			Energy:       p.Starve, // start with full energy
		}
	}

	return w
}

// PrintColored prints an ANSI-coloured ASCII representation of the world.
// Fish are rendered in green, sharks in red and empty cells as blue dots.
func (w *World) PrintColored() {

	const (
		reset = "\033[0m"
		blue  = "\033[34m"
		green = "\033[32m"
		red   = "\033[31m"
	)

	for y := 0; y < w.Size; y++ {
		for x := 0; x < w.Size; x++ {
			c := w.Grid[y][x]
			if c == nil {
				fmt.Printf("%s.%s ", blue, reset)
				continue
			}
			switch c.Kind {
			case FishCell:
				fmt.Printf("%sf%s ", green, reset)
			case SharkCell:
				fmt.Printf("%sS%s ", red, reset)
			}
		}
		fmt.Println()
	}
}

// Print is a small convenience wrapper that prints the world using
// coloured ASCII output.
func (w *World) Print() {
	w.PrintColored()
}

// Count returns the total number of fish and sharks currently in the world.
func (w *World) Count() (fish int, sharks int) {
	for y := 0; y < w.Size; y++ {
		for x := 0; x < w.Size; x++ {
			c := w.Grid[y][x]
			if c == nil {
				continue
			}
			if c.Kind == FishCell {
				fish++
			} else if c.Kind == SharkCell {
				sharks++
			}
		}
	}
	return
}

// Step performs one chronon of the simulation sequentially. It first
// updates all fish, then all sharks, writing into a new grid and
// finally swaps the new grid into place.
func (w *World) Step() {
	newGrid := make([][]*Creature, w.Size)
	for y := 0; y < w.Size; y++ {
		newGrid[y] = make([]*Creature, w.Size)
	}

	// --- FISH PHASE ---
	for y := 0; y < w.Size; y++ {
		for x := 0; x < w.Size; x++ {
			c := w.Grid[y][x]
			if c == nil || c.Kind != FishCell {
				continue
			}
			w.updateFish(x, y, c, newGrid)
		}
	}

	// --- SHARK PHASE ---
	for y := 0; y < w.Size; y++ {
		for x := 0; x < w.Size; x++ {
			c := w.Grid[y][x]
			if c == nil || c.Kind != SharkCell {
				continue
			}
			w.updateShark(x, y, c, newGrid)
		}
	}

	w.Grid = newGrid
}

// StepParallel performs one chronon of the simulation using multiple
// goroutines. Each worker writes into its own private grid, and the
// grids are merged afterwards. When both a fish and a shark contend
// for the same cell, the shark wins.
func (w *World) StepParallel(threads int) {
	if threads <= 1 {
		w.Step()
		return
	}
	if threads > w.Size {
		threads = w.Size
	}

	// Each worker gets its own private newGrid.
	localGrids := make([][][]*Creature, threads)
	for t := 0; t < threads; t++ {
		grid := make([][]*Creature, w.Size)
		for y := 0; y < w.Size; y++ {
			grid[y] = make([]*Creature, w.Size)
		}
		localGrids[t] = grid
	}

	rowsPerThread := (w.Size + threads - 1) / threads
	var wg sync.WaitGroup

	for t := 0; t < threads; t++ {
		startY := t * rowsPerThread
		endY := startY + rowsPerThread
		if startY >= w.Size {
			break
		}
		if endY > w.Size {
			endY = w.Size
		}

		localGrid := localGrids[t]

		wg.Add(1)
		go func(startY, endY int, lg [][]*Creature) {
			defer wg.Done()

			// FISH PHASE on assigned rows (reading from shared w.Grid).
			for y := startY; y < endY; y++ {
				for x := 0; x < w.Size; x++ {
					c := w.Grid[y][x]
					if c == nil || c.Kind != FishCell {
						continue
					}
					w.updateFish(x, y, c, lg)
				}
			}

			// SHARK PHASE on assigned rows.
			for y := startY; y < endY; y++ {
				for x := 0; x < w.Size; x++ {
					c := w.Grid[y][x]
					if c == nil || c.Kind != SharkCell {
						continue
					}
					w.updateShark(x, y, c, lg)
				}
			}
		}(startY, endY, localGrid)
	}

	wg.Wait()

	// Merge local grids into final newGrid.
	newGrid := make([][]*Creature, w.Size)
	for y := 0; y < w.Size; y++ {
		row := make([]*Creature, w.Size)
		for x := 0; x < w.Size; x++ {
			var chosen *Creature
			for t := 0; t < threads; t++ {
				c := localGrids[t][y][x]
				if c == nil {
					continue
				}
				if chosen == nil {
					chosen = c
				} else if c.Kind == SharkCell && chosen.Kind == FishCell {
					// If both fish and shark want the same cell, shark wins.
					chosen = c
				}
				// If both are the same kind we just keep the first; that's fine.
			}
			row[x] = chosen
		}
		newGrid[y] = row
	}

	w.Grid = newGrid
}

// neighbours returns the 4-neighbour coordinates (N,E,S,W) around (x, y),
// using toroidal wrapping at the world boundaries.
func (w *World) neighbours(x, y int) [][2]int {
	return [][2]int{
		{x, (y - 1 + w.Size) % w.Size}, // north
		{(x + 1) % w.Size, y},          // east
		{x, (y + 1) % w.Size},          // south
		{(x - 1 + w.Size) % w.Size, y}, // west
	}
}

// emptyNeighbours returns the coordinates of neighbouring cells that are
// empty in the current grid.
func (w *World) emptyNeighbours(x, y int) [][2]int {
	result := make([][2]int, 0, 4)
	for _, n := range w.neighbours(x, y) {
		nx, ny := n[0], n[1]
		if w.Grid[ny][nx] == nil {
			result = append(result, [2]int{nx, ny})
		}
	}
	return result
}

// fishNeighbours returns the coordinates of neighbouring cells that are
// currently occupied by fish.
func (w *World) fishNeighbours(x, y int) [][2]int {
	result := make([][2]int, 0, 4)
	for _, n := range w.neighbours(x, y) {
		nx, ny := n[0], n[1]
		if w.Grid[ny][nx] != nil && w.Grid[ny][nx].Kind == FishCell {
			result = append(result, [2]int{nx, ny})
		}
	}
	return result
}

// updateFish applies the Wa-Tor rules for a single fish at (x, y).
// It chooses a random empty neighbour to move into and handles breeding
// by optionally leaving a new fish behind.
func (w *World) updateFish(x, y int, c *Creature, newGrid [][]*Creature) {
	breed := c.BreedCounter + 1

	// Find empty neighbouring cells (based on old grid).
	empties := w.emptyNeighbours(x, y)

	// No free neighbours: fish stays, no reproduction.
	if len(empties) == 0 {
		if newGrid[y][x] == nil {
			newGrid[y][x] = &Creature{
				Kind:         FishCell,
				BreedCounter: breed,
				Energy:       0,
			}
		}
		return
	}

	// Choose a random destination.
	dest := empties[rand.Intn(len(empties))]
	dx, dy := dest[0], dest[1]

	// If someone already took that spot in the new grid, the fish fails to move.
	if newGrid[dy][dx] != nil {
		if newGrid[y][x] == nil {
			newGrid[y][x] = &Creature{
				Kind:         FishCell,
				BreedCounter: breed,
				Energy:       0,
			}
		}
		return
	}

	// Fish moves. Reproduction only happens when moving.
	if breed >= w.Params.FishBreed {
		// Leave a new fish behind and reset parent's counter.
		if newGrid[y][x] == nil {
			newGrid[y][x] = &Creature{
				Kind:         FishCell,
				BreedCounter: 0,
				Energy:       0,
			}
		}
		newGrid[dy][dx] = &Creature{
			Kind:         FishCell,
			BreedCounter: 0,
			Energy:       0,
		}
	} else {
		// Just move, no reproduction.
		newGrid[dy][dx] = &Creature{
			Kind:         FishCell,
			BreedCounter: breed,
			Energy:       0,
		}
	}
}

// updateShark applies the Wa-Tor rules for a single shark at (x, y).
// The shark first loses energy, then preferentially moves to an adjacent
// fish cell (eating the fish and gaining energy) or otherwise to an empty
// cell. It may reproduce by leaving a new shark behind.
func (w *World) updateShark(x, y int, c *Creature, newGrid [][]*Creature) {
	// Starvation: shark loses 1 energy every chronon.
	energy := c.Energy - 1
	if energy <= 0 {
		// Shark dies.
		return
	}

	breed := c.BreedCounter + 1

	fishN := w.fishNeighbours(x, y)
	destX, destY := x, y
	ate := false

	if len(fishN) > 0 {
		// Prefer eating a fish.
		dest := fishN[rand.Intn(len(fishN))]
		destX, destY = dest[0], dest[1]
		ate = true
	} else {
		empties := w.emptyNeighbours(x, y)
		if len(empties) > 0 {
			dest := empties[rand.Intn(len(empties))]
			destX, destY = dest[0], dest[1]
		} else {
			// No movement possible: stay where you are.
			destX, destY = x, y
		}
	}

	// If we ate a fish, reset energy.
	if ate {
		energy = w.Params.Starve
	}

	// If moving into an originally empty cell, but newGrid is already occupied,
	// we fail to move and stay in place.
	if !ate && (destX != x || destY != y) && newGrid[destY][destX] != nil {
		destX, destY = x, y
	}

	// Reproduction: happens only if shark actually moves to a neighbour cell.
	if breed >= w.Params.SharkBreed && (destX != x || destY != y) {
		// Leave a baby behind at the original position (with full energy).
		if newGrid[y][x] == nil {
			newGrid[y][x] = &Creature{
				Kind:         SharkCell,
				BreedCounter: 0,
				Energy:       w.Params.Starve,
			}
		}
		breed = 0 // reset parent's counter
	}

	// Place / move the parent shark at its destination.
	newGrid[destY][destX] = &Creature{
		Kind:         SharkCell,
		BreedCounter: breed,
		Energy:       energy,
	}
}
