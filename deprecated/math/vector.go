package math

import (
	"fmt"
	"math"
)

type Vector [4]float64

func (self Vector) String() string {
	return fmt.Sprintf("%5.2f %5.2f %5.2f %5.2f", self[0], self[1], self[2], self[3])
}

func (self Vector) Equals(v Vector, precision int) bool {
	p := math.Pow(10, float64(-precision))

	return (NearlyEquals(self[0], v[0], p) &&
		NearlyEquals(self[1], v[1], p) &&
		NearlyEquals(self[2], v[2], p) &&
		NearlyEquals(self[3], v[3], p))
}

func (self Vector) Float64() [4]float64 {
	return [4]float64(self)
}

func (self Vector) Float32() [4]float32 {
	v := [4]float32{}
	for i := range v {
		v[i] = float32(self[i])
	}

	return v
}

func (self Vector) Length() float64 {
	return math.Sqrt(self[0]*self[0] + self[1]*self[1] + self[2]*self[2] + self[3]*self[3])
}

func (self Vector) Dot(v Vector) float64 {
	return self[0]*v[0] + self[1]*v[1] + self[2]*v[2] + self[3]*v[3]
}

func (self Vector) Normalize() Vector {
	l := self.Length()
	if l == 0 {
		return self
	}

	return Vector{
		self[0] / l,
		self[1] / l,
		self[2] / l,
		self[3] / l,
	}
}

func (self Vector) Add(v Vector) Vector {
	return Vector{
		self[0] + v[0],
		self[1] + v[1],
		self[2] + v[2],
		self[3] + v[3],
	}
}

func (self Vector) Sub(v Vector) Vector {
	return Vector{
		self[0] - v[0],
		self[1] - v[1],
		self[2] - v[2],
		self[3] - v[3],
	}
}

func (self Vector) Mul(v Vector) Vector {
	return Vector{
		self[0] * v[0],
		self[1] * v[1],
		self[2] * v[2],
		self[3] * v[3],
	}
}

func (self Vector) MulScalar(s float64) Vector {
	return Vector{
		self[0] * s,
		self[1] * s,
		self[2] * s,
		self[3] * s,
	}
}

func (self Vector) Negate() Vector {
	return Vector{
		-self[0],
		-self[1],
		-self[2],
		-self[3],
	}
}

func (self Vector) Abs() Vector {
	return Vector{
		math.Abs(self[0]),
		math.Abs(self[1]),
		math.Abs(self[2]),
		math.Abs(self[3]),
	}
}

func (self Vector) Cross(v Vector) Vector {
	return Vector{
		self[1]*v[2] - self[2]*v[1],
		self[2]*v[0] - self[0]*v[2],
		self[0]*v[1] - self[1]*v[0],
		/* TODO: w? */
	}
}

func (self Vector) Clamp(min, max Vector) (result Vector) {
	if self[0] < min[0] {
		result[0] = min[0]
	} else if self[0] > max[0] {
		result[0] = max[0]
	} else {
		result[0] = self[0]
	}

	if self[1] < min[1] {
		result[1] = min[1]
	} else if self[1] > max[1] {
		result[1] = max[1]
	} else {
		result[1] = self[1]
	}

	if self[2] < min[2] {
		result[2] = min[2]
	} else if self[2] > max[2] {
		result[2] = max[2]
	} else {
		result[2] = self[2]
	}

	return
}

// Euclidean norm!
func (self Vector) DistanceTo(v Vector) float64 {
	return math.Sqrt(self.DistanceToSquared(v))
}

func (self Vector) DistanceToSquared(v Vector) float64 {
	return self.Dot(v)
}
