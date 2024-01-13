package simulator

import (
    "zawie/life/simulator/vec2"
    "golang.org/x/image/colornames"
    "image/color"
    "math/rand"
    "sync"
)

const maxVelocity = 2
const repulsionDistance = 10.0
const influenceDistance = 100.0

type Vec2 struct {
    X, Y int
}

type Particle struct {
    Position vec2.Vector
    id  int
	Velocity vec2.Vector
    Color color.Color
}

type Simulator struct {
    particles []Particle
    tick uint64
    bounds vec2.Vector
    regionToParticleIndex [][]int
}

var particleTypes = []color.Color{colornames.Hotpink, colornames.Limegreen, colornames.Yellow, colornames.Blue, colornames.Red}

var influenceMatrix [5][5]float64

func NewSimulator(X float64, Y float64, particleCount int) *Simulator {

    var sim Simulator
    sim.bounds = vec2.Vector {
        X: X,
        Y: Y,
    }
    for id,color := range particleTypes {
        for i := 0; i < particleCount; i++ {
            var p Particle
            p.Color = color
            p.id = id
            p.Position.X = rand.Float64() * sim.bounds.X
            p.Position.Y = rand.Float64() * sim.bounds.Y
            sim.particles = append(sim.particles, p)
        } 
    }

    for i, _ := range particleTypes {
        for j, _ := range particleTypes {
            if i == j {
                influenceMatrix[i][j] = .1
            } else {
                v := (2.0*rand.Float64() - 1.0)
                influenceMatrix[i][j] = v*v*v
            }
        }
    }

    return &sim
}

func (sim *Simulator) UpdateSize(X float64, Y float64) {
    sim.bounds = vec2.Vector {
        X: X,
        Y: Y,
    }
}

func (sim *Simulator) Step() {

    // Compute velocity
    var wg0 sync.WaitGroup
    var wg1 sync.WaitGroup
    wg0.Add(len(sim.particles))
    wg1.Add(len(sim.particles))

    for idx := range sim.particles {
        go func(i int) {
            defer wg1.Done()

            var force vec2.Vector
            for _, neighbor := range sim.getNearParticles(sim.particles[i]) {
                if neighbor == sim.particles[i] {
                    continue
                }
                force = vec2.Add(force, sim.computeForce(sim.particles[i], neighbor))
            }
            sim.particles[i].Velocity.X += force.X
            sim.particles[i].Velocity.Y += force.Y

            if vec2.Magnitude(sim.particles[i].Velocity) > maxVelocity {
                sim.particles[i].Velocity = vec2.Scale(vec2.Unit(sim.particles[i].Velocity), maxVelocity)
            }      
            
            wg0.Done()
            wg0.Wait() 

            // Modify position 
            sim.particles[i].Position.X += sim.particles[i].Velocity.X
            sim.particles[i].Position.Y += sim.particles[i].Velocity.Y

            // Wrap
            for sim.particles[i].Position.X  > sim.bounds.X {
                sim.particles[i].Position.X -= sim.bounds.X
            }
            for sim.particles[i].Position.X < 0 {
                sim.particles[i].Position.X += sim.bounds.X
            }

            for sim.particles[i].Position.Y  > sim.bounds.Y {
                sim.particles[i].Position.Y -= sim.bounds.Y
            }
            for sim.particles[i].Position.Y < 0 {
                sim.particles[i].Position.Y += sim.bounds.Y
            }
        }(idx)
    }

    wg1.Wait()

    sim.tick++
}

func (sim *Simulator) GetAllParticles() []Particle {
    return sim.particles
}

func (sim *Simulator) getNearParticles(particle Particle) []Particle {
    return sim.particles
}

func (sim *Simulator) computeForce(source Particle, influence Particle) vec2.Vector {
    var factor float64

    diff := vec2.Subtract(source.Position, influence.Position)
    distance := vec2.Magnitude(diff)
    direction := vec2.Scale(diff, 1/distance)

    if distance < repulsionDistance {
        factor -= 0.25*(distance - repulsionDistance)
    } else if distance < influenceDistance {
        factor += getInfluenceFactor(source.id, influence.id)/(distance - influenceDistance)
    }

    return vec2.Scale(direction, factor)
}

func getInfluenceFactor(a int, b int) float64 {
    return influenceMatrix[a][b]
}