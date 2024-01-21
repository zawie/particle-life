package simulator

import (
    "zawie/life/simulator/vec2"
    "sync"
    "fmt"
)

const MAX_TYPE_ID = PARTICLE_TYPE_COUNT * (MAX_POPULATION_SIZE + 2)

type Chunk struct {
    particleSet map[*Particle]struct{}
    particleCount int
    typeCounts [MAX_TYPE_ID]int
}

type Simulator struct {
    MaxSpeed float64
    InfluenceRadius float64
    RepulsionRadius float64
    ApproximationRadius float64
    UniversalForceMultiplier float64
    ChunkSize float64
    MinimumAmountToChunk int
    Tick uint64

    influenceMatrix [MAX_TYPE_ID][MAX_TYPE_ID]float64
    bounds vec2.Vector
    chunks [][]Chunk
    organismCount int
}

func NewSimulator(X float64, Y float64) *Simulator {

    var sim Simulator = Simulator{
        MaxSpeed: 100,
        RepulsionRadius: 5,
        InfluenceRadius: 100,
        ApproximationRadius: 50,
        UniversalForceMultiplier: 0.1,
        MinimumAmountToChunk: 2,
    }
    sim.ChunkSize = sim.ApproximationRadius

    (&sim).UpdateSize(X,Y)

    return &sim
}

func (sim *Simulator) AddOrganisms(organisms []*Organism) {

    for i, organism := range organisms {
        x := ((i % 5) + 1) * 400
        y := ((i / 5) + 1) * 400
        sim.AddOrganism(organism, vec2.Vector{X:float64(x) , Y: float64(y)})
    }
}

func (sim *Simulator) AddOrganism(organism *Organism, position vec2.Vector) {
    sim.organismCount++
    id := sim.organismCount

    organism.Position(position)
    for _, particle := range organism.GetAllParticles() {
        particle.OrganismId = id
        particle.TypeId +=  id*PARTICLE_TYPE_COUNT + particle.TypeId
        sim.AddParticle(particle)
    }


    for i := 0; i < PARTICLE_TYPE_COUNT; i++ {
        I := id*PARTICLE_TYPE_COUNT + i
		for J := 0; J < MAX_TYPE_ID; J++ {
            j := J % PARTICLE_TYPE_COUNT
            sim.influenceMatrix[I][J] = organism.GetExternalInfluenceFactor(i, j)
		}
	}

	for i := 0; i < PARTICLE_TYPE_COUNT; i++ {
        I := id*PARTICLE_TYPE_COUNT + i
		for j := 0; j < PARTICLE_TYPE_COUNT; j++ {
            J := id*PARTICLE_TYPE_COUNT + j
            sim.influenceMatrix[I][J] = organism.GetInternalInfluenceFactor(i, j)
		}
	}

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
                    sim.chunks[x][y].typeCounts[ptr.TypeId]++
                    sim.chunks[x][y].particleCount++
                    
                    delete(sim.chunks[i][j].particleSet, ptr)
                    sim.chunks[i][j].typeCounts[ptr.TypeId]--
                    sim.chunks[i][j].particleCount--
                }
            }
        }
    }
}

func (sim *Simulator) AddParticle(particle *Particle) {
    sim.wrapPosition(particle)
    x := int(particle.Position.X/sim.ChunkSize)
    y := int(particle.Position.Y/sim.ChunkSize)
    sim.chunks[x][y].particleSet[particle] = struct{}{}
    sim.chunks[x][y].typeCounts[particle.TypeId]++
    sim.chunks[x][y].particleCount++
}

func (sim *Simulator) Step() {
    var wg0 sync.WaitGroup
    var wg1 sync.WaitGroup

    threadCount := len(sim.chunks)
    wg0.Add(threadCount)
    wg1.Add(threadCount)

    for I, row := range sim.chunks {
        go func(i int) {
            defer wg1.Done()

            for j, _ := range row {
                sim.ComputeForceInChunk(i,j)
            }

            wg0.Done()
            wg0.Wait() 

            for j, _ := range row {
                sim.ComputePositionInChunk(i,j)
            }
        }(I)
    }

    wg1.Wait()
    sim.UpdateChunks()

    sim.Tick++
}

func (sim *Simulator) ComputeForceInChunk(i, j int) {
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

        speed := vec2.Magnitude(particle.Velocity)

        // Cap speed 
        if speed > sim.MaxSpeed {
            speed = sim.MaxSpeed
        }   

        // Add air resistance
        speed = speed - 0.05*(speed*speed)

        // Cap speed
        if speed < 0 {
            speed = 0.001
        }

        particle.Velocity = vec2.Scale(vec2.Unit(particle.Velocity), speed)
    }
}

func (sim *Simulator) ComputePositionInChunk(i, j int) {
    for particle, _ := range sim.chunks[i][j].particleSet {
        // Modify position 
        particle.Position.X += particle.Velocity.X
        particle.Position.Y += particle.Velocity.Y
        sim.wrapPosition(particle)
    }
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
                for t := 0; t < PARTICLE_TYPE_COUNT; t++ {
                    if chunk.typeCounts[t] == 0 {
                        continue
                    }
                    // fmt.Println(chunk.typeCounts[t])
                    approxParticle := Particle{
                        Position: chunkPosition,
                        Mass: chunk.typeCounts[t],
                        TypeId: t,
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

func (sim *Simulator) ComputeAverageKineticEnergy() (energy float64) {
    particles := sim.GetAllParticles()
    count := len(particles)
    for _, particle := range particles {
        speed := vec2.Magnitude(particle.Velocity)
        energy += speed*speed
    }
    energy = energy/float64(count)
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
        factor += sim.getInfluenceFactor(source.TypeId, influence.TypeId)/(distance - sim.InfluenceRadius)
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

func (sim *Simulator) getInfluenceFactor(a int, b int) float64 {
    return sim.influenceMatrix[a][b]
}

func (sim *Simulator) getChunkCenter(i int, j int) vec2.Vector {
    return vec2.Vector{X: (float64(i) + 0.5)*sim.ChunkSize, Y: (float64(j) + 0.5)*sim.ChunkSize}
}