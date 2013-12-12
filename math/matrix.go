package math

import (
	"fmt"
	"math"
)

/*	opengl - column first
	+-          -+ +-           -+ +-          -+
	| 0  4  8 12 | | 00 01 02 03 | | 1  0  0  x |
	| 1  5  9 13 | | 10 11 12 13 | | 0  1  0  y |
	| 2  6 10 14 | | 20 21 22 23 | | 0  0  1  z |
	| 3  7 11 15 | | 30 31 32 33 | | 0  0  0  1 |
	+-          -+ +-           -+ +-          -+
*/
type Matrix [16]float64

func (self Matrix) String() string {
	r := ""

	for i, n := range self {
		if i > 0 && i%4 == 0 {
			r += "\n"
		}

		r += fmt.Sprintf("%5.2f ", n)
	}

	return r
}

func (self Matrix) Float64() [16]float64 {
	m := [16]float64(self)
	return m
}

func (self Matrix) Float32() [16]float32 {
	m := [16]float32{}
	for i := range m {
		m[i] = float32(self[i])
	}

	return m
}

func Identity() Matrix {
	return Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func (self Matrix) Transform(v Vector) Vector {
	return Vector{
		self[0]*v[0] + self[4]*v[1] + self[8]*v[2] + self[12]*v[3],
		self[1]*v[0] + self[5]*v[1] + self[9]*v[2] + self[13]*v[3],
		self[2]*v[0] + self[6]*v[1] + self[10]*v[2] + self[14]*v[3],
		self[3]*v[0] + self[7]*v[1] + self[11]*v[2] + self[15]*v[3],
	}
}

func (self Matrix) Matrix3Float32() [9]float32 {
	m := [9]float32{}
	n := 0

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			i := (y * 4) + x
			m[n] = float32(self[i])
			n++
		}
	}

	return m
}

func (self Matrix) Transpose() Matrix {
	return Matrix{
		self[0], self[4], self[8], self[12],
		self[1], self[5], self[9], self[13],
		self[2], self[6], self[10], self[14],
		self[3], self[7], self[11], self[15],
	}
}

// http://en.wikipedia.org/wiki/Matrix_multiplication
func (self Matrix) Mul(m Matrix) Matrix {
	return Matrix{
		self[0]*m[0] + self[4]*m[1] + self[8]*m[2] + self[12]*m[3],
		self[1]*m[0] + self[5]*m[1] + self[9]*m[2] + self[13]*m[3],
		self[2]*m[0] + self[6]*m[1] + self[10]*m[2] + self[14]*m[3],
		self[3]*m[0] + self[7]*m[1] + self[11]*m[2] + self[15]*m[3],

		self[0]*m[4] + self[4]*m[5] + self[8]*m[6] + self[12]*m[7],
		self[1]*m[4] + self[5]*m[5] + self[9]*m[6] + self[13]*m[7],
		self[2]*m[4] + self[6]*m[5] + self[10]*m[6] + self[14]*m[7],
		self[3]*m[4] + self[7]*m[5] + self[11]*m[6] + self[15]*m[7],

		self[0]*m[8] + self[4]*m[9] + self[8]*m[10] + self[12]*m[11],
		self[1]*m[8] + self[5]*m[9] + self[9]*m[10] + self[13]*m[11],
		self[2]*m[8] + self[6]*m[9] + self[10]*m[10] + self[14]*m[11],
		self[3]*m[8] + self[7]*m[9] + self[11]*m[10] + self[15]*m[11],

		self[0]*m[12] + self[4]*m[13] + self[8]*m[14] + self[12]*m[15],
		self[1]*m[12] + self[5]*m[13] + self[9]*m[14] + self[13]*m[15],
		self[2]*m[12] + self[6]*m[13] + self[10]*m[14] + self[14]*m[15],
		self[3]*m[12] + self[7]*m[13] + self[11]*m[14] + self[15]*m[15],
	}
}

func (self Matrix) MulScalar(s float64) Matrix {
	return Matrix{
		self[0] * s, self[1] * s, self[2] * s, self[3] * s,
		self[4] * s, self[5] * s, self[6] * s, self[7] * s,
		self[8] * s, self[9] * s, self[10] * s, self[11] * s,
		self[12] * s, self[13] * s, self[14] * s, self[15] * s,
	}
}

// http://en.wikipedia.org/wiki/Scaling_matrix
func (self Matrix) Scale(v Vector) Matrix {
	return self.Mul(Matrix{
		v[0], 0, 0, 0,
		0, v[1], 0, 0,
		0, 0, v[2], 0,
		0, 0, 0, 1,
	})
}

func (self Matrix) ExtractScale() Vector {
	return Vector{
		math.Sqrt(self[0]*self[0] + self[1]*self[1] + self[2]*self[2]),
		math.Sqrt(self[4]*self[4] + self[5]*self[5] + self[6]*self[6]),
		math.Sqrt(self[8]*self[8] + self[9]*self[9] + self[10]*self[10]),
		0}
}

func (self Matrix) MaxScaleOnAxis() float64 {
	scaleX := self[0]*self[0] + self[1]*self[1] + self[2]*self[2]
	scaleY := self[4]*self[4] + self[5]*self[5] + self[6]*self[6]
	scaleZ := self[8]*self[8] + self[9]*self[9] + self[10]*self[10]

	return math.Sqrt(math.Max(scaleX, math.Max(scaleY, scaleZ)))
}

func (self Matrix) ExtractPosition() Vector {
	return Vector{
		self[12],
		self[13],
		self[14],
		0}
}

func (self *Matrix) SetPosition(position Vector) {
	self[12] = position[0]
	self[13] = position[1]
	self[14] = position[2]
}

// homogeneous rotation matrix
// http://en.wikipedia.org/wiki/Rotation_matrix
func (self Matrix) Rotate(angle float64, v Vector) Matrix {
	u := v.Normalize()
	x, y, z := u[0], u[1], u[2]
	sin := math.Sin(angle)
	cos := math.Cos(angle)
	k := 1 - cos

	return self.Mul(Matrix{
		x*x*k + cos, x*y*k + z*sin, x*z*k - y*sin, 0,
		y*x*k - z*sin, y*y*k + cos, y*z*k + x*sin, 0,
		z*x*k + y*sin, z*y*k - x*sin, z*z*k + cos, 0,
		0, 0, 0, 1,
	})
}

/*
func (self Matrix) ExtractRotation() Matrix {
	scaleX := 1 / math.Sqrt(self[0]*self[0]+self[1]*self[1]+self[2]*self[2])
	scaleY := 1 / math.Sqrt(self[4]*self[4]+self[5]*self[5]+self[6]*self[6])
	scaleZ := 1 / math.Sqrt(self[8]*self[8]+self[9]*self[9]+self[10]*self[10])

	return Matrix{
		self[0] * scaleX, self[1] * scaleX, self[2] * scaleX, 0,
		self[4] * scaleY, self[5] * scaleY, self[6] * scaleY, 0,
		self[8] * scaleZ, self[9] * scaleZ, self[10] * scaleZ, 0,
		0, 0, 0, 1,
	}
}
*/
// http://en.wikipedia.org/wiki/Translation_matrix
func (self Matrix) Translate(v Vector) Matrix {
	return self.Mul(Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		v[0], v[1], v[2], 1,
	})
}

func LookAt(eye, target, up Vector) Matrix {
	z := eye.Sub(target).Normalize()
	if z.Length() == 0 {
		z[2] = 1
	}

	x := up.Cross(z).Normalize()
	if x.Length() == 0 {
		z[0] += 0.0001
		x = up.Cross(z).Normalize()
	}

	y := z.Cross(x)

	return Matrix{
		x[0], x[1], x[2], 0,
		y[0], y[1], y[2], 0,
		z[0], z[1], z[2], 0,
		eye[0], eye[1], eye[2], 1,
	}
}

func (self Matrix) Equals(m Matrix, precision int) bool {
	p := math.Pow(10, float64(-precision))

	return (NearlyEquals(self[0], m[0], p) &&
		NearlyEquals(self[4], m[4], p) &&
		NearlyEquals(self[8], m[8], p) &&
		NearlyEquals(self[12], m[12], p) &&
		NearlyEquals(self[1], m[1], p) &&
		NearlyEquals(self[5], m[5], p) &&
		NearlyEquals(self[9], m[9], p) &&
		NearlyEquals(self[13], m[13], p) &&
		NearlyEquals(self[2], m[2], p) &&
		NearlyEquals(self[6], m[6], p) &&
		NearlyEquals(self[10], m[10], p) &&
		NearlyEquals(self[14], m[14], p) &&
		NearlyEquals(self[3], m[3], p) &&
		NearlyEquals(self[7], m[7], p) &&
		NearlyEquals(self[11], m[11], p) &&
		NearlyEquals(self[15], m[15], p))
}

// http://www.euclideanspace.com/maths/algebra/matrix/functions/determinant/index.htm
func (self Matrix) Determinant() float64 {
	return self[12]*self[9]*self[6]*self[3] - self[8]*self[13]*self[6]*self[3] - self[12]*self[5]*self[10]*self[3] + self[4]*self[13]*self[10]*self[3] +
		self[8]*self[5]*self[14]*self[3] - self[4]*self[9]*self[14]*self[3] - self[12]*self[9]*self[2]*self[7] + self[8]*self[13]*self[2]*self[7] +
		self[12]*self[1]*self[10]*self[7] - self[0]*self[13]*self[10]*self[7] - self[8]*self[1]*self[14]*self[7] + self[0]*self[9]*self[14]*self[7] +
		self[12]*self[5]*self[2]*self[11] - self[4]*self[13]*self[2]*self[11] - self[12]*self[1]*self[6]*self[11] + self[0]*self[13]*self[6]*self[11] +
		self[4]*self[1]*self[14]*self[11] - self[0]*self[5]*self[14]*self[11] - self[8]*self[5]*self[2]*self[15] + self[4]*self[9]*self[2]*self[15] +
		self[8]*self[1]*self[6]*self[15] - self[0]*self[9]*self[6]*self[15] - self[4]*self[1]*self[10]*self[15] + self[0]*self[5]*self[10]*self[15]
}

// http://www.euclideanspace.com/maths/algebra/matrix/functions/inverse/fourD/index.htm
func (self Matrix) Inverse() Matrix {
	var result Matrix

	result[0] = self[9]*self[14]*self[7] - self[13]*self[10]*self[7] + self[13]*self[6]*self[11] - self[5]*self[14]*self[11] - self[9]*self[6]*self[15] + self[5]*self[10]*self[15]
	result[4] = self[12]*self[10]*self[7] - self[8]*self[14]*self[7] - self[12]*self[6]*self[11] + self[4]*self[14]*self[11] + self[8]*self[6]*self[15] - self[4]*self[10]*self[15]
	result[8] = self[8]*self[13]*self[7] - self[12]*self[9]*self[7] + self[12]*self[5]*self[11] - self[4]*self[13]*self[11] - self[8]*self[5]*self[15] + self[4]*self[9]*self[15]
	result[12] = self[12]*self[9]*self[6] - self[8]*self[13]*self[6] - self[12]*self[5]*self[10] + self[4]*self[13]*self[10] + self[8]*self[5]*self[14] - self[4]*self[9]*self[14]
	result[1] = self[13]*self[10]*self[3] - self[9]*self[14]*self[3] - self[13]*self[2]*self[11] + self[1]*self[14]*self[11] + self[9]*self[2]*self[15] - self[1]*self[10]*self[15]
	result[5] = self[8]*self[14]*self[3] - self[12]*self[10]*self[3] + self[12]*self[2]*self[11] - self[0]*self[14]*self[11] - self[8]*self[2]*self[15] + self[0]*self[10]*self[15]
	result[9] = self[12]*self[9]*self[3] - self[8]*self[13]*self[3] - self[12]*self[1]*self[11] + self[0]*self[13]*self[11] + self[8]*self[1]*self[15] - self[0]*self[9]*self[15]
	result[13] = self[8]*self[13]*self[2] - self[12]*self[9]*self[2] + self[12]*self[1]*self[10] - self[0]*self[13]*self[10] - self[8]*self[1]*self[14] + self[0]*self[9]*self[14]
	result[2] = self[5]*self[14]*self[3] - self[13]*self[6]*self[3] + self[13]*self[2]*self[7] - self[1]*self[14]*self[7] - self[5]*self[2]*self[15] + self[1]*self[6]*self[15]
	result[6] = self[12]*self[6]*self[3] - self[4]*self[14]*self[3] - self[12]*self[2]*self[7] + self[0]*self[14]*self[7] + self[4]*self[2]*self[15] - self[0]*self[6]*self[15]
	result[10] = self[4]*self[13]*self[3] - self[12]*self[5]*self[3] + self[12]*self[1]*self[7] - self[0]*self[13]*self[7] - self[4]*self[1]*self[15] + self[0]*self[5]*self[15]
	result[14] = self[12]*self[5]*self[2] - self[4]*self[13]*self[2] - self[12]*self[1]*self[6] + self[0]*self[13]*self[6] + self[4]*self[1]*self[14] - self[0]*self[5]*self[14]
	result[3] = self[9]*self[6]*self[3] - self[5]*self[10]*self[3] - self[9]*self[2]*self[7] + self[1]*self[10]*self[7] + self[5]*self[2]*self[11] - self[1]*self[6]*self[11]
	result[7] = self[4]*self[10]*self[3] - self[8]*self[6]*self[3] + self[8]*self[2]*self[7] - self[0]*self[10]*self[7] - self[4]*self[2]*self[11] + self[0]*self[6]*self[11]
	result[11] = self[8]*self[5]*self[3] - self[4]*self[9]*self[3] - self[8]*self[1]*self[7] + self[0]*self[9]*self[7] + self[4]*self[1]*self[11] - self[0]*self[5]*self[11]
	result[15] = self[4]*self[9]*self[2] - self[8]*self[5]*self[2] + self[8]*self[1]*self[6] - self[0]*self[9]*self[6] - self[4]*self[1]*self[10] + self[0]*self[5]*self[10]

	det := self.Determinant()
	if det == 0 {
		//panic("can't invert matrix, det = 0")
		result = Identity()
	}

	return result.MulScalar(1 / det)
}

func (self Matrix) Normal() Matrix {
	return self.Inverse().Transpose()
}

func NewPerspectiveMatrix(fovy, aspect, near, far float64) Matrix {
	fovy *= DEG2RAD
	nmf := near - far
	f := 1. / math.Tan(fovy/2.0)

	return Matrix{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (near + far) / nmf, -1,
		0, 0, (2 * far * near) / nmf, 0,
	}
}

func NewFrustumMatrix(left, right, top, bottom, near, far float64) Matrix {
	rml := right - left
	tmb := top - bottom
	fmn := far - near

	a := (right + left) / rml
	b := (top + bottom) / tmb
	c := -(far + near) / fmn
	d := (2.0 * far * near) / fmn

	return Matrix{
		(2.0 * near) / rml, 0, 0, 0,
		0, (2.0 * near) / tmb, 0, 0,
		a, b, c, -1,
		0, 0, d, 0,
	}
}

func NewOrthoMatrix(left, right, bottom, top, near, far float64) Matrix {
	rml := right - left
	tmb := top - bottom
	fmn := far - near

	return Matrix{
		2 / rml, 0, 0, 0,
		0, 2 / tmb, 0, 0,
		0, 0, -2 / fmn, 0,
		-(right + left) / rml, -(top + bottom) / tmb, -(far + near) / fmn, 1,
	}
}

func ComposeMatrix(position Vector, rotation Quaternion, scale Vector) Matrix {
	return Identity().Translate(position).Mul(rotation.RotationMatrix()).Scale(scale)
}
