package math

import (
	"math"
	"testing"
)

func TestQuaternion_Equals(t *testing.T) {
	tests := []struct {
		A, B     Quaternion
		Expected bool
	}{
		{Quaternion{1, 0, 0, 0}, Quaternion{1, 0, 0, 0}, true},
		{Quaternion{1, 2, 3, 4}, Quaternion{1, 2, 3, 4}, true},
		{Quaternion{0.0000000000001, 0, 0, 0}, Quaternion{0, 0, 0, 0}, true},
		{Quaternion{math.MaxFloat64, 1, 0, 0}, Quaternion{math.MaxFloat64, 1, 0, 0}, true},
		{Quaternion{0, 0, 1, 0}, Quaternion{1, 0, 0, 0}, false},
		{Quaternion{1, 2, 3, 0}, Quaternion{-4, 5, 6, 0}, false},
	}

	for _, c := range tests {
		if r := c.A.Equals(c.B, 6); r != c.Expected {
			t.Errorf("Quaternion(%v).Equals(Quaternion(%v), 6) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func TestQuaternionFromAxisAngle(t *testing.T) {
	tests := []struct {
		Axis     Vector
		Angle    float64
		Expected Quaternion
	}{
		{Vector{1, 0, 0}, 0.0 * DEG2RAD, Quaternion{0, 0, 0, 1}},
		{Vector{1, 0, 0}, 45.0 * DEG2RAD, Quaternion{0.382, 0, 0, 0.923}},
		{Vector{0, 1, 0}, 90.0 * DEG2RAD, Quaternion{0, 0.707, 0, 0.707}},
		{Vector{1, 1, 0}, -13.0 * DEG2RAD, Quaternion{-0.080, -0.080, 0, 0.996}},
		//{Vector{0, 1, 1}, 128.0 * DEG2RAD, Quaternion{0, 0.635, 0.635, 0.772}},
		{Vector{-0.5774, -0.5774, -0.5774}, 120.0 * DEG2RAD, Quaternion{-0.5, -0.5, -0.5, 0.5}},
	}

	for _, c := range tests {
		if r := QuaternionFromAxisAngle(c.Axis, c.Angle); !r.Equals(c.Expected, 2) {
			t.Errorf("QuaternionFromAxisAngle(%v, %v) != %v (got %v)", c.Axis, c.Angle, c.Expected, r)
		}
	}
}

func TestQuaternionFromEuler(t *testing.T) {
	tests := []struct {
		Euler    Vector
		Order    EulerAngleOrder
		Expected Quaternion
	}{
		{Vector{0, 0, 0}, DefaultOrder, Quaternion{0, 0, 0, 1}},
		{Vector{0, 45.0 * DEG2RAD, 0}, DefaultOrder, Quaternion{0, 0.3826834323650898, 0, 0.9238795325112867}},
		{Vector{0, 0, 90.0 * DEG2RAD}, DefaultOrder, Quaternion{0, 0, 0.7071067811865475, 0.7071067811865476}},
		{Vector{128.0 * DEG2RAD, 0, 0}, DefaultOrder, Quaternion{0.898794046299167, 0, 0, 0.43837114678907746}},
		{Vector{146 * DEG2RAD, 12 * DEG2RAD, 28 * DEG2RAD}, DefaultOrder, Quaternion{0.9302087080763699, 0.2597370618110635, -0.026648151107188656, 0.25795017771502843}},

		{Vector{45.0 * DEG2RAD, 90.0 * DEG2RAD, 180.0 * DEG2RAD}, RotateXYZ, Quaternion{0.6532814824, -0.2705980501, 0.6532814824, -0.2705980501}},
		{Vector{45.0 * DEG2RAD, 90.0 * DEG2RAD, 180.0 * DEG2RAD}, RotateYXZ, Quaternion{0.6532814824, -0.2705980501, 0.6532814824, 0.2705980501}},
		{Vector{45.0 * DEG2RAD, 90.0 * DEG2RAD, 180.0 * DEG2RAD}, RotateZXY, Quaternion{-0.6532814824, 0.2705980501, 0.6532814824, -0.2705980501}},
		{Vector{45.0 * DEG2RAD, 90.0 * DEG2RAD, 180.0 * DEG2RAD}, RotateZYX, Quaternion{-0.6532814824, 0.2705980501, 00.6532814824, 0.2705980501}},
		{Vector{45.0 * DEG2RAD, 90.0 * DEG2RAD, 180.0 * DEG2RAD}, RotateXZY, Quaternion{-0.6532814824, -0.2705980501, 0.6532814824, 0.2705980501}},
	}

	for _, c := range tests {
		if r := QuaternionFromEuler(c.Euler, c.Order); !r.Equals(c.Expected, 6) {
			t.Errorf("QuaternionFromEuler(%v, %v) != %v (got %v)", c.Euler, c.Order, c.Expected, r)
		}
	}
}

func TestQuaternionFromRotationMatrix(t *testing.T) {
	tests := []struct {
		Rotation Matrix
		Expected Quaternion
	}{
		{
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
			Quaternion{0, 0, 0, 1},
		},
		{
			Matrix{
				0, 0, -1, 0,
				0, 1, 0, 0,
				1, 0, 0, 0,
				0, 0, 0, 1,
			},
			Quaternion{0, 0.7071, 0, 0.7071},
		},
		{
			Matrix{
				0, 1, 0, 0,
				-1, 0, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
			Quaternion{0, 0, 0.7071, 0.7071},
		},
		{
			Matrix{
				0, 1, 0, 0,
				0, 0, -1, 0,
				-1, 0, 0, 0,
				0, 0, 0, 1,
			},
			Quaternion{-0.5, -0.5, 0.5, 0.5},
		},
		{
			Matrix{
				0, 1, 0, 0,
				1, 0, 0, 0,
				0, 0, -1, 0,
				0, 0, 0, 1,
			},
			Quaternion{0.7071067811865475, 0.7071067811865475, 0, 0},
		},
		{
			Matrix{
				-1, 0, 0, 0,
				0, 0, 1, 0,
				0, 1, 0, 0,
				0, 0, 0, 1,
			},
			Quaternion{0, 0.7071067811865475, 0.7071067811865475, 0},
		},
		{
			Matrix{
				-1, 0, 0, 0,
				0, -1, 0, 0,
				0, 0, -1, 0,
				0, 0, 0, 1,
			},
			Quaternion{0, 0, 0.7071067811865475, 0},
		},
		{
			Matrix{
				1, 1, 0, 0,
				1, -1, 0, 0,
				0, 0, -1, 0,
				0, 0, 0, 1,
			},
			Quaternion{1, 0.5, 0, 0},
		},
	}

	for _, c := range tests {
		if r := QuaternionFromRotationMatrix(c.Rotation); !r.Equals(c.Expected, 2) {
			t.Errorf("QuaternionFromRotationMatrix(\n%v) != %v (got %v)", c.Rotation, c.Expected, r)
		}
	}
}

func TestQuaternion_RotationMatrix(t *testing.T) {
	tests := []struct {
		Rotation Quaternion
		Expected Matrix
	}{
		{
			Quaternion{0, 0, 0, 1},
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
		},
		{
			Quaternion{0, 0.7071, 0, 0.7071},
			Matrix{
				0, 0, -1, 0,
				0, 1, 0, 0,
				1, 0, 0, 0,
				0, 0, 0, 1,
			},
		},
		{
			Quaternion{0, 0, 0.7071, 0.7071},
			Matrix{
				0, 1, 0, 0,
				-1, 0, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
		},
		{
			Quaternion{-0.5, -0.5, 0.5, 0.5},
			Matrix{
				0, 1, 0, 0,
				0, 0, -1, 0,
				-1, 0, 0, 0,
				0, 0, 0, 1,
			},
		},
	}

	for _, c := range tests {
		if r := c.Rotation.RotationMatrix(); !r.Equals(c.Expected, 2) {
			t.Errorf("(%v).RotationMatrix() != \n%v (got \n%v)", c.Rotation, c.Expected, r)
		}
	}
}

func TestQuaternion_Length(t *testing.T) {
	tests := []struct {
		Rotation Quaternion
		Expected float64
	}{
		{Quaternion{1, 0, 0, 0}, 1},
		{Quaternion{0.0000000000001, 0, 0, 0}, 0},
		{Quaternion{math.MaxFloat64, 1, 0, 0}, math.Inf(+1)},
		{Quaternion{1, 2, 3, 4}, math.Sqrt(1*1 + 2*2 + 3*3 + 4*4)},
		{Quaternion{3.1, 4.2, 1.3, 0}, math.Sqrt(3.1*3.1 + 4.2*4.2 + 1.3*1.3)},
	}

	for _, c := range tests {
		if r := c.Rotation.Length(); !NearlyEquals(c.Expected, r, 0.000001) {
			t.Errorf("Quaternion(%v).Length() != %v (got %v)", c.Rotation, c.Expected, r)
		}
	}
}

func TestQuaternion_Normalize(t *testing.T) {
	tests := []struct {
		Rotation Quaternion
		Expected Quaternion
	}{
		{Quaternion{0, 0, 0, 0}, Quaternion{0, 0, 0, 1}},
		{Quaternion{1, 0, 0, 0}, Quaternion{1, 0, 0, 0}},
		{Quaternion{0.0000000000001, 0, 0, 0}, Quaternion{1, 0, 0, 0}},
		{Quaternion{math.MaxFloat32, 1, 0, 0}, Quaternion{1, 0, 0, 0}},
		{Quaternion{1, 2, 3, 4}, Quaternion{1.0 / 5.477, 2.0 / 5.477, 3.0 / 5.477, 4.0 / 5.477}},
		{Quaternion{3.1, 4.2, 1.3, 0}, Quaternion{3.1 / 5.3795, 4.2 / 5.3795, 1.3 / 5.3795, 0}},
	}

	for _, c := range tests {
		if r := c.Rotation.Normalize(); !r.Equals(c.Expected, 4) {
			t.Errorf("Quaternion(%v).Normalize() != %v (got %v)", c.Rotation, c.Expected, r)
		}
	}
}

func TestQuaternion_Conjugate(t *testing.T) {
	tests := []struct {
		Rotation Quaternion
		Expected Quaternion
	}{
		{Quaternion{1, 0, 0, 0}, Quaternion{-1, 0, 0, 0}},
		{Quaternion{0.0000000000001, 0, 0, 0}, Quaternion{0, 0, 0, 0}},
		{Quaternion{math.MaxFloat32, 1, 0, 0}, Quaternion{-math.MaxFloat32, -1, 0, 0}},
		{Quaternion{1, 2, 3, 4}, Quaternion{-1, -2, -3, 4}},
		{Quaternion{3.1, 4.2, 1.3, 0}, Quaternion{-3.1, -4.2, -1.3, 0}},
		{Quaternion{1, 2, 3, 4}, (Quaternion{1, 2, 3, 4}).Conjugate()},
	}

	for _, c := range tests {
		if r := c.Rotation.Conjugate(); !r.Equals(c.Expected, 4) {
			t.Errorf("Quaternion(%v).Conjugate() != %v (got %v)", c.Rotation, c.Expected, r)
		}
	}
}

func TestQuaternion_MulScalar(t *testing.T) {
	tests := []struct {
		Rotation Quaternion
		Scalar   float64
		Expected Quaternion
	}{
		{Quaternion{0, 0, 0, 0}, 1, Quaternion{0, 0, 0, 0}},
		{Quaternion{1, 0, 0, 0}, 2, Quaternion{2, 0, 0, 0}},
		{Quaternion{1, 2, 3, 4}, 3, Quaternion{3, 6, 9, 12}},
	}

	for _, c := range tests {
		if r := c.Rotation.MulScalar(c.Scalar); !r.Equals(c.Expected, 6) {
			t.Errorf("Quaternion(%v).MulScalar(%v) != %v (got %v)", c.Rotation, c.Scalar, c.Expected, r)
		}
	}
}

func TestQuaternion_Add(t *testing.T) {
	tests := []struct {
		A, B     Quaternion
		Expected Quaternion
	}{
		{Quaternion{0, 0, 0, 0}, Quaternion{0, 0, 0, 0}, Quaternion{0, 0, 0, 0}},
		{Quaternion{1, 0, 0, 0}, Quaternion{1, 0, 0, 0}, Quaternion{2, 0, 0, 0}},
		{Quaternion{1, 2, 3, 4}, Quaternion{5, 6, 7, 8}, Quaternion{6, 8, 10, 12}},
	}

	for _, c := range tests {
		if r := c.A.Add(c.B); !r.Equals(c.Expected, 6) {
			t.Errorf("Quaternion(%v).Add(Quaternion(%v)) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func TestQuaternion_Sub(t *testing.T) {
	tests := []struct {
		A, B     Quaternion
		Expected Quaternion
	}{
		{Quaternion{0, 0, 0, 0}, Quaternion{0, 0, 0, 0}, Quaternion{0, 0, 0, 0}},
		{Quaternion{1, 0, 0, 0}, Quaternion{1, 0, 0, 0}, Quaternion{0, 0, 0, 0}},
		{Quaternion{1, 2, 3, 4}, Quaternion{5, 6, 7, 8}, Quaternion{-4, -4, -4, -4}},
	}

	for _, c := range tests {
		if r := c.A.Sub(c.B); !r.Equals(c.Expected, 6) {
			t.Errorf("Quaternion(%v).Sub(Quaternion(%v)) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func TestQuaternion_Mul(t *testing.T) {
	tests := []struct {
		A, B     Quaternion
		Expected Quaternion
	}{
		{Quaternion{0, 0, 0, 0}, Quaternion{0, 0, 0, 0}, Quaternion{0, 0, 0, 0}},
		{Quaternion{0, 0, 0, 1}, Quaternion{0, 0, 0, 1}, Quaternion{0, 0, 0, 1}},
		{Quaternion{1, 2, 3, 4}, Quaternion{5, 6, 7, 8}, Quaternion{24, 48, 48, -6}},
	}

	for _, c := range tests {
		if r := c.A.Mul(c.B); !r.Equals(c.Expected, 6) {
			t.Errorf("Quaternion(%v).MultiplyQuaternions(Quaternion(%v)) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func TestQuaternion_Slerp(t *testing.T) {
	tests := []struct {
		A, B     Quaternion
		Scalar   float64
		Expected Quaternion
	}{
		{Quaternion{0, 0, 0, 0}, Quaternion{0, 0, 0, 0}, 0, Quaternion{0, 0, 0, 0}},
		{Quaternion{1, 0, 0, 0}, Quaternion{1, 0, 0, 0}, 0.5, Quaternion{1, 0, 0, 0}},
		{Quaternion{0, 0, 0, 1}, Quaternion{1, 0, 0, 0}, 0.5, Quaternion{0.7071067811865475, 0, 0, 0.7071067811865475}},
		{Quaternion{-0.5, -0.5, 0.5, 0.5}, Quaternion{-0.080, -0.080, 0, 0.996}, 1, Quaternion{-0.080, -0.080, 0, 0.996}},
		{Quaternion{-0.5, -0.5, 0.5, 0.5}, Quaternion{-0.080, -0.080, 0, 0.996}, 0, Quaternion{-0.5, -0.5, 0.5, 0.5}},
		{Quaternion{-0.5, -0.5, 0.5, 0.5}, Quaternion{-0.080, -0.080, 0, 0.996}, 0.2, Quaternion{-0.44231939784548874, -0.44231939784548874, 0.4237176207195655, 0.6553097459373098}},
		{Quaternion{-0.080, -0.080, 0, 0.996}, Quaternion{-0.5, -0.5, 0.5, 0.5}, 0.8, Quaternion{-0.44231939784548874, -0.44231939784548874, 0.4237176207195655, 0.6553097459373098}},
		{Quaternion{0, 0, 0, 1}, Quaternion{0, 0, 0, -0.9999999}, 0, Quaternion{0, 0, 0, 4.99999e-8}},
	}

	for _, c := range tests {
		if r := c.A.Slerp(c.B, c.Scalar); !r.Equals(c.Expected, 6) {
			t.Errorf("Quaternion(%v).Slerp(%v, %v) != %v (got %v)", c.A, c.B, c.Scalar, c.Expected, r)
		}
	}
}

func TestQuaternion_String(t *testing.T) {
	tests := []struct {
		Q        Quaternion
		Expected string
	}{
		{Quaternion{0, 0, 0, 0}, " 0.00  0.00  0.00  0.00"},
		{Quaternion{1, 2, 3, 4}, " 1.00  2.00  3.00  4.00"},
		{Quaternion{0.12345, 123456, 0.456, math.Inf(1)}, " 0.12 123456.00  0.46  +Inf"},
	}

	for _, c := range tests {
		if r := c.Q.String(); r != c.Expected {
			t.Errorf("Quaternion(%v).String() != %v (got %v)", c.Q, c.Expected, r)
		}
	}
}
