package math

import (
	"testing"
)

func TestColor_Equals(t *testing.T) {
	tests := []struct {
		A, B     Color
		Expected bool
	}{
		{
			Color{0, 0, 0},
			Color{0, 0, 0},
			true,
		},
		{
			Color{1, 1, 1},
			Color{1, 1, 1},
			true,
		},
		{
			Color{1, 1, 1},
			Color{0, 0, 0},
			false,
		},
		{
			Color{0.12345, 0.34568, 0.56789},
			Color{0.12345, 0.34568, 0.56789},
			true,
		},
		{
			Color{0.0000000000001, 0.0000000000001, 0.0000000000001},
			Color{0, 0, 0},
			true,
		},
		{
			Color{0, 0.00000001, 0},
			Color{0, 0, 0},
			false,
		},
	}

	for _, c := range tests {
		if r := c.A.Equals(c.B, 6); r != c.Expected {
			t.Errorf("Color(%v).Equals(Color(%v), 6) != %v (got %v)", c.A, c.B, c.Expected, r)
		}
	}
}

func TestColorFromHex(t *testing.T) {
	tests := []struct {
		Hex      int
		Expected Color
	}{
		{
			0x000000,
			Color{0, 0, 0},
		},
		{
			0xffffff,
			Color{1, 1, 1},
		},
		{
			0x00ff00,
			Color{0, 1, 0},
		},
		{
			0x123456,
			Color{0x12 / 255.0, 0x34 / 255.0, 0x56 / 255.0},
		},
	}

	for _, c := range tests {
		if r := ColorFromHex(c.Hex); !r.Equals(c.Expected, 6) {
			t.Errorf("ColorFromHex(%#x) != %v (got %v)", c.Hex, c.Expected, r)
		}
	}
}

func TestColorFromRGB(t *testing.T) {
	tests := []struct {
		R, G, B  float64
		Expected Color
	}{
		{
			0, 0, 0,
			Color{0, 0, 0},
		},
		{
			0, 1, 0,
			Color{0, 1, 0},
		},
		{
			0.1, 0.2, 0.3,
			Color{0.1, 0.2, 0.3},
		},
	}

	for _, c := range tests {
		if r := ColorFromRGB(c.R, c.G, c.B); !r.Equals(c.Expected, 6) {
			t.Errorf("ColorFromRGB(%1.2f, %1.2f, %1.2f) != %v (got %v)", c.R, c.G, c.B, c.Expected, r)
		}
	}
}

func TestColorFromHSL(t *testing.T) {
	tests := []struct {
		H, S, L  float64
		Expected Color
	}{
		{
			0, 0, 0,
			Color{0, 0, 0},
		},
		{
			0, 0, 1,
			Color{1, 1, 1},
		},
		{
			(0.0 + 180.0) / 360.0, 1, 0.5,
			Color{1, 0, 0},
		},
		{
			(60.0 + 180.0) / 360.0, 1, 0.5,
			Color{1, 1, 0},
		},
		{
			(115.0 + 180.0) / 360.0, 0.55, 0.8,
			Color{180.0 / 255.0, 232.0 / 255.0, 175.0 / 255.0},
		},
		{
			-0.5, 0.5, 0.5,
			Color{191.0 / 255.0, 63.0 / 255.0, 63.0 / 255.0},
		},
		{
			0.5, 0.5, 0.5,
			Color{191.0 / 255.0, 63.0 / 255.0, 63.0 / 255.0},
		},
	}

	for _, c := range tests {
		if r := ColorFromHSL(c.H, c.S, c.L); !r.Equals(c.Expected, 2) {
			t.Errorf("ColorFromHSL(%1.2f, %1.2f, %1.2f) != %v (got %v)", c.H, c.S, c.L, c.Expected, r)
		}
	}
}

func TestColorFromVector(t *testing.T) {
	tests := []struct {
		V        Vector
		Expected Color
	}{
		{
			Vector{0, 0, 0},
			Color{0, 0, 0},
		},
		{
			Vector{0, 1, 0},
			Color{0, 1, 0},
		},
		{
			Vector{0.1, 0.2, 0.3},
			Color{0.1, 0.2, 0.3},
		},
	}

	for _, c := range tests {
		if r := ColorFromVector(c.V); !r.Equals(c.Expected, 6) {
			t.Errorf("ColorFromVector(%v) != %v (got %v)", c.V, c.Expected, r)
		}
	}
}

func TestColor_Normalize(t *testing.T) {
	tests := []struct {
		C        Color
		Expected Color
	}{
		{
			Color{0, 0, 0},
			Color{0, 0, 0},
		},
		{
			Color{0, 255, 0},
			Color{0, 1, 0},
		},
		{
			Color{1, 2, 3},
			Color{1.0 / 255.0, 2.0 / 255.0, 3.0 / 255.0},
		},
	}

	for _, c := range tests {
		if r := c.C.Normalize(); !r.Equals(c.Expected, 6) {
			t.Errorf("Color(%v).Normalize() != %v (got %v)", c.C, c.Expected, r)
		}
	}

}
