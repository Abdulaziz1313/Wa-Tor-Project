//--------------------------------
//Author: Abdulaziz Hameed Aloufi
//Student ID: C00266252
//--------------------------------

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Params holds all configuration parameters for a single run of the
// Wa-Tor simulation. These values are set from command–line flags.
type Params struct {
	NumShark   int    // starting population of sharks
	NumFish    int    // starting population of fish
	FishBreed  int    // chronons before a fish can reproduce
	SharkBreed int    // chronons before a shark can reproduce
	Starve     int    // chronons a shark can survive without food
	GridSize   int    // width and height of the toroidal grid (GridSize x GridSize)
	Threads    int    // number of goroutines to use for the parallel step
	Steps      int    // number of simulation steps (chronons) to run
	PrintEvery int    // how often to print the world in text mode (0 = never)
	CSVFile    string // optional path to CSV file for population statistics

	Graphics bool // if true, run the graphical (Ebiten) version instead of text mode
}

// parseParams parses command–line flags into a Params struct and performs
// basic validation of the input values.
func parseParams() Params {
	p := Params{}

	flag.IntVar(&p.NumShark, "numShark", 100, "Starting population of sharks")
	flag.IntVar(&p.NumFish, "numFish", 200, "Starting population of fish")
	flag.IntVar(&p.FishBreed, "fishBreed", 3, "Chronons before a fish can reproduce")
	flag.IntVar(&p.SharkBreed, "sharkBreed", 5, "Chronons before a shark can reproduce")
	flag.IntVar(&p.Starve, "starve", 3, "Chronons a shark can live without food")
	flag.IntVar(&p.GridSize, "gridSize", 20, "Grid dimension (NxN)")
	flag.IntVar(&p.Threads, "threads", 1, "Number of threads (goroutines) to use")
	flag.IntVar(&p.Steps, "steps", 200, "Number of simulation steps (chronons)")
	flag.IntVar(&p.PrintEvery, "printEvery", 20, "How often to print the grid (0 = never)")
	flag.StringVar(&p.CSVFile, "csv", "", "Optional CSV file to write stats (e.g. stats.csv)")
	flag.BoolVar(&p.Graphics, "graphics", false, "Run with graphical window (Ebiten)")

	flag.Parse()

	if p.NumShark < 0 || p.NumFish < 0 || p.GridSize <= 0 {
		fmt.Println("Error: invalid parameter values")
		os.Exit(1)
	}
	if p.Steps <= 0 {
		fmt.Println("Error: steps must be > 0")
		os.Exit(1)
	}

	return p
}

// main is the entry point of the Wa-Tor simulation. It parses parameters,
// prints a short summary and then runs either the text or graphical
// version of the simulation.
func main() {
	params := parseParams()

	fmt.Println("Wa-Tor Simulation")
	fmt.Println("-----------------")
	fmt.Printf("Sharks      : %d\n", params.NumShark)
	fmt.Printf("Fish        : %d\n", params.NumFish)
	fmt.Printf("FishBreed   : %d\n", params.FishBreed)
	fmt.Printf("SharkBreed  : %d\n", params.SharkBreed)
	fmt.Printf("Starve      : %d\n", params.Starve)
	fmt.Printf("GridSize    : %d x %d\n", params.GridSize, params.GridSize)
	fmt.Printf("Threads     : %d\n", params.Threads)
	fmt.Printf("Steps       : %d\n", params.Steps)
	fmt.Printf("PrintEvery  : %d\n", params.PrintEvery)
	if params.CSVFile != "" {
		fmt.Printf("CSV output  : %s\n", params.CSVFile)
	}

	if params.Graphics {
		fmt.Println("Mode        : graphics")
		RunSimulationGraphics(params)
	} else {
		fmt.Println("Mode        : text")
		RunSimulation(params)
	}
}

// RunSimulation executes the Wa-Tor simulation in text mode.
// It optionally writes population statistics to a CSV file.
func RunSimulation(p Params) {
	world := NewWorld(p)

	var csvWriter *bufio.Writer
	var csvFile *os.File
	var err error

	if p.CSVFile != "" {
		csvFile, err = os.Create(p.CSVFile)
		if err != nil {
			fmt.Println("Error creating CSV file:", err)
			os.Exit(1)
		}
		defer csvFile.Close()

		csvWriter = bufio.NewWriter(csvFile)
		defer csvWriter.Flush()

		// CSV header: step, fish count, shark count
		fmt.Fprintln(csvWriter, "step,fish,sharks")
	}

	start := time.Now()

	for step := 0; step < p.Steps; step++ {
		fish, sharks := world.Count()

		// Log stats to CSV if requested.
		if csvWriter != nil {
			fmt.Fprintf(csvWriter, "%d,%d,%d\n", step, fish, sharks)
		}

		// Optionally print the world in ASCII.
		if p.PrintEvery > 0 && step%p.PrintEvery == 0 {
			clearScreen()
			fmt.Printf("Step %d\n", step)
			fmt.Printf("Fish=%d  Sharks=%d\n", fish, sharks)
			world.PrintColored()
			time.Sleep(50 * time.Millisecond) // small delay so animation is visible
		}

		// Sequential vs parallel step.
		if p.Threads > 1 {
			world.StepParallel(p.Threads)
		} else {
			world.Step()
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\nSimulation finished in %v\n", elapsed)
}

// clearScreen clears the terminal using the appropriate mechanism for the
// current operating system. On Windows it calls "cls"; on Unix-like systems
// it writes ANSI escape codes.
func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	} else {
		// ANSI escape for Unix-like systems.
		fmt.Print("\033[2J\033[H")
	}
}
