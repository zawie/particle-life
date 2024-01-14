package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"github.com/faiface/pixel/text"
	"zawie/life/simulator"
	"time"
	"fmt"
)

const framesPerSecond = 60

func main() {

	const X = 2000
	const Y = 1000

	fmt.Println("Creating simulator...")
	sim := simulator.NewSimulator(X, Y, 500)

	fmt.Println("Opening window...")
	pixelgl.Run(func() {


		speed := 1
		oldSpeed := speed
		debugMode := false 

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
		atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

		fmt.Println("Starting main loop...")
		for !win.Closed() {
			start := time.Now()
			size := win.Bounds().Size()
			sim.UpdateSize(size.X, size.Y)
			for i := 0; i < 1 << (speed-1); i++ { 
				sim.Step()
			}

			imd := imdraw.New(nil)
			particles := sim.GetAllParticles()
			for _, particle := range particles {
				imd.Color = particle.Color
				imd.Push(pixel.V(particle.Position.X, particle.Position.Y))
				imd.Circle(1, 0)

				if debugMode && particle.Id == 0 {
					for _, neighbor := range sim.GetNearParticles(particle.Position) {
						imd.Color = colornames.Limegreen
						imd.Push(pixel.V(particle.Position.X, particle.Position.Y))
						imd.Push(pixel.V(neighbor.Position.X, neighbor.Position.Y))
						imd.Line(1)
					}

					imd.Color = colornames.White
					imd.Push(pixel.V(particle.Position.X, particle.Position.Y))
					imd.Circle(5, 1)
				}
			}

			// Text controls
			if win.JustPressed(pixelgl.KeyG) {
				debugMode = !debugMode
			}

			// Speed controls
			if win.JustPressed(pixelgl.KeyL) {
				if speed < 10 {
					speed++
				}
			}

			if win.JustPressed(pixelgl.KeyJ) {
				if speed > 0 {
					speed--
					if speed == 0 {
						oldSpeed = 1
					}
				}
			}

			if win.JustPressed(pixelgl.KeyK) {
				if speed > 0 {
					oldSpeed = speed
					speed = 0
				} else {
					 speed = oldSpeed
				}
			}
			
			grid := imdraw.New(nil)
			if debugMode {
				grid.Color = colornames.Gray
				for x := 0.0; x <= size.X; x += 100.0 {
					for y := 0.0; y <= size.Y; y += 100.0 {
						grid.Push(pixel.V(x,y))
						grid.Push(pixel.V(x+100,y+100))
						grid.Rectangle(1)
					}
				}
			}

			// Sleep to ensure we are updating grapgics at a consistent rate
			time.Sleep(start.Add(time.Second * 1/framesPerSecond).Sub(time.Now()))

			win.Clear(colornames.Black)
			imd.Draw(win)
			if debugMode {
				grid.Draw(win)
			}

			topLeftTxt := text.New(pixel.V(30, size.Y-30), atlas)
			topLeftTxt.Color = colornames.White
			if speed == 0 {
				topLeftTxt.WriteString("PAUSED")
			} else {
				topLeftTxt.WriteString(fmt.Sprintf("SPEED x%d", 1 << (speed-1)))
			}

			topLeftTxt.Draw(win, pixel.IM)

			win.Update()
		}

		fmt.Println("Window closed. Terminating gracefully.")
	})

}