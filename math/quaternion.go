package math

import (
	"fmt"
	"math"
)

type Quaternion [4]float64

func (self Quaternion) String() string {
	return fmt.Sprintf("%5.2f %5.2f %5.2f %5.2f", self[0], self[1], self[2], self[3])
}

func (self Quaternion) Equals(q Quaternion, precision int) bool {
	p := math.Pow(10, float64(-precision))

	return (NearlyEquals(self[0], q[0], p) &&
		NearlyEquals(self[1], q[1], p) &&
		NearlyEquals(self[2], q[2], p) &&
		NearlyEquals(self[3], q[3], p))
}

// http://www.euclideanspace.com/maths/algebra/realNormedAlgebra/quaternions/code/index.htm
func (self Quaternion) Length() float64 {
	return math.Sqrt(self[0]*self[0] + self[1]*self[1] + self[2]*self[2] + self[3]*self[3])
}

func (self Quaternion) Normalize() Quaternion {
	l := self.Length()
	if l == 0 {
		return Quaternion{
			0,
			0,
			0,
			1,
		}
	}

	return Quaternion{
		self[0] / l,
		self[1] / l,
		self[2] / l,
		self[3] / l,
	}
}

func (self Quaternion) Conjugate() Quaternion {
	return Quaternion{
		-self[0],
		-self[1],
		-self[2],
		self[3],
	}
}

func (self Quaternion) Mul(q Quaternion) Quaternion {
	return Quaternion{
		self[0]*q[3] + self[1]*q[2] - self[2]*q[1] + self[3]*q[0],
		-self[0]*q[2] + self[1]*q[3] + self[2]*q[0] + self[3]*q[1],
		self[0]*q[1] - self[1]*q[0] + self[2]*q[3] + self[3]*q[2],
		-self[0]*q[0] - self[1]*q[1] - self[2]*q[2] + self[3]*q[3],
	}
}

func (self Quaternion) MulScalar(s float64) Quaternion {
	return Quaternion{
		self[0] * s,
		self[1] * s,
		self[2] * s,
		self[3] * s,
	}
}

func (self Quaternion) Add(q Quaternion) Quaternion {
	return Quaternion{
		self[0] + q[0],
		self[1] + q[1],
		self[2] + q[2],
		self[3] + q[3],
	}
}

func (self Quaternion) Sub(q Quaternion) Quaternion {
	return Quaternion{
		self[0] - q[0],
		self[1] - q[1],
		self[2] - q[2],
		self[3] - q[3],
	}
}

/*
	Slerp generates a quaternion between two given quaternions in proportion to the variable t
	if t=0 it returns q, if t=1 then self, if t is between them returned quaternion will interpolate between them.
	http://www.euclideanspace.com/maths/algebra/realNormedAlgebra/quaternions/slerp/index.htm
*/
func (self Quaternion) Slerp(q Quaternion, t float64) Quaternion {
	cosHalfTheta := self[3]*q[3] + self[0]*q[0] + self[1]*q[1] + self[2]*q[2]
	if math.Abs(cosHalfTheta) >= 1.0 {
		return self
	}

	halfTheta := math.Acos(cosHalfTheta)
	sinHalfTheta := math.Sqrt(1.0 - cosHalfTheta*cosHalfTheta)
	if math.Abs(sinHalfTheta) < 0.001 {
		return Quaternion{
			(self[0]*0.5 + q[0]*0.5),
			(self[1]*0.5 + q[1]*0.5),
			(self[2]*0.5 + q[2]*0.5),
			(self[3]*0.5 + q[3]*0.5),
		}
	}

	ratioA := math.Sin((1.0-t)*halfTheta) / sinHalfTheta
	ratioB := math.Sin(t*halfTheta) / sinHalfTheta
	return Quaternion{
		(self[0]*ratioA + q[0]*ratioB),
		(self[1]*ratioA + q[1]*ratioB),
		(self[2]*ratioA + q[2]*ratioB),
		(self[3]*ratioA + q[3]*ratioB),
	}
}

// http://www.euclideanspace.com/maths/geometry/rotations/conversions/angleToQuaternion/index.htm
func QuaternionFromAxisAngle(axis Vector, angle float64) Quaternion {
	halfAngle := angle / 2
	s := math.Sin(halfAngle)
	axis = axis.Normalize()

	return Quaternion{
		axis[0] * s,         // x, i
		axis[1] * s,         // y, j
		axis[2] * s,         // z, k
		math.Cos(halfAngle), // w, d
	}
}

type EulerAngleOrder string

const (
	RotateXYZ EulerAngleOrder = "XYZ"
	RotateYZX                 = "YZX"
	RotateZXY                 = "ZXY"
	RotateXZY                 = "XZY"
	RotateYXZ                 = "YXZ"
	RotateZYX                 = "ZYX"

	DefaultOrder = RotateYZX // RotateXYZ
)

// http://www.euclideanspace.com/maths/geometry/rotations/conversions/eulerToQuaternion/index.htm
func QuaternionFromEuler(e Vector, order EulerAngleOrder) Quaternion {
	c1, s1 := math.Cos(e[0]/2), math.Sin(e[0]/2)
	c2, s2 := math.Cos(e[1]/2), math.Sin(e[1]/2)
	c3, s3 := math.Cos(e[2]/2), math.Sin(e[2]/2)

	var q Quaternion

	switch order {
	case RotateXYZ:
		q[0] = s1*c2*c3 + c1*s2*s3
		q[1] = c1*s2*c3 - s1*c2*s3
		q[2] = c1*c2*s3 + s1*s2*c3
		q[3] = c1*c2*c3 - s1*s2*s3
	case RotateYXZ:
		q[0] = s1*c2*c3 + c1*s2*s3
		q[1] = c1*s2*c3 - s1*c2*s3
		q[2] = c1*c2*s3 - s1*s2*c3
		q[3] = c1*c2*c3 + s1*s2*s3
	case RotateZXY:
		q[0] = s1*c2*c3 - c1*s2*s3
		q[1] = c1*s2*c3 + s1*c2*s3
		q[2] = c1*c2*s3 + s1*s2*c3
		q[3] = c1*c2*c3 - s1*s2*s3
	case RotateZYX:
		q[0] = s1*c2*c3 - c1*s2*s3
		q[1] = c1*s2*c3 + s1*c2*s3
		q[2] = c1*c2*s3 - s1*s2*c3
		q[3] = c1*c2*c3 + s1*s2*s3
	case RotateYZX:
		q[0] = s1*c2*c3 + c1*s2*s3
		q[1] = c1*s2*c3 + s1*c2*s3
		q[2] = c1*c2*s3 - s1*s2*c3
		q[3] = c1*c2*c3 - s1*s2*s3
	case RotateXZY:
		q[0] = s1*c2*c3 - c1*s2*s3
		q[1] = c1*s2*c3 - s1*c2*s3
		q[2] = c1*c2*s3 + s1*s2*c3
		q[3] = c1*c2*c3 + s1*s2*s3
	}

	return q
}

// http://www.euclideanspace.com/maths/geometry/rotations/conversions/matrixToQuaternion/index.htm
func QuaternionFromRotationMatrix(m Matrix) Quaternion {
	tr := m[0] + m[5] + m[10]

	var q Quaternion

	if tr > 0 {
		s := 0.5 / math.Sqrt(tr+1.0)

		q[0] = (m[6] - m[9]) * s
		q[1] = (m[8] - m[2]) * s
		q[2] = (m[1] - m[4]) * s
		q[3] = 0.25 / s
	} else if (m[0] > m[5]) && (m[0] > m[10]) {
		s := 2.0 * math.Sqrt(1.0+m[0]-m[5]-m[10])

		q[0] = 0.25 * s
		q[1] = (m[4] + m[1]) / s
		q[2] = (m[8] + m[2]) / s
		q[3] = (m[6] - m[9]) / s
	} else if m[5] > m[10] {
		s := 2.0 * math.Sqrt(1.0+m[5]-m[0]-m[10])

		q[0] = (m[4] + m[1]) / s
		q[1] = 0.25 * s
		q[2] = (m[9] + m[6]) / s
		q[3] = (m[8] - m[2]) / s
	} else {
		s := 2.0 * math.Sqrt(1.0+m[10]-m[0]-m[5])

		q[0] = (m[8] + m[2]) / s
		q[1] = (m[9] + m[6]) / s
		q[2] = 0.25 * s
		q[3] = (m[1] - m[4]) / s
	}

	return q
}

// http://www.euclideanspace.com/maths/geometry/rotations/conversions/quaternionToMatrix/index.htm
func (q Quaternion) RotationMatrix() Matrix {
	x, y, z, w := q[0], q[1], q[2], q[3]
	x2, y2, z2 := x+x, y+y, z+z
	xx, xy, xz := x*x2, x*y2, x*z2
	yy, yz, zz := y*y2, y*z2, z*z2
	wx, wy, wz := w*x2, w*y2, w*z2

	return Matrix{
		1 - (yy + zz), xy + wz, xz - wy, 0,
		xy - wz, 1 - (xx + zz), yz + wx, 0,
		xz + wy, yz - wx, 1 - (xx + yy), 0,
		0, 0, 0, 1,
	}
}
