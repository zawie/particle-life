package simulator

import (
    "zawie/life/simulator/vec2"
    "golang.org/x/image/colornames"
    "image/color"
    "math/rand"
    "sync"
    "fmt"
)

type Vec2 struct {
    X, Y int
}

type Particle struct {
    Position vec2.Vector
    Id  int
	Velocity vec2.Vector
    Color color.Color //TODO: Don't store color in particle
    Mass int
    typeId int
}

const MAX_TYPE_COUNT = 16
type Chunk struct {
    particleSet map[*Particle]struct{}
    particleCount int
    typeCounts [MAX_TYPE_COUNT]int //TODO: Make dynamic based on number of types
}

type Simulator struct {
    MaxVelocity float64
    InfluenceRadius float64
    RepulsionRadius float64
    ApproximationRadius float64
    UniversalForceMultiplier float64
    ChunkSize float64
    MinimumAmountToChunk int

    tick uint64
    bounds vec2.Vector
    chunks [][]Chunk
}

var particleTypes = []color.Color{colornames.Hotpink, colornames.Limegreen, colornames.Yellow, colornames.Blue, colornames.Red}

var influenceMatrix [5][5]float64

func NewSimulator(X float64, Y float64, particleCount int) *Simulator {

    var sim Simulator = Simulator{
        MaxVelocity: 2,
        RepulsionRadius: 7.0,
        InfluenceRadius: 150.0,
        ApproximationRadius: 50.0,
        UniversalForceMultiplier: 1.0,
        MinimumAmountToChunk: 2,
    }
    sim.ChunkSize = sim.InfluenceRadius/4

    (&sim).UpdateSize(X,Y)

    particlesAdded := 0 
    for particlesAdded< particleCount {
        var p Particle
        p.Id = particlesAdded
        p.typeId = rand.Int() % len(particleTypes)
        p.Color = particleTypes[p.typeId]
        p.Position.X = rand.Float64() * sim.bounds.X
        p.Position.Y = rand.Float64() * sim.bounds.Y
        p.Velocity.X = rand.Float64() - 0.5
        p.Velocity.Y = rand.Float64() - 0.5
        p.Mass = 1
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
        I := int((X + (sim.ChunkSize-1))/sim.ChunkSize)
        J := int((Y + (sim.ChunkSize-1))/sim.ChunkSize)
        for i := len(sim.chunks); i < I; i++ {
            sim.chunks = append(sim.chunks, []Chunk{})
        }
        for i := 0; i < len(sim.chunks); i++ {
            for j := len(sim.chunks[i]); j < J; j++ {
                sim.chunks[i] = append(sim.chunks[i], Chunk{
                    particleSet: make(map[*Particle]struct{}),
                })
            }
        }  
    }

    if shrunk {
        fmt.Println("Shrunk!")
    }

    for _, particle := range sim.GetAllParticles() {
        sim.wrapPosition(particle)
    }
    sim.UpdateChunks()
}

func (sim *Simulator) UpdateChunks() {
    for i, row := range sim.chunks {
        for j, chunk := range row {
            for ptr,_ := range chunk.particleSet {
                x := int(ptr.Position.X/sim.ChunkSize)
                y := int(ptr.Position.Y/sim.ChunkSize)
                if x != i || y != j {
                    sim.chunks[x][y].particleSet[ptr] = struct{}{}
                    sim.chunks[x][y].typeCounts[ptr.typeId]++
                    sim.chunks[x][y].particleCount++
                    
                    delete(sim.chunks[i][j].particleSet, ptr)
                    sim.chunks[i][j].typeCounts[ptr.typeId]--
                    sim.chunks[i][j].particleCount--
                }
            }
        }
    }
}

func (sim *Simulator) AddParticle(particle *Particle) {
    x := int(particle.Position.X/sim.ChunkSize)
    y := int(particle.Position.Y/sim.ChunkSize)
    sim.chunks[x][y].particleSet[particle] = struct{}{}
    sim.chunks[x][y].typeCounts[particle.typeId]++
    sim.chunks[x][y].particleCount++
}

func (sim *Simulator) Step() {

    // Compute velocity
    var wg0 sync.WaitGroup
    var wg1 sync.WaitGroup

    threadCount := len(sim.chunks) * len(sim.chunks[0])
    wg0.Add(threadCount)
    wg1.Add(threadCount)

    for I, row := range sim.chunks {
        for J, _ := range row {
            go func(i int, j int) {
                defer wg1.Done()

                particles := sim.chunks[i][j].particleSet

                neighbors := sim.GetNeighborhood(sim.getChunkCenter(i,j))

                for particle, _ := range particles {
               
                    var force vec2.Vector
                    for _, neighbor := range neighbors {
                        if neighbor == particle {
                            continue
                        }
                        force = vec2.Add(force, sim.computeForce(particle, neighbor))
                    }
                    particle.Velocity.X += force.X
                    particle.Velocity.Y += force.Y

                    if vec2.Magnitude(particle.Velocity) > sim.MaxVelocity {
                        particle.Velocity = vec2.Scale(vec2.Unit(particle.Velocity), sim.MaxVelocity)
                    }      
                }

                wg0.Done()
                wg0.Wait() 

                for particle, _ := range particles {
                    // Modify position 
                    particle.Position.X += particle.Velocity.X
                    particle.Position.Y += particle.Velocity.Y
                    sim.wrapPosition(particle)
                }
            }(I, J)
        }
    }

    wg1.Wait()
    sim.UpdateChunks()

    sim.tick++
}

func (sim *Simulator) GetAllParticles() (particles []*Particle) {
    for _, row := range sim.chunks {
        for _, chunk := range row {
            for particle, _ := range chunk.particleSet {
                particles = append(particles, particle)
            }
        }
    }

    return
}

func (sim *Simulator) GetNeighborhood(position vec2.Vector) (near []*Particle) {
    
    i := int(position.X/sim.ChunkSize)
    j := int(position.Y/sim.ChunkSize)

    chunkRadiusCount := int((sim.InfluenceRadius + (sim.ChunkSize-1))/sim.ChunkSize)

    for l := -chunkRadiusCount; l <= chunkRadiusCount; l++ {
        for k := -chunkRadiusCount; k <= chunkRadiusCount; k++ {
            a := i + l
            b := j + k
            // Skip out of bounds chunks 
            // TODO: Implement wrap around
            if (a < 0 || b < 0 || a >= len(sim.chunks) || b >= len(sim.chunks[0])) {
                continue
            }
            chunk := sim.chunks[a][b]
            chunkPosition := sim.getChunkCenter(a,b)
            distance := vec2.Magnitude(vec2.Subtract(position, chunkPosition))

            if distance > sim.InfluenceRadius {
                continue
            }
            
            if distance > sim.ApproximationRadius && chunk.particleCount >= sim.MinimumAmountToChunk {
                // Give particle representing average of the chunk
                for t := 0; t < MAX_TYPE_COUNT; t++ {
                    if chunk.typeCounts[t] == 0 {
                        continue
                    }
                    // fmt.Println(chunk.typeCounts[t])
                    approxParticle := Particle{
                        Position: chunkPosition,
                        Mass: chunk.typeCounts[t],
                        typeId: t,
                    }
                    near = append(near, &approxParticle)
                }
            } else {
                 // Give all the points
                 for ptr,_ := range chunk.particleSet {
                    near = append(near, ptr)
                }
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

    if distance < sim.RepulsionRadius {
        factor -= 10*(distance - sim.RepulsionRadius)
    }
    if distance < sim.InfluenceRadius {
        factor += getInfluenceFactor(source.typeId, influence.typeId)/(distance - sim.InfluenceRadius)
    }

    return vec2.Scale(direction, factor * sim.UniversalForceMultiplier * float64(influence.Mass))
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

func (sim *Simulator) getChunkCenter(i int, j int) vec2.Vector {
    return vec2.Vector{X: (float64(i) + 0.5)*sim.ChunkSize, Y: (float64(j) + 0.5)*sim.ChunkSize}
}