package simulator

import (
    "zawie/life/simulator/vec2"
    "math/rand"
)

const minInteractionDistance = 10

type ParticleType int

const (
    Red ParticleType = iota
    Blue
    Green
)

type Vec2 struct {
    X, Y int
}

type Particle struct {
    Position vec2.Vector
	Velocity vec2.Vector
    Type ParticleType
}

type Simulator struct {
    particles []Particle
    tick uint64
    bounds vec2.Vector
}

func NewSimulator(X float64, Y float64, particleCount int) *Simulator {

    var sim Simulator
    sim.bounds = vec2.Vector {
        X: X,
        Y: Y,
    }

    for i := 0; i < particleCount; i++ {
        var p Particle
        p.Position.X = rand.Float64() * sim.bounds.X
        p.Position.Y = rand.Float64() * sim.bounds.Y
        sim.particles = append(sim.particles, p)
    }

    return &sim
}

func (sim *Simulator) Step() {

    // Compute velocity
    for i := range sim.particles {
        var force vec2.Vector
        for _, neighbor := range sim.getNearParticles(sim.particles[i]) {
            if neighbor == sim.particles[i] {
                continue
            }
            force = vec2.Add(force, sim.computeForce(sim.particles[i], neighbor))
        }
        sim.particles[i].Velocity.X += force.X
        sim.particles[i].Velocity.Y += force.Y
    }

    // Modify position 
    for i := range sim.GetAllParticles() {
        sim.particles[i].Position.X += sim.particles[i].Velocity.X
        sim.particles[i].Position.Y += sim.particles[i].Velocity.Y

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
    }

    sim.tick++
}

func (sim *Simulator) GetAllParticles() []Particle {
    return sim.particles
}

func (sim *Simulator) getNearParticles(particle Particle) []Particle {
    return sim.particles
}

func (sim *Simulator) computeForce(source Particle, influence Particle) vec2.Vector {
    diff := vec2.Subtract(source.Position, influence.Position)
    distance := vec2.Magnitude(diff)
    if distance == 0 {
        distance = 0.0001
    }
    direction := vec2.Scale(diff, 1/distance)

    return  vec2.Scale(direction, 0.0001)
}