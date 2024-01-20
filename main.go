package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"github.com/faiface/pixel/text"
	"zawie/life/simulator"
	"zawie/life/simulator/vec2"
	"image/color"
	"time"
	"fmt"
)

const framesPerSecond = 60

const (
	PLAIN_MODE int = iota
	DEBUG_MODE
	TEMP_MODE
)

func main() {

	const X = 2000
	const Y = 1000

	fmt.Println("Creating simulator...")

	particleCount :=  1000

	targetId := 0

	var temperatureHistory []float64

	fmt.Println("Opening window...")
	pixelgl.Run(func() {
		mode := PLAIN_MODE

		speed := 1
		oldSpeed := speed

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

		sim := simulator.NewSimulator(X, Y, particleCount)
		fmt.Println("Starting main loop...")
		for !win.Closed() {
			start := time.Now()

			size := win.Bounds().Size()
			if win.JustPressed(pixelgl.KeyR) {
				sim = simulator.NewSimulator(size.X, size.Y, particleCount)
				temperatureHistory = make([]float64, 0)
			} else {
				sim.UpdateSize(size.X, size.Y)
			}

			if speed > 0 {
				for i := 0; i < 1 << (speed-1); i++ { 
					sim.Step()
					temperatureHistory = append(temperatureHistory, sim.ComputeAverageKineticEnergy())
				}
			}	
			
			imd := imdraw.New(nil)
			particles := sim.GetAllParticles()
			for _, particle := range particles {
				if mode == TEMP_MODE {
					imd.Color = getRatioColor(vec2.Magnitude(particle.Velocity)/0.9)
				} else {
					imd.Color = particle.Color
				}
				imd.Push(pixel.V(particle.Position.X, particle.Position.Y))
				imd.Circle(1, 0)

				if mode == DEBUG_MODE && particle.Id == targetId {
					for _, neighbor := range sim.GetNeighborhood(particle.Position) {
						imd.Color = colornames.Limegreen
						imd.Push(pixel.V(particle.Position.X, particle.Position.Y))
						imd.Push(pixel.V(neighbor.Position.X, neighbor.Position.Y))
						imd.Line(1)
					}

					imd.Color = colornames.Red
					imd.Push(pixel.V(particle.Position.X, particle.Position.Y))
					imd.Circle(sim.RepulsionRadius, 1)

					imd.Color = colornames.Blue
					imd.Push(pixel.V(particle.Position.X, particle.Position.Y))
					imd.Circle(sim.ApproximationRadius, 1)

					imd.Color = colornames.White
					imd.Push(pixel.V(particle.Position.X, particle.Position.Y))
					imd.Circle(sim.InfluenceRadius, 1)
				}
			}

			// Text controls
			if win.JustPressed(pixelgl.KeyG) {
				if mode == DEBUG_MODE {
					mode = PLAIN_MODE
				} else {
					mode = DEBUG_MODE
				}
				targetId++
				targetId %= 100 //TODO: Make dynamic
			}

			if win.JustPressed(pixelgl.KeyT) {
				if mode == TEMP_MODE {
					mode = PLAIN_MODE
				} else {
					mode = TEMP_MODE
				}
				targetId++
				targetId %= 100 //TODO: Make dynamic
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
			if mode == DEBUG_MODE {
				grid.Color = color.RGBA{R: 15, G:15, B:15, A:0}
				for x := 0.0; x <= size.X; x += float64(sim.ChunkSize) {
					for y := 0.0; y <= size.Y; y += float64(sim.ChunkSize)  {
						grid.Push(pixel.V(x,y))
						grid.Push(pixel.V(x+sim.ChunkSize,y+sim.ChunkSize))
						grid.Rectangle(1)
					}
				}
			}

			// Sleep to ensure we are updating grapgics at a consistent rate
			time.Sleep(start.Add(time.Second * 1/framesPerSecond).Sub(time.Now()))

			win.Clear(colornames.Black)
			imd.Draw(win)
			if mode == DEBUG_MODE {
				grid.Draw(win)
			}

			topLeftTxt := text.New(pixel.V(30, size.Y-30), atlas)
			topLeftTxt.Color = colornames.White
			if speed == 0 {
				topLeftTxt.WriteString("PAUSED")
			} else if speed > 1 {
				topLeftTxt.WriteString(fmt.Sprintf("SPEED x%d", 1 << (speed-1)))
			}
			topLeftTxt.Draw(win, pixel.IM)

			temperatureGraph := imdraw.New(nil)
			// Display temperature graph
			max := 0.01
			for _, t := range temperatureHistory {
				if t > max {
					max = t
				}
			}

			graphHeight := 100.0
			count := 0
			agg := 0.0
			batchSize := 50
			x := len(temperatureHistory) - int(size.X) - batchSize
			if x < 0 {
				x = 0
			}
			for ; x < len(temperatureHistory); x++{
				t := temperatureHistory[x]

				if count <= batchSize {
					count++
					agg += t
				} else {
					agg += t - temperatureHistory[x-batchSize]
				}

				y := (agg/float64(count))/max * graphHeight
				temperatureGraph.Color = getRatioColor((agg/float64(count))/max )
				temperatureGraph.Push(pixel.V(float64(x-(len(temperatureHistory)-int(size.X))), float64(y)))
				temperatureGraph.Line(2)
				temperatureGraph.Push(pixel.V(float64(x-(len(temperatureHistory)-int(size.X))), float64(y)))
			}
			if mode == TEMP_MODE {
				temperatureGraph.Draw(win)
			}
			win.Update()
		}

		fmt.Println("Window closed. Terminating gracefully.")
	})

}

func getRatioColor(ratio float64) color.RGBA {
	// Ensure ratio is within the valid range
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1.0 {
		ratio = 1.0
	}
	
	// Return the color.RGBA
	return color.RGBA{uint8(ratio * 255), uint8(ratio * 255),uint8(ratio * 255), 255}
}