package math

import (
	"math"
	"testing"
)

func TestNearlyEquals(t *testing.T) {
	tests := []struct {
		A, B     float64
		Expected bool
	}{
		{1000000.0, 1000000.0, true},
		{-1000000.0, -1000000.0, true},
		{1.0000001, 1.0000002, true},
		{0.000000001000001, 0.000000001000002, true},
		{0, 0, true},
		{0.0000000000001, 0, true},
		{math.MaxFloat64, math.MaxFloat64, true},
		{math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, true},

		// Regular large numbers - generally not problematic
		{1000000.0, 1000001.0, true},
		{1000001, 1000000.0, true},
		{10000.0, 10001.0, false},
		{10001.0, 10000.0, false},

		// Negative large numbers
		{-1000000.0, -1000001.0, true},
		{-1000001.0, -1000000.0, true},
		{-10000.0, -10001.0, false},
		{-10001.0, -10000.0, false},

		// Numbers around 1
		{1.0000001, 1.0000002, true},
		{1.0000002, 1.0000001, true},
		{1.0002, 1.0001, false},
		{1.0001, 1.0002, false},

		// Numbers around -1
		{-1.000001, -1.000002, true},
		{-1.000002, -1.000001, true},
		{-1.0001, -1.0002, false},
		{-1.0002, -1.0001, false},

		// Numbers between 1 and 0
		{0.000000001000001, 0.000000001000002, true},
		{0.000000001000002, 0.000000001000001, true},
		{0.000000000001002, 0.000000000001001, false},
		{0.000000000001001, 0.000000000001002, false},

		// Numbers between -1 and 0
		{-0.000000001000001, -0.000000001000002, true},
		{-0.000000001000002, -0.000000001000001, true},
		{-0.000000000001002, -0.000000000001001, false},
		{-0.000000000001001, -0.000000000001002, false},

		// Comparisons involving zero
		{0.0, 0.0, true},
		{0.0, -0.0, true},
		{-0.0, -0.0, true},
		{0.00000001, 0.0, false},
		{0.0, 0.00000001, false},
		{-0.00000001, 0.0, false},
		{0.0, -0.00000001, false},

		// Comparisons involving infinities
		{math.Inf(1), math.Inf(1), true},
		{math.Inf(-1), math.Inf(-1), true},
		{math.Inf(-1), math.Inf(1), false},
		{math.Inf(1), math.MaxFloat64, false},
		{math.Inf(-1), -math.MaxFloat64, false},

		// Comparisons involving NaN values
		{math.NaN(), math.NaN(), false},
		{math.NaN(), 0.0, false},
		{-0.0, math.NaN(), false},
		{math.NaN(), -0.0, false},
		{0.0, math.NaN(), false},
		{math.NaN(), math.Inf(1), false},
		{math.Inf(1), math.NaN(), false},
		{math.NaN(), math.Inf(-1), false},
		{math.Inf(-1), math.NaN(), false},
		{math.NaN(), math.MaxFloat64, false},
		{math.MaxFloat64, math.NaN(), false},
		{math.NaN(), -math.MaxFloat64, false},
		{-math.MaxFloat64, math.NaN(), false},
		{math.NaN(), math.SmallestNonzeroFloat64, false},
		{math.SmallestNonzeroFloat64, math.NaN(), false},
		{math.NaN(), -math.SmallestNonzeroFloat64, false},
		{-math.SmallestNonzeroFloat64, math.NaN(), false},

		// Comparisons of numbers on opposite sides of 0
		{1.000000001, -1.0, false},
		{-1.0, 1.000000001, false},
		{-1.000000001, 1.0, false},
		{1.0, -1.000000001, false},
		{10 * math.SmallestNonzeroFloat64, 10 * -math.SmallestNonzeroFloat64, true},
		//{10000 * math.SmallestNonzeroFloat64, 10000 * -math.SmallestNonzeroFloat64, false},

		// The really tricky part - comparisons of numbers very close to zero.
		{math.SmallestNonzeroFloat64, -math.SmallestNonzeroFloat64, true},
		{-math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, true},
		{math.SmallestNonzeroFloat64, 0.0, true},
		{0.0, math.SmallestNonzeroFloat64, true},
		{-math.SmallestNonzeroFloat64, 0.0, true},
		{0.0, -math.SmallestNonzeroFloat64, true},

		{0.000000001, -math.SmallestNonzeroFloat64, false},
		{0.000000001, math.SmallestNonzeroFloat64, false},
		{math.SmallestNonzeroFloat64, 0.000000001, false},
		{-math.SmallestNonzeroFloat64, 0.000000001, false},
	}

	for _, c := range tests {
		if r := NearlyEquals(c.A, c.B, 0.000001); r != c.Expected {
			t.Errorf("NearlyEquals(%v, %v, %v) != %v (got %v)", c.A, c.B, 0.00001, c.Expected, r)
		}
	}
}

func BenchmarkNearlyEquals(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NearlyEquals(math.MaxFloat64, math.SmallestNonzeroFloat64, 0.000001)
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		Value     float64
		Precision int
		Expected  float64
	}{
		{0.5, 0, 1},
		{0.123, 2, 0.12},
		{9.99999999, 6, 10},
		{-9.99999999, 6, -10},
		{-0.000099, 4, -0.0001},
	}

	for _, c := range tests {
		if r := Round(c.Value, c.Precision); r != c.Expected {
			t.Errorf("Round(%v, %v) != %v (got %v)", c.Value, c.Precision, c.Expected, r)
		}
	}
}

func BenchmarkRound(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Round(0.123456, 4)
	}
}

func TestDeg2Rad(t *testing.T) {
	tests := []struct {
		Value    float64
		Expected float64
	}{
		{360, math.Pi * 2},
		{180, math.Pi},
		{90, math.Pi / 2},
	}

	for _, c := range tests {
		if r := c.Value * DEG2RAD; !NearlyEquals(r, c.Expected, 0.0000001) {
			t.Errorf("%v*DEG2RAD != %v (got %v)", c.Value, c.Expected, r)
		}
	}
}

func TestRad2Deg(t *testing.T) {
	tests := []struct {
		Value    float64
		Expected float64
	}{
		{math.Pi * 2, 360},
		{math.Pi, 180},
		{math.Pi / 2, 90},
	}

	for _, c := range tests {
		if r := c.Value * RAD2DEG; !NearlyEquals(r, c.Expected, 0.0000001) {
			t.Errorf("%v*RAD2DEG != %v (got %v)", c.Value, c.Expected, r)
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		Value    float64
		Expected float64
	}{
		{0.5, 0.5},
		{0.123, 0.123},
		{9.99999999, 1},
		{12456, 1},
		{-0.5, -0.5},
		{-0.123, -0.123},
		{-9.99999999, -1},
		{-12456, -1},
	}

	for _, c := range tests {
		if r := Clamp(c.Value); !NearlyEquals(r, c.Expected, 0.0000001) {
			t.Errorf("Clamp(%v) != %v (got %v)", c.Value, c.Expected, r)
		}
	}
}

func TestIsPowerOfTwo(t *testing.T) {
	tests := []struct {
		Value    int
		Expected bool
	}{
		{1, true},
		{2, true},
		{4, true},
		{1024, true},
		{1073741824, true},
		{-256, false},
		{9, false},
		{-346, false},
	}

	for _, c := range tests {
		if r := IsPowerOfTwo(c.Value); r != c.Expected {
			t.Errorf("IsPowerOfTwo(%v) != %v (got %v)", c.Value, c.Expected, r)
		}
	}
}

func TestNextHighestPowerOfTwo(t *testing.T) {
	tests := []struct {
		Value    int
		Expected int
	}{
		{3, 4},
		{7, 8},
		{255, 256},
		{1025, 2048},
		{536870913, 1073741824},
		{2, 2},
		{1024, 1024},
		{1073741824, 1073741824},
	}

	for _, c := range tests {
		if r := NextHighestPowerOfTwo(c.Value); r != c.Expected {
			t.Errorf("NextHighestPowerOfTwo(%v) != %v (got %v)", c.Value, c.Expected, r)
		}
	}
}
