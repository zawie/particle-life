package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"zawie/life/simulator"
	"time"
	"fmt"
)

const updatesPerSecond = 60

func main() {

	const X = 2000
	const Y = 1000

	fmt.Println("Creating simulator...")
	sim := simulator.NewSimulator(X, Y, 1000)

	fmt.Println("Opening window...")
	pixelgl.Run(func() {
		cfg := pixelgl.WindowConfig{
			Title:  "Zawie's Particle Life",
			Bounds: pixel.R(0, 0, X, Y),
			VSync:  true,
			Maximized: true,
			Resizable: true,
		}
		win, err := pixelgl.NewWindow(cfg)
		if err != nil {
			panic(err)
		}

		fmt.Println("Starting main loop...")
		for !win.Closed() {
			start := time.Now()
			sim.Step()

			imd := imdraw.New(nil)
			for _, particle := range sim.GetAllParticles() {
				imd.Color = particle.Color
				imd.Push(pixel.V(particle.Position.X, particle.Position.Y))
				imd.Circle(1, 0)
			}

			size := win.Bounds().Size()
			sim.UpdateSize(size.X, size.Y)

			// Sleep to ensure we are updating grapgics at a consistent rate
			time.Sleep(start.Add(time.Second * 1/updatesPerSecond).Sub(time.Now()))

			win.Clear(colornames.Black)
			imd.Draw(win)
			win.Update()
		}

		fmt.Println("Window closed. Terminating gracefully.")
	})

}