package math

import (
	"math"
)

type Color struct {
	R, G, B float64
}

func ColorFromHex(hex int) Color {
	return Color{
		R: float64(hex>>16&255) / 255,
		G: float64(hex>>8&255) / 255,
		B: float64(hex&255) / 255,
	}
}

func ColorFromRGB(r, g, b float64) Color {
	return Color{r, g, b}
}

func ColorFromHSL(h, s, l float64) Color {
	if s == 0 {
		return Color{l, l, l}
	}

	var p float64
	if l <= 0.5 {
		p = l * (1.0 + s)
	} else {
		p = l + s - (l * s)
	}

	var q float64 = (2.0 * l) - p

	return Color{
		R: hue2rgb(p, q, h+1.0/3.0),
		G: hue2rgb(p, q, h),
		B: hue2rgb(p, q, h-1.0/3.0),
	}
}

func hue2rgb(p, q, t float64) float64 {
	switch {
	case t < 0:
		t += 1
	case t > 1.0:
		t -= 1.0
	}

	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6.0*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*6.0*(2.0/3.0-t)
	}

	return p
}

func ColorFromVector(v Vector) Color {
	return Color{v[0], v[1], v[2]}
}

func (c Color) Equals(e Color, precision int) bool {
	p := math.Pow(10, float64(-precision))

	return (NearlyEquals(c.R, e.R, p) &&
		NearlyEquals(c.G, e.G, p) &&
		NearlyEquals(c.B, e.B, p))
}

func (c Color) Normalize() Color {
	return Color{
		R: c.R / 255.0,
		G: c.G / 255.0,
		B: c.B / 255.0,
	}
}
