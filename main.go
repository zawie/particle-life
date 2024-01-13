package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"zawie/life/simulator"
)

func main() {

	const X = 2000
	const Y = 1000

	sim := simulator.NewSimulator(X, Y, 100)

	pixelgl.Run(func() {
		cfg := pixelgl.WindowConfig{
			Title:  "Pixel Rocks!",
			Bounds: pixel.R(0, 0, X, Y),
			VSync:  true,
		}
		win, err := pixelgl.NewWindow(cfg)
		if err != nil {
			panic(err)
		}
		
		for !win.Closed() {
			sim.Step()

			imd := imdraw.New(nil)
			for _, particle := range sim.GetAllParticles() {
				imd.Color = colornames.Limegreen
				imd.Push(pixel.V(particle.Position.X, particle.Position.Y))
				imd.Circle(1, 1)
			}

			win.Clear(colornames.Black)
			imd.Draw(win)
			win.Update()
		}
	})

}