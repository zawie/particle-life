package simulator

import (
    "zawie/life/simulator/vec2"
    "golang.org/x/image/colornames"
    "image/color"
    "math/rand"
    "sync"
    "fmt"
)

const maxVelocity = 2
const repulsionDistance = 10.0
const influenceDistance = 100.0
const chunkSize = 100

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
    tick uint64
    bounds vec2.Vector
    chunks [][]map[*Particle]struct{}
}

var particleTypes = []color.Color{colornames.Hotpink, colornames.Limegreen, colornames.Yellow, colornames.Blue, colornames.Red}

var influenceMatrix [5][5]float64

func NewSimulator(X float64, Y float64, particleCount int) *Simulator {

    var sim Simulator
    (&sim).UpdateSize(X,Y)

    particlesAdded := 0 
    for particlesAdded< particleCount {
        var p Particle
        p.id = rand.Int() % len(particleTypes)
        p.Color = particleTypes[p.id]
        p.Position.X = rand.Float64() * sim.bounds.X
        p.Position.Y = rand.Float64() * sim.bounds.Y
        p.Velocity.X = rand.Float64() - 0.5
        p.Velocity.Y = rand.Float64() - 0.5
        sim.AddParticle(&p)
        particlesAdded++
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
    grew := X > sim.bounds.X || Y > sim.bounds.Y 
    shrunk := X < sim.bounds.X || Y < sim.bounds.Y 

    sim.bounds = vec2.Vector {
        X: X,
        Y: Y,
    }

    if grew {
        fmt.Println("Grew!")
        I := int((X + (chunkSize-1))/chunkSize)
        J := int((Y + (chunkSize-1))/chunkSize)
        for i := len(sim.chunks); i < I; i++ {
            sim.chunks = append(sim.chunks, []map[*Particle]struct{}{})
        }
        for i := 0; i < len(sim.chunks); i++ {
            for j := len(sim.chunks[i]); j < J; j++ {
                sim.chunks[i] = append(sim.chunks[i], make(map[*Particle]struct{}))
            }
        }  
    }

    if shrunk {
        fmt.Println("Shrunk!")
    }

    sim.UpdateChunks()
}

func (sim *Simulator) UpdateChunks() {
    for i, row := range sim.chunks {
        for j, chunk := range row {
            for ptr,_ := range chunk {
                sim.wrapPosition(ptr)
                x := int(ptr.Position.X/chunkSize)
                y := int(ptr.Position.Y/chunkSize)
                if x != i || y != j {
                    sim.chunks[x][y][ptr] = struct{}{}
                    delete(chunk, ptr)
                }
            }
        }
    }
}

func (sim *Simulator) AddParticle(particle *Particle) {
    x := int(particle.Position.X/chunkSize)
    y := int(particle.Position.Y/chunkSize)
    sim.chunks[x][y][particle] = struct{}{}
}

func (sim *Simulator) Step() {

    // Compute velocity
    var wg0 sync.WaitGroup
    var wg1 sync.WaitGroup

    threadCount := len(sim.chunks) * len(sim.chunks[0])
    wg0.Add(threadCount)
    wg1.Add(threadCount)

    for _, row := range sim.chunks {
        for _, chunk := range row {
            go func(particles map[*Particle]struct{}) {
                defer wg1.Done()

                for particle, _ := range particles {
                    var force vec2.Vector
                    for _, neighbor := range sim.getNearParticles(particle.Position) {
                        if neighbor == particle {
                            continue
                        }
                        force = vec2.Add(force, sim.computeForce(particle, neighbor))
                    }
                    particle.Velocity.X += force.X
                    particle.Velocity.Y += force.Y

                    if vec2.Magnitude(particle.Velocity) > maxVelocity {
                        particle.Velocity = vec2.Scale(vec2.Unit(particle.Velocity), maxVelocity)
                    }      
                }

                wg0.Done()
                wg0.Wait() 

                for particle, _ := range particles {

                    // Modify position 
                    particle.Position.X += particle.Velocity.X
                    particle.Position.Y += particle.Velocity.Y

                    // Wrap
                    sim.wrapPosition(particle)
                }
            }(chunk)
        }
    }

    wg1.Wait()
    sim.UpdateChunks()

    sim.tick++
}

func (sim *Simulator) GetAllParticles() (particles []*Particle) {
    for _, row := range sim.chunks {
        for _, chunk := range row {
            for particle, _ := range chunk {
                particles = append(particles, particle)
            }
        }
    }

    return
}

func (sim *Simulator) getNearParticles(position vec2.Vector) (near []*Particle) {
    
    i := int(position.X/chunkSize)
    j := int(position.Y/chunkSize)

    for l := range []int{-1,0,1} {
        for k := range []int{-1,0,1} {
            a := i + l
            b := j + k
            if (a < 0 || b < 0 || a >= len(sim.chunks) || b >= len(sim.chunks[0])) {
                continue
            }
            for ptr,_ := range sim.chunks[a][b] {
                near = append(near, ptr)
            }
        }
    }

    return
}

func (sim *Simulator) computeForce(source *Particle, influence *Particle) vec2.Vector {
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

func (sim *Simulator) wrapPosition(particle *Particle) {
    for particle.Position.X  > sim.bounds.X {
        particle.Position.X -= sim.bounds.X
    }
    for particle.Position.X < 0 {
        particle.Position.X += sim.bounds.X
    }

    for particle.Position.Y  > sim.bounds.Y {
        particle.Position.Y -= sim.bounds.Y
    }
    for particle.Position.Y < 0 {
        particle.Position.Y += sim.bounds.Y
    }
}

func getInfluenceFactor(a int, b int) float64 {
    return influenceMatrix[a][b]
}