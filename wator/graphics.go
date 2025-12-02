//--------------------------------
//Author: Abdulaziz Hameed Aloufi
//Student ID: C00266252
//--------------------------------

package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// Game wraps the Ebiten game state for the graphical Wa-Tor simulation.
// It holds the current world, parameters, step counter and pre-created
// images used to draw fish and sharks.
type Game struct {
	world    *World
	params   Params
	step     int // current simulation step
	frame    int // frame counter used to slow down the simulation
	fishImg  *ebiten.Image
	sharkImg *ebiten.Image
}

// pixelSize controls how many pixels wide and high each simulation cell is.
const pixelSize = 4

// windowScale controls how much larger the window is than the logical
// resolution. Ebiten will scale the logical canvas up to this window size.
const windowScale = 2

// RunSimulationGraphics starts the Wa-Tor simulation using Ebiten for
// graphical output. It opens a window and runs until the configured
// number of steps has been reached or the user closes the window.
func RunSimulationGraphics(p Params) {
	world := NewWorld(p)

	// Pre-create small images for fish and sharks to improve performance.
	fishImg := ebiten.NewImage(pixelSize, pixelSize)
	fishImg.Fill(color.RGBA{0, 200, 255, 255}) // cyan-ish fish

	sharkImg := ebiten.NewImage(pixelSize, pixelSize)
	sharkImg.Fill(color.RGBA{255, 100, 50, 255}) // orange-ish shark

	g := &Game{
		world:    world,
		params:   p,
		step:     0,
		frame:    0,
		fishImg:  fishImg,
		sharkImg: sharkImg,
	}

	logicalW := p.GridSize * pixelSize
	logicalH := p.GridSize * pixelSize

	ebiten.SetWindowSize(logicalW*windowScale, logicalH*windowScale)
	ebiten.SetWindowTitle("Wa-Tor Simulation")

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Layout reports the logical resolution of the game. Ebiten will scale
// this resolution up to the actual window size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.params.GridSize * pixelSize, g.params.GridSize * pixelSize
}

// Update advances the simulation. It is called every frame by Ebiten.
// To slow the simulation down, we only perform a simulation step every
// few frames. When the configured number of steps is reached, the game
// terminates.
func (g *Game) Update() error {

	if g.step >= g.params.Steps {
		return ebiten.Termination
	}

	g.frame++

	const framesPerStep = 4
	if g.frame%framesPerStep != 0 {
		return nil
	}

	if g.params.Threads > 1 {
		g.world.StepParallel(g.params.Threads)
	} else {
		g.world.Step()
	}

	g.step++
	return nil
}

// Draw renders the current world state to the Ebiten screen. Fish and
// sharks are drawn in different colours on a dark “water” background,
// and a simple HUD shows the step counter and population sizes.
func (g *Game) Draw(screen *ebiten.Image) {

	water := color.RGBA{0, 10, 40, 255}
	screen.Fill(water)

	for y := 0; y < g.params.GridSize; y++ {
		for x := 0; x < g.params.GridSize; x++ {
			cell := g.world.CellAt(x, y)

			var img *ebiten.Image
			switch cell {
			case FishCell:
				img = g.fishImg
			case SharkCell:
				img = g.sharkImg
			default:
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*pixelSize), float64(y*pixelSize))
			screen.DrawImage(img, op)
		}
	}

	fishCount, sharkCount := g.world.Count()
	hud := fmt.Sprintf("Step: %d / %d   Fish: %d   Sharks: %d",
		g.step, g.params.Steps, fishCount, sharkCount)

	text.Draw(screen, hud, basicfont.Face7x13, 8, 16, color.White)
}
