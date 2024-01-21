package simulator

import (
	"zawie/life/simulator/vec2"
	"image/color"
)
const PARTICLE_TYPE_COUNT = 8

type Particle struct {
    TypeId int
	OrganismId int
	Mass int
    Position vec2.Vector
	Velocity vec2.Vector
}

func GenerateColorShade(value int, step int) color.RGBA {
	// Define the color range (adjust these values as needed)
	minValue := 0
	maxValue := 255

	value = value*step
	// Ensure the input value is within the color range
	if value < minValue {
		value = minValue
	} else if value > maxValue {
		value = maxValue
	}

	// Calculate the color components based on the input value
	red := uint8((value * 255) / maxValue)
	green := uint8((255 - red) * 2) // Adjust formula for green component
	blue := uint8((red * 2) - 255)  // Adjust formula for blue component

	// Create an RGBA color
	rgba := color.RGBA{
		R: red,
		G: green,
		B: blue,
		A: 255, // Alpha channel, 255 for fully opaque
	}

	return rgba
}