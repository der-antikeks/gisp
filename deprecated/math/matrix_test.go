package math

import (
	"math"
	"testing"
)

func TestMatrix_Equals(t *testing.T) {
	tests := []struct {
		A, B     Matrix
		Expected bool
	}{
		{
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			}, Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			}, true,
		}, {
			Matrix{
				1, 2, 3, 0,
				4, 5, 6, 0,
				7, 8, 9, 0,
				0, 0, 0, 1,
			}, Matrix{
				1, 2, 3, 0,
				4, 5, 6, 0,
				7, 8, 9, 0,
				0, 0, 0, 1,
			}, true,
		}, {
			Matrix{
				1, 2, 3, -1,
				4, 5, 6, -2,
				7, 8, 9, -3,
				-4, -5, -6, 1,
			}, Matrix{
				1, 2, 3, 0,
				4, 5, 6, 0,
				7, 8, 9, 0,
				0, 0, 0, 1,
			}, false,
		},
	}

	for _, c := range tests {
		if r := c.A.Equals(c.B, 6); r != c.Expected {
			t.Errorf("Matrix(\n%v).Equals(Matrix(\n%v), 6) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func BenchmarkMatrix_Equals(b *testing.B) {
	b.StopTimer()
	ma := Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	mb := Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ma.Equals(mb, 6)
	}
}

func TestIdentity(t *testing.T) {
	tests := []struct {
		A, B     Matrix
		Expected bool
	}{
		{Identity(), Matrix{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		}, true},
	}

	for _, c := range tests {
		if r := c.A.Equals(c.B, 6); r != c.Expected {
			t.Errorf("Matrix(\n%v).Equals(Matrix(\n%v), 6) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func TestMatrix_Mul(t *testing.T) {
	tests := []struct {
		A, B     Matrix
		Expected Matrix
	}{
		{Identity(), Identity(), Identity()},
		{
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				4, 5, 6, 1,
			}, Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				1, 2, 3, 1,
			}, Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				5, 7, 9, 1,
			},
		}, {
			Matrix{
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, -1, 0, 0,
				0, 0, 0, 1,
			}, Matrix{
				0, 0, -1, 0,
				0, 1, 0, 0,
				1, 0, 0, 0,
				0, 0, 0, 1,
			}, Matrix{
				0, 1, 0, 0,
				0, 0, 1, 0,
				1, 0, 0, 0,
				0, 0, 0, 1,
			},
		},
	}

	for _, c := range tests {
		if r := c.A.Mul(c.B); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).Mul(Matrix(\n%v)) != \n%v (got \n%v)", c.A, c.B, c.Expected, r)
		}
	}
}

func BenchmarkMatrix_Mul(b *testing.B) {
	b.StopTimer()
	ma := Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		4, 5, 6, 1,
	}
	mb := Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		1, 2, 3, 1,
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ma.Mul(mb)
	}
}

func TestMatrix_Float64(t *testing.T) {
	tests := []struct {
		A        Matrix
		Expected [16]float64
	}{
		{Identity(), [16]float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		}}, {
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				4, 5, 6, 1,
			}, [16]float64{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				4, 5, 6, 1,
			},
		},
	}

	for _, c := range tests {
		if r := c.A.Float64(); r != c.Expected {
			t.Errorf("Matrix(\n%v).Float64() != \n%v (got \n%v)", c.A, c.Expected, r)
		}
	}
}

func TestMatrix_Transpose(t *testing.T) {
	tests := []struct {
		A        Matrix
		Expected Matrix
	}{
		{Identity(), Identity()},
		{
			Matrix{
				1, 2, 3, 4,
				5, 6, 7, 8,
				9, 10, 11, 12,
				13, 14, 15, 16,
			}, Matrix{
				1, 5, 9, 13,
				2, 6, 10, 14,
				3, 7, 11, 15,
				4, 8, 12, 16,
			},
		},
	}

	for _, c := range tests {
		if r := c.A.Transpose(); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).Transpose() != \n%v (got \n%v)", c.A, c.Expected, r)
		}
	}
}

func BenchmarkMatrix_Transpose(b *testing.B) {
	b.StopTimer()
	ma := Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		4, 5, 6, 1,
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ma.Transpose()
	}
}

func TestMatrix_Scale(t *testing.T) {
	tests := []struct {
		A        Matrix
		Scale    Vector
		Expected Matrix
	}{
		{Identity(), Vector{1, 1, 1}, Identity()},
		{
			Matrix{
				1, 2, 3, 4,
				5, 1, 7, 8,
				9, 10, 1, 12,
				13, 14, 15, 1,
			}, Vector{2, 2, 2},
			Matrix{
				2, 4, 6, 8,
				10, 2, 14, 16,
				18, 20, 2, 24,
				13, 14, 15, 1,
			},
		},
	}

	for _, c := range tests {
		if r := c.A.Scale(c.Scale); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).Scale(%v) != \n%v (got \n%v)", c.A, c.Scale, c.Expected, r)
		}
	}
}

func BenchmarkMatrix_Scale(b *testing.B) {
	b.StopTimer()
	ma := Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		4, 5, 6, 1,
	}
	vb := Vector{2, 2, 2}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ma.Scale(vb)
	}
}

func TestMatrix_Rotate(t *testing.T) {
	tests := []struct {
		A        Matrix
		Angle    float64
		Axis     Vector
		Expected Matrix
	}{
		{
			Identity(),
			math.Pi / 2,
			Vector{1, 0, 0},
			Matrix{
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, -1, 0, 0,
				0, 0, 0, 1,
			},
		}, {
			Identity(),
			math.Pi / 2,
			Vector{0, 1, 0},
			Matrix{
				0, 0, -1, 0,
				0, 1, 0, 0,
				1, 0, 0, 0,
				0, 0, 0, 1,
			},
		}, {
			Identity(),
			math.Pi / 2,
			Vector{0, 0, 1},
			Matrix{
				0, 1, 0, 0,
				-1, 0, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
		},
	}

	for _, c := range tests {
		if r := c.A.Rotate(c.Angle, c.Axis); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).Rotate(%v, Vector(%v) != \n%v (got \n%v)", c.A, c.Angle, c.Axis, c.Expected, r)
		}
	}
}

func BenchmarkMatrix_Rotate(b *testing.B) {
	b.StopTimer()
	ma := Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		4, 5, 6, 1,
	}
	angle := math.Pi / 2
	axis := Vector{1, 0, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ma.Rotate(angle, axis)
	}
}

func TestMatrix_Translate(t *testing.T) {
	tests := []struct {
		A           Matrix
		Translation Vector
		Expected    Matrix
	}{
		{Identity(), Vector{1, 1, 1}, Matrix{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			1, 1, 1, 1,
		}},
		{
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				13, 14, 15, 1,
			}, Vector{1, 2, 3},
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				14, 16, 18, 1,
			},
		},
	}

	for _, c := range tests {
		if r := c.A.Translate(c.Translation); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).Translate(%v) != \n%v (got \n%v)", c.A, c.Translation, c.Expected, r)
		}
	}
}

func BenchmarkMatrix_Translate(b *testing.B) {
	b.StopTimer()
	ma := Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		4, 5, 6, 1,
	}
	vb := Vector{2, 2, 2}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ma.Translate(vb)
	}
}

func TestMatrix_Transform(t *testing.T) {
	tests := []struct {
		A        Matrix
		P        Vector
		Expected Vector
	}{
		{Identity(), Vector{1, 1, 1}, Vector{1, 1, 1}},
		{
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				13, 14, 15, 1,
			},
			Vector{1, 2, 3},
			Vector{1, 2, 3},
		}, {
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				13, 14, 15, 1,
			},
			Vector{1, 2, 3, 1},
			Vector{14, 16, 18, 1},
		}, {
			Matrix{
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, -1, 0, 0,
				1, 2, 3, 1,
			},
			Vector{1, 2, 3},
			Vector{1, -3, 2},
		}, {
			Matrix{
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, -1, 0, 0,
				0, 0, -10, 1,
			},
			Vector{1, 2, 3, 1},
			Vector{1, -3, -8, 1},
		}, {
			Identity().Rotate(math.Pi/2, Vector{1, 0, 0}).Translate(Identity().Rotate(math.Pi/2, Vector{1, 0, 0}).Transform(Vector{0, 0, 10})),
			Vector{1, 2, 3, 1},
			Vector{1, -3, -8, 1},
		},
	}

	for _, c := range tests {
		if r := c.A.Transform(c.P); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).Transform(%v) != \n%v (got \n%v)", c.A, c.P, c.Expected, r)
		}
	}
}

func BenchmarkMatrix_Transform(b *testing.B) {
	b.StopTimer()
	ma := Matrix{
		1, 0, 0, 0,
		0, 0, 1, 0,
		0, -1, 0, 0,
		4, 5, 6, 1,
	}
	vb := Vector{2, 2, 2}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ma.Transform(vb)
	}
}

/*
func TestMatrix_LookAt(t *testing.T) {
	tests := []struct {
		A        Matrix
		Target   Vector
		Expected Matrix
	}{
		{
			Identity().Rotate(math.Pi/2, Vector{1, 0, 0}),
			Vector{0, 0, 1},
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
		}, {
			Matrix{
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, -1, 0, 0,
				0, 0, 0, 1,
			},
			Vector{1, 0, 0},
			Matrix{
				0, 0, -1, 0,
				0, 1, 0, 0,
				1, 0, 0, 0,
				0, 0, 0, 1,
			},
		},
	}

	for _, c := range tests {
		if r := LookAt(c.Target, Vector{0, 1, 0}); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).LookAt(%v, Vector{0, 1, 0}) != \n%v (got \n%v)", c.A, c.Target, c.Expected, r)
		}
	}
}

func BenchmarkMatrix_LookAt(b *testing.B) {
	b.StopTimer()
	ma := Matrix{
		1, 0, 0, 0,
		0, 0, 1, 0,
		0, -1, 0, 0,
		4, 5, 6, 1,
	}
	target := Vector{2, 2, 2}
	up := Vector{0, 1, 0}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ma.LookAt(target, up)
	}
}
*/
func TestMatrix_ExtractScale(t *testing.T) {
	tests := []struct {
		A        Matrix
		Expected Vector
	}{
		{
			Identity(),
			Vector{1, 1, 1},
		}, {
			Identity().Scale(Vector{1, 2, 3}),
			Vector{1, 2, 3},
		}, {
			Identity().Scale(Vector{2, 3, 4}).Rotate(math.Pi/2, Vector{1, 0, 0}).Translate(Vector{10, 12, -5}),
			Vector{2, 4, 3},
		},
	}

	for _, c := range tests {
		if r := c.A.ExtractScale(); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).ExtractScale() != \n%v (got \n%v)", c.A, c.Expected, r)
		}
	}
}

func BenchmarkMatrix_ExtractScale(b *testing.B) {
	b.StopTimer()
	ma := Matrix{
		1, 0, 0, 0,
		0, 0, 1, 0,
		0, -1, 0, 0,
		4, 5, 6, 1,
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ma.ExtractScale()
	}
}

/*
func TestMatrix_ExtractRotate(t *testing.T) {
	tests := []struct {
		A        Matrix
		Expected Vector
	}{
		{
			Identity(),
			Vector{0, 0, 0},
		}, {
			Identity().Rotate(math.Pi / 2, Vector{1, 0, 0}),
			Vector{math.Pi / 2, 0, 0},
		}, {
			Identity().Scale(Vector{2, 3, 4}).Rotate(math.Pi / 2, Vector{0, 1, 0}).Translate(Vector{10, 12, -5}),
			Vector{0, math.Pi / 2, 0},
		}, {
			Identity().Rotate(math.Pi * 1, Vector{0, 1, 0}),
			Vector{0, math.Pi * 1, 0},
		}, {
			Identity().Rotate(math.Pi * 1.5, Vector{1, 0, 0}),
			Vector{math.Pi * 1.5, 0, 0},
		},
	}

	for _, c := range tests {
		if r := c.A.ExtractRotate(); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).ExtractRotate() != \n%v (got \n%v)", c.A, c.Expected, r)
		}
	}
}
*/

func TestMatrix_Determinant(t *testing.T) {
	tests := []struct {
		A        Matrix
		Expected float64
	}{
		{
			Identity(),
			1,
		}, {
			Matrix{
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, -1, 0, 0,
				4, 5, 6, 1,
			}, 1,
		}, {
			Matrix{
				1, 0, 0, 1,
				0, 1, 0, 0,
				0, 0, 1, 0,
				1, 0, 0, 1,
			}, 0,
		}, {
			Matrix{
				1, 2, 3, 4,
				5, 1, 7, 8,
				9, 10, 1, 12,
				13, 14, 15, 1,
			}, -4350,
		},
	}

	for _, c := range tests {
		if r := c.A.Determinant(); !NearlyEquals(r, c.Expected, 0.000001) {
			t.Errorf("Matrix(\n%v).Determinant() != %v (got %v)", c.A, c.Expected, r)
		}
	}
}

func TestMatrix_MulScalar(t *testing.T) {
	tests := []struct {
		A        Matrix
		S        float64
		Expected Matrix
	}{
		{
			Identity(), 1, Identity(),
		}, {
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				10, 20, 30, 1,
			}, 10,
			Matrix{
				10, 0, 0, 0,
				0, 10, 0, 0,
				0, 0, 10, 0,
				100, 200, 300, 10,
			},
		}, {
			Matrix{
				1, 2, 3, 4,
				5, 1, 7, 8,
				9, 10, 1, 12,
				13, 14, 15, 1,
			}, 2,
			Matrix{
				2, 4, 6, 8,
				10, 2, 14, 16,
				18, 20, 2, 24,
				26, 28, 30, 2,
			},
		},
	}

	for _, c := range tests {
		if r := c.A.MulScalar(c.S); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).MulScalar(%v) != \n%v (got \n%v)", c.A, c.S, c.Expected, r)
		}
	}
}

// TODO: det == 0
func TestMatrix_Inverse(t *testing.T) {
	tests := []struct {
		A        Matrix
		Expected Matrix
	}{
		{
			Identity(),
			Identity(),
		}, {
			Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				10, 20, 30, 1,
			}, Matrix{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				-10, -20, -30, 1,
			},
		}, {
			Matrix{
				0, 1, 0, 0,
				-1, 0, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			}, Matrix{
				0, -1, 0, 0,
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
		}, {
			Matrix{
				1, 0, 0, 0,
				0, 2, 0, 0,
				0, 0, 3, 0,
				0, 0, 0, 1,
			}, Matrix{
				1, 0, 0, 0,
				0, 1.0 / 2.0, 0, 0,
				0, 0, 1.0 / 3.0, 0,
				0, 0, 0, 1,
			},
		}, {
			Matrix{
				2, 0, 0, 0,
				0, 0, 1, 0,
				0, -1, 1, 0,
				4, 5, 6, 1,
			}, Matrix{
				0.5, 0, 0, 0,
				0, 1, -1, 0,
				0, 1, 0, 0,
				-2, -11, 5, 1,
			},
		},
	}

	for _, c := range tests {
		if r := c.A.Inverse(); !r.Equals(c.Expected, 6) {
			t.Errorf("Matrix(\n%v).Inverse() != \n%v (got \n%v)", c.A, c.Expected, r)
		}
	}
}

func BenchmarkMatrix_Inverse(b *testing.B) {
	b.StopTimer()
	ma := Matrix{
		1, 0, 0, 0,
		0, 0, 1, 0,
		0, -1, 0, 0,
		4, 5, 6, 1,
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ma.Inverse()
	}
}

func TestMatrix_ComposeMatrix(t *testing.T) {
	tests := []struct {
		Position Vector
		Rotation Quaternion
		Scale    Vector
		Expected Matrix
	}{
		{
			Vector{0, 0, 0},
			Quaternion{0, 0, 0, 1},
			Vector{1, 1, 1},
			Identity(),
		},
		{
			Vector{0, 0, 0},
			QuaternionFromEuler(Vector{0, 90 * DEG2RAD, 0}, DefaultOrder),
			Vector{1, 1, 1},
			Matrix{
				0, 0, -1, 0,
				0, 1, 0, 0,
				1, 0, 0, 0,
				0, 0, 0, 1,
			},
		},
		{
			Vector{10, 0, 0},
			QuaternionFromEuler(Vector{0, 90 * DEG2RAD, 0}, DefaultOrder),
			Vector{1, 1, 1},
			Matrix{
				0, 0, -1, 0,
				0, 1, 0, 0,
				1, 0, 0, 0,
				10, 0, 0, 1,
			},
		},
		{
			Vector{10, 0, 0},
			QuaternionFromEuler(Vector{0, 90 * DEG2RAD, 0}, DefaultOrder),
			Vector{2, 2, 2},
			Matrix{
				0, 0, -2, 0,
				0, 2, 0, 0,
				2, 0, 0, 0,
				10, 0, 0, 1,
			},
		},
		{
			Vector{0, 10, 2},
			QuaternionFromEuler(Vector{0, 0, 53 * DEG2RAD}, DefaultOrder),
			Vector{1, 1, 1},
			Matrix{
				0.60032, 0.79744, 0, 0,
				-0.79744, 0.60032, 0, 0,
				0, 0, 0.99815, 0,
				0, 10, 2, 1,
			},
		},
	}

	for _, c := range tests {
		if r := ComposeMatrix(c.Position, c.Rotation, c.Scale); !r.Equals(c.Expected, 2) {
			t.Errorf("ComposeMatrix(\nP: %v, \nR: %v, \nS: %v) != \n%v (got \n%v)", c.Position, c.Rotation, c.Scale, c.Expected, r)
		}
	}
}

// TODO:
func TestMatrix_String(t *testing.T)          {}
func TestMatrix_Float32(t *testing.T)         {}
func TestMatrix_Matrix3Float32(t *testing.T)  {}
func TestMatrix_MaxScaleOnAxis(t *testing.T)  {}
func TestMatrix_ExtractPosition(t *testing.T) {}
func TestMatrix_SetPosition(t *testing.T)     {}
func TestLookAt(t *testing.T)                 {}
func TestMatrix_Normal(t *testing.T)          {}
func TestNewPerspectiveMatrix(t *testing.T)   {}
func TestNewFrustumMatrix(t *testing.T)       {}
func TestNewOrthoMatrix(t *testing.T)         {}
