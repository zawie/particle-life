package quadtree

import (
	"zawie/life/simulator/vec2"
)

type QuadtreeNode struct {
	Bounds    Rect
	Particles    []Particle
	Children  [4]*QuadtreeNode
	MaxPoints int
}

type Rect struct {
	X, Y, Width, Height int
}

func NewQuadtreeNode(bounds Rect, maxParticles int) *QuadtreeNode {
	return &QuadtreeNode{
		Bounds:    bounds,
		Particles:    make([]Particle, 0),
		Children:  [4]*QuadtreeNode{nil, nil, nil, nil},
		maxParticles: maxParticles,
	}
}

func (node *QuadtreeNode) Insert(particle Particle) {
	if !node.Bounds.contains(particle) {
		return
	}

	if len(node.Particles) < node.MaxPoints {
		node.Particles = append(node.Particles, particle)
		return
	}

	if node.Children[0] == nil {
		node.split()
	}

	for i := 0; i < 4; i++ {
		node.Children[i].Insert(point)
	}
}

func (node *QuadtreeNode) SearchRange(rangeRect Rect) []Particle {
	result := make([]Particle, 0)

	if !node.Bounds.intersects(rangeRect) {
		return result
	}

	for _, point := range node.Points {
		if rangeRect.contains(point) {
			result = append(result, point)
		}
	}

	if node.Children[0] != nil {
		for i := 0; i < 4; i++ {
			result = append(result, node.Children[i].SearchRange(rangeRect)...)
		}
	}

	return result
}

func (node *QuadtreeNode) split() {
	width := node.Bounds.Width / 2
	height := node.Bounds.Height / 2
	x := node.Bounds.X
	y := node.Bounds.Y

	node.Children[0] = NewQuadtreeNode(Rect{x, y, width, height}, node.MaxPoints)
	node.Children[1] = NewQuadtreeNode(Rect{x + width, y, width, height}, node.MaxPoints)
	node.Children[2] = NewQuadtreeNode(Rect{x, y + height, width, height}, node.MaxPoints)
	node.Children[3] = NewQuadtreeNode(Rect{x + width, y + height, width, height}, node.MaxPoints)

	for _, point := range node.Points {
		for i := 0; i < 4; i++ {
			node.Children[i].Insert(point)
		}
	}

	node.Points = nil
}

func (rect Rect) contains(point Particle) bool {
	return point.X >= rect.X && point.X <= rect.X+rect.Width &&
		point.Y >= rect.Y && point.Y <= rect.Y+rect.Height
}

func (rect Rect) intersects(other Rect) bool {
	return rect.X < other.X+other.Width && rect.X+rect.Width > other.X &&
		rect.Y < other.Y+other.Height && rect.Y+rect.Height > other.Y
}