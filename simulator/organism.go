package simulator

import (
	"zawie/life/simulator/vec2"
	"math/rand"
)

const MAX_POPULATION_SIZE = 8
const MAX_ORGANISM_PARTICLE_COUNT = 150

// Create a struct named Message that implements the Printer interface
type Organism struct {
	particles []*Particle
 	internalMatrix [PARTICLE_TYPE_COUNT][PARTICLE_TYPE_COUNT]float64
 	externalMatrix [PARTICLE_TYPE_COUNT][PARTICLE_TYPE_COUNT]float64
}

func CreateRandomPopulation(n int) (organisms []*Organism) {
	if n > MAX_POPULATION_SIZE {
		panic("Requesting too many organisms!!!")
	}

	for i := 0; i < n; i++ {
		organisms = append(organisms, RandomOrganism())
	}

	return
}

func RandomOrganism() *Organism {
	var org Organism

	for i := 0; i < MAX_ORGANISM_PARTICLE_COUNT; i++ {
		var p Particle
		p.TypeId = rand.Int() % PARTICLE_TYPE_COUNT
		p.Position.X = rand.Float64() * 100.0 - 50.0
		p.Position.Y = rand.Float64() * 100.0 - 50.0
		p.Mass = 1
		org.particles = append(org.particles, &p)
	}

	for i := 0; i < PARTICLE_TYPE_COUNT; i++ {
		for j := 0; j < PARTICLE_TYPE_COUNT; j++ {
			v := (2.0*rand.Float64() - 1.0)
			org.externalMatrix[i][j] = v - 1.0
			if i == j {
				org.internalMatrix[i][j] = .1
			} else {
				org.internalMatrix[i][j] = v + 1.0
			}
		}
	}

	return &org
}

func (org Organism) GetAllParticles() []*Particle {
	return org.particles
}

func (org Organism) GetInternalInfluenceFactor(a int, b int) float64 {
    return org.internalMatrix[a][b]
}

func (org Organism) GetExternalInfluenceFactor(a int, b int) float64 {
    return org.externalMatrix[a][b]
}

func (org Organism) Position(position vec2.Vector) {
	avg := vec2.Vector{}
	for _, particle := range org.particles {
		avg = vec2.Add(avg, particle.Position)
    }

	avg = vec2.Scale(avg, float64(1/len(org.particles)))

	for _, particle := range org.particles {
		particle.Position = vec2.Add(vec2.Subtract(particle.Position, avg), position)
    }
}