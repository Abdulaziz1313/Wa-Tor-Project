# Wa-Tor Simulation – Performance Results

Author: Abdulaziz Aloufi (C00266252)  

---

## 1. Goal

This document presents performance measurements for my Wa-Tor simulation, focusing on:

- **Parallel speedup** when using different numbers of threads (`-threads` parameter).
- How well the simulation scales with multiple goroutines.
- A short discussion of why the observed speedup is limited.

The simulation is based on the rules described in Dewdney’s *“Sharks and Fish wage an ecological war on the toroidal planet of Wa-Tor”*.

---

## 2. Experimental Setup

All performance measurements were taken under the following conditions:

- **Platform:** Linux (Ubuntu under WSL)
- **Implementation:** Go (with goroutines for parallel steps)
- **Build/Run:** `go run .` from the project root
- **Mode:** Text mode (no graphics window) – `-graphics` left at default (`false`)
- **Output:** No grid printing during timing runs (`-printEvery=0`)

### Fixed Simulation Parameters

For all tests in this document, the following parameters were used:

- `GridSize = 200` → world is **200 × 200** cells (40,000 cells)
- `NumFish = 8000`
- `NumShark = 2000`
- `FishBreed = 3`
- `SharkBreed = 5`
- `Starve = 3`
- `Steps = 2000`
- `PrintEvery = 0` (no text printing during runs)
- `CSVFile` not used in these runs

The only parameter that changed between runs was:

- `Threads ∈ {1, 2, 4, 8}`

---

## 3. Commands Used

Each run was executed from the project directory using:

```bash
go run . \
  -gridSize=200 \
  -numFish=8000 \
  -numShark=2000 \
  -steps=2000 \
  -printEvery=0 \
  -threads=N
where N is one of 1, 2, 4, 8.

Example (1 thread):

bash
Copy code
go run . -gridSize=200 -numFish=8000 -numShark=2000 -steps=2000 -printEvery=0 -threads=1
The program prints the total simulation time at the end in the form:

text
Copy code
Simulation finished in 8.600170383s

| Threads | Time (seconds) |
| ------: | -------------: |
|       1 |  8.600170383 s |
|       2 |  8.267587827 s |
|       4 |  9.483419002 s |
|       8 |  9.573353838 s |
```



## 4. Graphs
4.1 Runtime vs Threads

This graph shows that runtime does not decrease as we increase the number of threads.
After 2 threads, runtime actually increases.

4.2 Speedup vs Threads

The speedup graph confirms:

Slight improvement at 2 threads.

Sub-linear and then negative speedup at 4 and 8 threads.

## 5. Discussion of Parallel Design

Parallelism is implemented in World.StepParallel(threads int):

The grid is split into row ranges.

Each goroutine:

Processes fish in its chunk.

Then processes sharks in its chunk.

Writes to a local new grid ([][]*Creature).

After all goroutines finish, their local grids are merged into a single final grid for the next step.

This design has important performance consequences:

- **5.1 Per-thread local grids**

Each thread allocates and writes to its own [][]*Creature.

At the end of each step, all these grids must be merged.

Allocation + merge overhead is quite large.

- **5.2 Merge cost dominates**

The actual computation per cell (checking neighbours, moving creatures) is relatively simple.

For this problem size (40,000 cells × 2000 steps), the cost of merging thread copies of the grid becomes significant.

As thread count increases, the merge work also increases, but the useful simulation work per goroutine does not grow.

- **5.3 Overhead of goroutines and scheduling**

Each step starts multiple goroutines and uses a sync.WaitGroup to synchronise them.

With more threads, the scheduling/coordination overhead grows, but the total amount of useful computation is fixed.

- **5.4 Amdahl’s Law**

Some parts of the code are effectively sequential, especially:

Merging of partial grids.

Some bookkeeping and allocation.

According to Amdahl’s Law, if a significant portion of the program is sequential or synchronised, the maximum possible speedup is limited.
In this implementation, the overhead is large enough that beyond 2 threads, parallelisation gives negative benefit.

## 6. Conclusion

The Wa-Tor simulation:

Is functionally correct and supports:

Text mode and graphical mode.

Configurable numbers of fish, sharks, and chronons.

A -threads parameter for parallel updates.

Performance experiments on a 200 × 200 grid with 2000 steps show:

Slight speedup at 2 threads.

No speedup, and even slowdown, at 4 and 8 threads due to merge and scheduling overhead.

These results illustrate a key concurrency lesson:

Adding more threads does not automatically give better performance;
algorithm design, memory allocation, and synchronisation costs are critical.