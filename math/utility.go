package math

import (
	"math"
)

const (
	DEG2RAD = math.Pi / 180
	RAD2DEG = 180 / math.Pi
	Pi      = math.Pi
)

func Round(v float64, precision int) float64 {
	var r float64

	if tmp := v * math.Pow(10, float64(precision)); tmp > 0 {
		r = math.Floor(tmp + 0.5)
	} else {
		r = math.Ceil(tmp - 0.5)
	}

	return r / math.Pow(10, float64(precision))
}

/*
	NearlyEquals compares two float64 with an error margin
	http://floating-point-gui.de/errors/comparison/
*/
func NearlyEquals(a, b, epsilon float64) bool {
	// shortcut, handles infinities
	if a == b {
		return true
	}

	diff := math.Abs(a - b)

	// a or b or both are zero
	if a*b == 0 {
		return diff < (epsilon * epsilon)
	}

	absA := math.Abs(a)
	absB := math.Abs(b)

	// use relative error
	return diff/(absA+absB) < epsilon
}

func Clamp(v float64) float64 {
	return math.Min(math.Max(v, -1), 1)
}

func IsPowerOfTwo(n int) bool {
	return (n & (n - 1)) == 0
}

func NextHighestPowerOfTwo(n int) int {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}
