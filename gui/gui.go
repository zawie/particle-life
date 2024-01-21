package gui

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
    "zawie/life/simulator/vec2"
    "golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"zawie/life/simulator"
    "image/color"
	"time"
)

type Model interface {
	Step()
	Reset()
	GetAllParticles() []*simulator.Particle
	GetNeighborhood(pos vec2.Vector) []*simulator.Particle
	ChunkSize() float64
}

type Mode uint8
const (
	PLAIN_MODE Mode = iota
	DEBUG_MODE
	TEMP_MODE
)
const DEFAULT_MODE = PLAIN_MODE


type Gui struct {
	cfg  pixelgl.WindowConfig;
	window *pixelgl.Window;
	atlas *text.Atlas;
	maxFramesPerSecond float64;
	mode Mode;
	model Model;

	speed uint8;
	oldSpeed uint8;
}

var particleTypes = []color.Color{colornames.Hotpink, colornames.Limegreen, colornames.Yellow, colornames.Blue, colornames.Red}

var influenceMatrix [5][5]float64

func NewGui(model Model) *Gui {
	cfg := pixelgl.WindowConfig{
		Title:  "Zawie's Particle Life",
		Bounds: pixel.R(0, 0, 2000, 1000),
		VSync:  true,
		Maximized: true,
		Resizable: true,
	}

    return &Gui{
		cfg: cfg,
		speed: 1,
		model: model,
		atlas: text.NewAtlas(basicfont.Face7x13, text.ASCII),
		maxFramesPerSecond: 60,
	}
}

func (gui *Gui) GetSize() pixel.Vec {
	if gui.window != nil {
		return gui.window.Bounds().Size()
	}
	return gui.cfg.Bounds.Size()
}

func (gui *Gui) Run() {
	pixelgl.Run(func() {

		win, err := pixelgl.NewWindow(gui.cfg)
		if err != nil {
			panic(err)
		}
		gui.window = win

		for !gui.window.Closed() {
			start := time.Now()
			if gui.speed > 0 {
				for i := 0; i < 1 << (gui.speed-1); i++ { 
					gui.model.Step()
				}
			}	
			gui.processInput()
			// Sleep to ensure or FPS is no more than max
			time.Sleep(start.Add(time.Second / time.Duration(gui.maxFramesPerSecond)).Sub(time.Now()))
			gui.window.Clear(colornames.Black)
			gui.render()
			gui.window.Update()
		}
	})

	fmt.Println("Window closed. Terminating gracefully.")
}

func (gui *Gui) render() {
	if gui.mode == DEBUG_MODE {
		gui.renderGrid()
	}
	gui.renderParticles()
	gui.renderText()
}

func (gui *Gui) processInput() {
		// GUI mode
		if gui.window.JustPressed(pixelgl.KeyG) {
			gui.toggleMode(DEBUG_MODE)
		}
		if gui.window.JustPressed(pixelgl.KeyT) {
			gui.toggleMode(TEMP_MODE)
		}

		// Model reset
		if gui.window.JustPressed(pixelgl.KeyR) {
			gui.model.Reset()
		}

		// Speed
		if gui.window.JustPressed(pixelgl.KeyJ) {
			if gui.speed > 0 {
				gui.speed--
				if gui.speed == 0 {
					gui.oldSpeed = 1
				}
			}
		}
		if gui.window.JustPressed(pixelgl.KeyK) {
			if gui.speed > 0 {
				gui.oldSpeed = gui.speed
				gui.speed = 0
			} else {
				 gui.speed = gui.oldSpeed
			}
		}
		if gui.window.JustPressed(pixelgl.KeyL) {
			if gui.speed < 10 {
				gui.speed++
			}
		}
}

func (gui *Gui) toggleMode(mode Mode) {
	if gui.mode == mode {
		gui.mode = DEFAULT_MODE
	} else {
		gui.mode = mode
	}
}

func (gui *Gui) renderGrid() {
	grid := imdraw.New(nil)
	chunkSize := gui.model.ChunkSize()
	grid.Color = color.RGBA{R: 15, G:15, B:15, A:0}
	for x := 0.0; x <= gui.GetSize().X; x += chunkSize {
		for y := 0.0; y <= gui.GetSize().Y; y += chunkSize  {
			grid.Push(pixel.V(x,y))
			grid.Push(pixel.V(x+chunkSize,y+chunkSize))
			grid.Rectangle(1)
		}
	}
	grid.Draw(gui.window)
}

func (gui *Gui) renderParticles() {
	dots := imdraw.New(nil)
	lines := imdraw.New(nil)
	particles := gui.model.GetAllParticles()
	for _, particle := range particles {
		if gui.mode == TEMP_MODE {
			dots.Color = getRatioColor(vec2.Magnitude(particle.Velocity)/0.9)
		} else {
			dots.Color = particle.Color
		}
		dots.Push(pixel.V(particle.Position.X, particle.Position.Y))
		dots.Circle(1, 0)

		if gui.mode == DEBUG_MODE {
			for _, neighbor := range gui.model.GetNeighborhood(particle.Position) {
				lines.Push(pixel.V(particle.Position.X, particle.Position.Y))
				lines.Push(pixel.V(neighbor.Position.X, neighbor.Position.Y))
				if neighbor.Mass > 1 {
					lines.Color = colornames.White
				} else {
					lines.Color = colornames.Gray
				}
				lines.Line(1)
			}
		}
	}
	lines.Draw(gui.window)
	dots.Draw(gui.window)
}

func (gui *Gui) renderText() {
	topLeftTxt := text.New(pixel.V(30, gui.GetSize().Y-30), gui.atlas)
	topLeftTxt.Color = colornames.White
	if gui.speed == 0 {
		topLeftTxt.WriteString("PAUSED")
	} else if gui.speed > 1 {
		topLeftTxt.WriteString(fmt.Sprintf("SPEED x%d", 1 << (gui.speed-1)))
	}
	topLeftTxt.Draw(gui.window, pixel.IM)
}

func getRatioColor(ratio float64) color.RGBA {
	// Ensure ratio is within the valid range
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1.0 {
		ratio = 1.0
	}
	
	return color.RGBA{uint8(ratio * 255), uint8(ratio * 255),uint8(ratio * 255), 255}
}
