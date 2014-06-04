package math

import (
	"math"
	"testing"
)

func TestVector_Equals(t *testing.T) {
	tests := []struct {
		A, B     Vector
		Expected bool
	}{
		{Vector{1, 0, 0, 0}, Vector{1, 0, 0, 0}, true},
		{Vector{1, 2, 3, 4}, Vector{1, 2, 3, 4}, true},
		{Vector{0.0000000000001, 0, 0, 0}, Vector{0, 0, 0, 0}, true},
		{Vector{math.MaxFloat64, 1, 0, 0}, Vector{math.MaxFloat64, 1, 0, 0}, true},
		{Vector{0, 0, 1, 0}, Vector{1, 0, 0, 0}, false},
		{Vector{1, 2, 3, 0}, Vector{-4, 5, 6, 0}, false},
	}

	for _, c := range tests {
		if r := c.A.Equals(c.B, 6); r != c.Expected {
			t.Errorf("Vector(%v).Equals(Vector(%v), 6) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func BenchmarkVector_Equals(b *testing.B) {
	b.StopTimer()
	va := Vector{math.MaxFloat64, 2, 3, 0}
	vb := Vector{5, 6, math.MaxFloat64, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		va.Equals(vb, 6)
	}
}

func TestVector_Length(t *testing.T) {
	tests := []struct {
		Value    Vector
		Expected float64
	}{
		{Vector{1, 2, 3, 1}, math.Sqrt(1*1 + 2*2 + 3*3 + 1*1)},
		{Vector{3.1, 4.2, 1.3, 0}, math.Sqrt(3.1*3.1 + 4.2*4.2 + 1.3*1.3)},
	}

	for _, c := range tests {
		if r := c.Value.Length(); !NearlyEquals(r, c.Expected, 0.000001) {
			t.Errorf("Vector(%v).Length() != %v (got %v)", c.Value, c.Expected, r)
		}
	}
}

func BenchmarkVector_Length(b *testing.B) {
	b.StopTimer()
	v := Vector{3.1, 4.2, 1.3, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		v.Length()
	}
}

func TestVector_Dot(t *testing.T) {
	tests := []struct {
		A, B     Vector
		Expected float64
	}{
		{Vector{0, 0, 0, 0}, Vector{0, 0, 0, 0}, 0},
		{Vector{1, 2, 3, 0}, Vector{4, 5, 6, 0}, 32},
		{Vector{1, 2, 3, 4}, Vector{5, 6, 7, 8}, 70},
	}

	for _, c := range tests {
		if r := c.A.Dot(c.B); !NearlyEquals(r, c.Expected, 0.000001) {
			t.Errorf("Vector(%v).Dot(Vector(%v)) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func BenchmarkVector_Dot(b *testing.B) {
	b.StopTimer()
	va := Vector{1, 2, 3, 4}
	vb := Vector{5, 6, 7, 8}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		va.Dot(vb)
	}
}

func TestVector_Cross(t *testing.T) {
	tests := []struct {
		A, B     Vector
		Expected Vector
	}{
		{Vector{1, 0, 0, 0}, Vector{0, 1, 0, 0}, Vector{0, 0, 1, 0}},
		{Vector{2, 0, 0, 0}, Vector{0, 3, 0, 0}, Vector{0, 0, 6, 0}},
		{Vector{0, 1, 0, 0}, Vector{1, 0, 0, 0}, Vector{0, 0, -1, 0}},
		{Vector{0, 0, 1, 0}, Vector{1, 0, 0, 0}, Vector{0, 1, 0, 0}},
		{Vector{1, 2, 3, 0}, Vector{-4, 5, 6, 0}, Vector{-3, -18, 13, 0}},
	}

	for _, c := range tests {
		if r := c.A.Cross(c.B); !r.Equals(c.Expected, 6) {
			t.Errorf("Vector(%v).Cross(Vector(%v)) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func BenchmarkVector_Cross(b *testing.B) {
	b.StopTimer()
	va := Vector{1, 2, 3, 4}
	vb := Vector{5, 6, 7, 8}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		va.Cross(vb)
	}
}

func TestVector_Normalize(t *testing.T) {
	tests := []struct {
		A        Vector
		Expected Vector
	}{
		{Vector{1, 2, -2}, Vector{1.0 / 3.0, 2.0 / 3.0, -2.0 / 3.0}},
		{Vector{1, 0, 0}, Vector{1, 0, 0}},
		{Vector{0, 0, 0}, Vector{0, 0, 0}},
	}

	for _, c := range tests {
		if r := c.A.Normalize(); !r.Equals(c.Expected, 6) {
			t.Errorf("Vector(%v).Normalize() != %v (got %v)", c.A, c.Expected, r)
		}
	}
}

func BenchmarkVector_Normalize(b *testing.B) {
	b.StopTimer()
	va := Vector{1, 2, 3, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		va.Normalize()
	}
}

func TestVector_Abs(t *testing.T) {
	tests := []struct {
		A        Vector
		Expected Vector
	}{
		{Vector{0, 0, 0}, Vector{0, 0, 0}},
		{Vector{1, 2, 3}, Vector{1, 2, 3}},
		{Vector{-1, -2, -3}, Vector{1, 2, 3}},
		{Vector{-1, 2, -3}, Vector{1, 2, 3}},
	}

	for _, c := range tests {
		if r := c.A.Abs(); !r.Equals(c.Expected, 6) {
			t.Errorf("Vector(%v).Abs() != %v (got %v)", c.A, c.Expected, r)
		}
	}
}

func BenchmarkVector_Abs(b *testing.B) {
	b.StopTimer()
	va := Vector{1, 2, 3, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		va.Abs()
	}
}

func TestVector_Negate(t *testing.T) {
	tests := []struct {
		A        Vector
		Expected Vector
	}{
		{Vector{0, 0, 0}, Vector{0, 0, 0}},
		{Vector{1, 2, 3}, Vector{-1, -2, -3}},
		{Vector{-1, -2, -3}, Vector{1, 2, 3}},
		{Vector{-1, 2, -3}, Vector{1, -2, 3}},
	}

	for _, c := range tests {
		if r := c.A.Negate(); !r.Equals(c.Expected, 6) {
			t.Errorf("Vector(%v).Negate() != %v (got %v)", c.A, c.Expected, r)
		}
	}
}

func BenchmarkVector_Negate(b *testing.B) {
	b.StopTimer()
	va := Vector{1, 2, 3, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		va.Negate()
	}
}

func TestVector_Add(t *testing.T) {
	tests := []struct {
		A, B     Vector
		Expected Vector
	}{
		{Vector{0, 0, 0, 0}, Vector{0, 0, 0, 0}, Vector{0, 0, 0, 0}},
		{Vector{1, 0, 0, 0}, Vector{1, 0, 0, 0}, Vector{2, 0, 0, 0}},
		{Vector{1, 2, 3, 4}, Vector{5, 6, 7, 8}, Vector{6, 8, 10, 12}},
	}

	for _, c := range tests {
		if r := c.A.Add(c.B); !r.Equals(c.Expected, 6) {
			t.Errorf("Vector(%v).Add(Vector(%v)) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func BenchmarkVector_Add(b *testing.B) {
	b.StopTimer()
	va := Vector{1, 2, 3, 0}
	vb := Vector{5, 6, 7, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		va.Add(vb)
	}
}

func TestVector_Sub(t *testing.T) {
	tests := []struct {
		A, B     Vector
		Expected Vector
	}{
		{Vector{0, 0, 0, 0}, Vector{0, 0, 0, 0}, Vector{0, 0, 0, 0}},
		{Vector{1, 0, 0, 0}, Vector{1, 0, 0, 0}, Vector{0, 0, 0, 0}},
		{Vector{1, 2, 3, 4}, Vector{5, 6, 7, 8}, Vector{-4, -4, -4, -4}},
	}

	for _, c := range tests {
		if r := c.A.Sub(c.B); !r.Equals(c.Expected, 6) {
			t.Errorf("Vector(%v).Sub(Vector(%v)) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func BenchmarkVector_Sub(b *testing.B) {
	b.StopTimer()
	va := Vector{1, 2, 3, 0}
	vb := Vector{5, 6, 7, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		va.Sub(vb)
	}
}

func TestVector_Mul(t *testing.T) {
	tests := []struct {
		A, B     Vector
		Expected Vector
	}{
		{Vector{0, 0, 0, 0}, Vector{0, 0, 0, 0}, Vector{0, 0, 0, 0}},
		{Vector{1, 0, 0, 0}, Vector{1, 0, 0, 0}, Vector{1, 0, 0, 0}},
		{Vector{1, 2, 3, 4}, Vector{5, 6, 7, 8}, Vector{5, 12, 21, 32}},
	}

	for _, c := range tests {
		if r := c.A.Mul(c.B); !r.Equals(c.Expected, 6) {
			t.Errorf("Vector(%v).Mul(Vector(%v)) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func BenchmarkVector_Mul(b *testing.B) {
	b.StopTimer()
	va := Vector{1, 2, 3, 0}
	vb := Vector{5, 6, 7, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		va.Mul(vb)
	}
}

func TestVector_MulScalar(t *testing.T) {
	tests := []struct {
		A        Vector
		B        float64
		Expected Vector
	}{
		{Vector{0, 0, 0, 0}, 1, Vector{0, 0, 0, 0}},
		{Vector{1, 0, 0, 0}, 2, Vector{2, 0, 0, 0}},
		{Vector{1, 2, 3, 4}, 3, Vector{3, 6, 9, 12}},
	}

	for _, c := range tests {
		if r := c.A.MulScalar(c.B); !r.Equals(c.Expected, 6) {
			t.Errorf("Vector(%v).MulScalar(%v) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func BenchmarkVector_MulScalar(b *testing.B) {
	b.StopTimer()
	va := Vector{1, 2, 3, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		va.MulScalar(5)
	}
}

// TODO:
func TestVector_String(t *testing.T)            {}
func TestVector_Float64(t *testing.T)           {}
func TestVector_Float32(t *testing.T)           {}
func TestVector_Clamp(t *testing.T)             {}
func TestVector_DistanceTo(t *testing.T)        {}
func TestVector_DistanceToSquared(t *testing.T) {}
