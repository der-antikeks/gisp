package math

import (
	"math"
)

type Boundary struct {
	Min, Max Vector
}

func NewBoundary() Boundary {
	return Boundary{
		Min: Vector{math.Inf(1), math.Inf(1), math.Inf(1), 1},
		Max: Vector{math.Inf(-1), math.Inf(-1), math.Inf(-1), 1},
	}
}

func BoundaryFromPoints(pts ...Vector) Boundary {
	b := NewBoundary()
	for _, p := range pts {
		b.AddPoint(p)
	}

	return b
}

func (b Boundary) Equals(e Boundary, precision int) bool {
	return b.Min.Equals(e.Min, precision) && b.Max.Equals(e.Max, precision)
}

func (b *Boundary) AddPoint(p Vector) {
	b.Min[0], b.Max[0] = math.Min(b.Min[0], p[0]), math.Max(b.Max[0], p[0])
	b.Min[1], b.Max[1] = math.Min(b.Min[1], p[1]), math.Max(b.Max[1], p[1])
	b.Min[2], b.Max[2] = math.Min(b.Min[2], p[2]), math.Max(b.Max[2], p[2])
}

func (b *Boundary) AddBoundary(a Boundary) {
	if b.Equals(a, 6) {
		return
	}

	b.AddPoint(a.Max)
	b.AddPoint(a.Min)
}

func (b Boundary) Center() Vector {
	return b.Min.Add(b.Max).MulScalar(0.5)
}

func (b Boundary) Size() Vector {
	return b.Max.Sub(b.Min)
}

func (b Boundary) Sphere() (center Vector, radius float64) {
	return b.Center(), b.Size().Length() * 0.5
}
