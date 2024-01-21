package main

import (
	"zawie/life/simulator"
	"zawie/life/simulator/vec2"
	"zawie/life/gui"
)


type modelImpl struct {
	step  func()
	reset func()

	simulator 	*simulator.Simulator
	gui 		*gui.Gui
	particleCount	int
}

func (m *modelImpl) Step() { 
	size := m.gui.GetSize()
	m.simulator.UpdateSize(size.X, size.Y)
	m.simulator.Step() 
}
func (m *modelImpl) Reset() { 
	size := m.gui.GetSize()
	m.simulator = simulator.NewSimulator(size.X, size.Y, m.particleCount)
}
func (m modelImpl) GetAllParticles() []*simulator.Particle { 
	return m.simulator.GetAllParticles() 
}
func (m modelImpl) GetNeighborhood(pos vec2.Vector) []*simulator.Particle { return m.simulator.GetNeighborhood(pos) }
func (m modelImpl) ChunkSize() float64 { return m.simulator.ChunkSize }


func main() {
	model := &modelImpl{
		particleCount: 1000,
	}
	model.gui = gui.NewGui(model)
	model.Reset()
	model.gui.Run()
}