package math

type Plane struct {
	normal   Vector
	distance float64
}

func (p Plane) Normalize() Plane {
	magnitude := p.normal.Length()

	return Plane{
		normal:   p.normal.MulScalar(1.0 / magnitude),
		distance: p.distance / magnitude,
	}
}

type Frustum [6]Plane

func FrustumFromMatrix(m Matrix) Frustum {
	f := Frustum{
		Plane{Vector{m[3] - m[0], m[7] - m[4], m[11] - m[8]}, m[15] - m[12]},
		Plane{Vector{m[3] + m[0], m[7] + m[4], m[11] + m[8]}, m[15] + m[12]},
		Plane{Vector{m[3] + m[1], m[7] + m[5], m[11] + m[9]}, m[15] + m[13]},
		Plane{Vector{m[3] - m[1], m[7] - m[5], m[11] - m[9]}, m[15] - m[13]},
		Plane{Vector{m[3] - m[2], m[7] - m[6], m[11] - m[10]}, m[15] - m[14]},
		Plane{Vector{m[3] + m[2], m[7] + m[6], m[11] + m[10]}, m[15] + m[14]},
	}

	for i, p := range f {
		f[i] = p.Normalize()
	}

	return f
}

func (f Frustum) ContainsPoint(point Vector) bool {
	for _, p := range f {
		if point.Dot(p.normal)+p.distance <= 0 {
			return false
		}
	}

	return true
}

func (f Frustum) IntersectsSphere(center Vector, radius float64) bool {
	for _, p := range f {
		if center.Dot(p.normal)+p.distance <= -radius {
			return false
		}
	}

	return true
}
