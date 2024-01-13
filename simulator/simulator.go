package simulator

import (
    "zawie/life/simulator/vec2"
    "golang.org/x/image/colornames"
    "image/color"
    "math/rand"
)

const minInteractionDistance = 10

const maxVelocity = 1

type Vec2 struct {
    X, Y int
}

type Particle struct {
    Position vec2.Vector
	Velocity vec2.Vector
    Color color.Color
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

    particleTypes := []color.Color{colornames.Hotpink, colornames.Limegreen, colornames.Yellow, colornames.Blue, colornames.Red}
    for _,t := range particleTypes {
        for i := 0; i < particleCount; i++ {
            var p Particle
            p.Color = t
            p.Position.X = rand.Float64() * sim.bounds.X
            p.Position.Y = rand.Float64() * sim.bounds.Y
            sim.particles = append(sim.particles, p)
        } 
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

        if vec2.Magnitude(sim.particles[i].Velocity) > maxVelocity {
            sim.particles[i].Velocity = vec2.Scale(vec2.Unit(sim.particles[i].Velocity), maxVelocity)
        }
    }

    // Modify position 
    for i := range sim.GetAllParticles() {
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
    var factor float64

    diff := vec2.Subtract(source.Position, influence.Position)
    distance := vec2.Magnitude(diff)
    direction := vec2.Scale(diff, 1/distance)

    repulsionDistance := 10.0
    influenceDistance := 100.0
    if distance < repulsionDistance {
        factor -= 0.25*(distance - repulsionDistance)
    } else if distance < influenceDistance {
        factor += 0.1/(distance - influenceDistance)
    }

    return vec2.Scale(direction, factor)
}