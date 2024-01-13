package vec2

import (
	"math"
)

type Vector struct {
    X, Y float64
}


func Magnitude(u Vector) float64 {
	return math.Sqrt(u.X*u.X + u.Y*u.Y)
}

func Add(u, v Vector) Vector {
	return Vector{
		X: u.X + v.X,
		Y: u.Y + v.Y,
	}
}

func Subtract(u, v Vector) Vector {
	return Vector{
		X: u.X - v.X,
		Y: u.Y - v.Y,
	}
}

func Unit(u Vector) Vector {
	m := Magnitude(u)
	return Vector{
		X: u.X/m,
		Y: u.Y/m,
	}
}

func Scale(u Vector, scalar float64) Vector {
	return Vector{
		X: u.X * scalar,
		Y: u.Y * scalar,
	}
}


