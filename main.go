package main

import (
	"fmt"
	"zawie/life/simulator"
	"zawie/life/simulator/vec2"
	"zawie/life/gui"
)


type modelImpl struct {
	step  func()
	reset func()

	population  []*simulator.Organism
	simulator 	*simulator.Simulator
	gui 		*gui.Gui
}

func (m *modelImpl) Step() { 
	size := m.gui.GetSize()
	m.simulator.UpdateSize(size.X, size.Y)
	m.simulator.Step() 
}
func (m *modelImpl) Reset() { 
	size := m.gui.GetSize()
	m.simulator = simulator.NewSimulator(size.X, size.Y)
	fmt.Println("Generating population...")
	m.population = simulator.CreateRandomPopulation(simulator.MAX_POPULATION_SIZE)
	fmt.Println("Done generating population.")
	m.simulator.AddOrganisms(m.population)
}
func (m modelImpl) GetAllParticles() []*simulator.Particle { 
	return m.simulator.GetAllParticles() 
}
func (m modelImpl) GetNeighborhood(pos vec2.Vector) []*simulator.Particle { return m.simulator.GetNeighborhood(pos) }
func (m modelImpl) ChunkSize() float64 { return m.simulator.ChunkSize }


func main() {
	model := &modelImpl{}
	model.gui = gui.NewGui(model)
	model.Reset()
	model.gui.Run()
}