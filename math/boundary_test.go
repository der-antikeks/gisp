package math

import (
	"math"
	"testing"
)

func TestBoundary_Equals(t *testing.T) {
	tests := []struct {
		A, B     Boundary
		Expected bool
	}{
		{
			Boundary{Vector{0, 0, 0}, Vector{0, 0, 0}},
			Boundary{Vector{0, 0, 0}, Vector{0, 0, 0}},
			true,
		},
		{
			Boundary{Vector{0, 0, 0}, Vector{1, 1, 1}},
			Boundary{Vector{0, 0, 0}, Vector{1, 1, 1}},
			true,
		},
		{
			Boundary{Vector{0, 0, 0}, Vector{1, 1, 1}},
			Boundary{Vector{1, 1, 1}, Vector{0, 0, 0}},
			false,
		},
		{
			Boundary{Vector{1, 2, 3}, Vector{5, 6, 7}},
			Boundary{Vector{1, 2, 3}, Vector{5, 6, 7}},
			true,
		},
		{
			Boundary{Vector{0.0000000000001, 0.0000000000001, 0.0000000000001}, Vector{1, 1, 1}},
			Boundary{Vector{0, 0, 0}, Vector{1, 1, 1}},
			true,
		},
		{
			Boundary{Vector{0, 0.00000001, 0}, Vector{1, 1, 1}},
			Boundary{Vector{0, 0, 0}, Vector{1, 1, 1}},
			false,
		},
	}

	for _, c := range tests {
		if r := c.A.Equals(c.B, 6); r != c.Expected {
			t.Errorf("Boundary(%v).Equals(Boundary(%v), 6) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func TestNewBoundary(t *testing.T) {
	tests := []struct {
		B        Boundary
		Expected Boundary
	}{
		{
			NewBoundary(),
			Boundary{
				Vector{math.Inf(1), math.Inf(1), math.Inf(1), 1},
				Vector{math.Inf(-1), math.Inf(-1), math.Inf(-1), 1},
			},
		},
	}

	for _, c := range tests {
		if !c.B.Equals(c.Expected, 6) {
			t.Errorf("Boundary(%v) != Boundary(%v)", c.B, c.Expected)
		}
	}
}

func TestBoundaryFromPoints(t *testing.T) {
	tests := []struct {
		Points   []Vector
		Expected Boundary
	}{
		{
			[]Vector{},
			NewBoundary(),
		},
		{
			[]Vector{Vector{0, 0, 0}},
			Boundary{Vector{0, 0, 0, 1}, Vector{0, 0, 0, 1}},
		},
		{
			[]Vector{Vector{-1, 0, 0}, Vector{0, 1, 0}},
			Boundary{Vector{-1, 0, 0, 1}, Vector{0, 1, 0, 1}},
		},
		{
			[]Vector{Vector{-1, 0, 0}, Vector{0, 1, 0}, Vector{1, 0, -1}},
			Boundary{Vector{-1, 0, -1, 1}, Vector{1, 1, 0, 1}},
		},
	}

	for _, c := range tests {
		if r := BoundaryFromPoints(c.Points...); !r.Equals(c.Expected, 6) {
			t.Errorf("BoundaryFromPoints(%v) != %v (got %v)", c.Points, c.Expected, r)
		}
	}
}

func TestBoundary_AddPoint(t *testing.T) {
	b := NewBoundary()

	tests := []struct {
		Point    Vector
		Expected Boundary
	}{
		{
			Vector{0, 0, 0},
			Boundary{Vector{0, 0, 0, 1}, Vector{0, 0, 0, 1}},
		},
		{
			Vector{-1, 0, 0},
			Boundary{Vector{-1, 0, 0, 1}, Vector{0, 0, 0, 1}},
		},
		{
			Vector{1, -1, 1},
			Boundary{Vector{-1, -1, 0, 1}, Vector{1, 0, 1, 1}},
		},
	}

	for _, c := range tests {
		if b.AddPoint(c.Point); !b.Equals(c.Expected, 6) {
			t.Errorf("Boundary().AddPoint(%v) != %v (got %v)", c.Point, c.Expected, b)
		}
	}
}

func TestBoundary_Center(t *testing.T) {
	tests := []struct {
		B        Boundary
		Expected Vector
	}{
		{
			BoundaryFromPoints(Vector{0, 0, 0}),
			Vector{0, 0, 0, 1},
		},
		{
			BoundaryFromPoints(Vector{-1, 0, 0}, Vector{1, 0, 0}),
			Vector{0, 0, 0, 1},
		},
		{
			BoundaryFromPoints(Vector{0, 0, 0}, Vector{1, 2, 3}),
			Vector{1.0 / 2.0, 2.0 / 2.0, 3.0 / 2.0, 1},
		},
	}

	for _, c := range tests {
		if r := c.B.Center(); !r.Equals(c.Expected, 6) {
			t.Errorf("Boundary(%v).Center() != %v (got %v)", c.B, c.Expected, r)
		}
	}
}

func TestBoundary_Size(t *testing.T) {
	tests := []struct {
		B        Boundary
		Expected Vector
	}{
		{
			BoundaryFromPoints(Vector{0, 0, 0}),
			Vector{0, 0, 0, 0},
		},
		{
			BoundaryFromPoints(Vector{-1, 0, 0}, Vector{1, 0, 0}),
			Vector{2, 0, 0, 0},
		},
		{
			BoundaryFromPoints(Vector{0, 0, 0}, Vector{1, 2, 3}),
			Vector{1, 2, 3, 0},
		},
	}

	for _, c := range tests {
		if r := c.B.Size(); !r.Equals(c.Expected, 6) {
			t.Errorf("Boundary(%v).Size() != %v (got %v)", c.B, c.Expected, r)
		}
	}
}

func TestBoundary_Sphere(t *testing.T) {
	tests := []struct {
		B         Boundary
		ExpCenter Vector
		ExpRadius float64
	}{
		{
			BoundaryFromPoints(Vector{0, 0, 0}),
			Vector{0, 0, 0, 1},
			0,
		},
		{
			BoundaryFromPoints(Vector{-1, 0, 0}, Vector{1, 0, 0}),
			Vector{0, 0, 0, 1},
			2.0 / 2.0,
		},
		{
			BoundaryFromPoints(Vector{0, 0, 0}, Vector{1, 2, 3}),
			Vector{1.0 / 2.0, 2.0 / 2.0, 3.0 / 2.0, 1},
			math.Sqrt(1.0+2.0*2.0+3.0*3.0) / 2.0,
		},
	}

	for _, c := range tests {
		if ce, ra := c.B.Sphere(); !ce.Equals(c.ExpCenter, 6) || !NearlyEquals(ra, c.ExpRadius, math.Pow(10, float64(-6))) {
			t.Errorf("Boundary(%v).Sphere() != %v, %v (got %v, %v)", c.B, c.ExpCenter, c.ExpRadius, ce, ra)
		}
	}
}
