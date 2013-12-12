package engine

import (
	"fmt"

	"github.com/der-antikeks/gisp/math"
)

type Scene struct {
	Object

	objects []Renderable
	lights  []Object
	/*
		fog     struct {
			fogNear  float64
			fogFar   float64
			fogColor math.Color
		}
	*/

	// 3d
	position math.Vector
	up       math.Vector
	rotation math.Quaternion
	scale    math.Vector

	matrix                 math.Matrix
	matrixNeedsUpdate      bool
	matrixWorld            math.Matrix
	matrixWorldNeedsUpdate bool

	// relationship
	parent   Object
	children []Object
}

func NewScene() *Scene {
	return &Scene{
		up:    math.Vector{0, 1, 0},
		scale: math.Vector{1, 1, 1},

		matrixNeedsUpdate:      true,
		matrixWorldNeedsUpdate: true,
	}
}

func (s *Scene) AddObject(o Object) {
	switch ot := o.(type) {
	case Renderable:
		var found bool
		for _, c := range s.objects {
			if ot == c {
				found = true
				break
			}
		}

		if !found {
			s.objects = append(s.objects, ot)
		}

	//case Light:
	//	s.lights = append(s.lights, r)
	case *Group:
	case *Scene:
	case Camera:
	default:
		fmt.Printf("unknown object type: %T\n", o)
	}

	for _, c := range o.Children() {
		s.AddObject(c)
	}
}

func (s *Scene) RemoveObject(o Object) {
	switch ot := o.(type) {
	case Renderable:
		position := -1
		for i, c := range s.objects {
			if ot == c {
				position = i
				break
			}
		}

		if position != -1 {
			copy(s.objects[position:], s.objects[position+1:])
			s.objects[len(s.objects)-1] = nil
			s.objects = s.objects[:len(s.objects)-1]
		}

	//case Light:
	case *Group:
	case *Scene:
	case Camera:
	default:
		fmt.Printf("unknown object type: %T\n", o)
	}

	for _, c := range o.Children() {
		s.RemoveObject(c)
	}
}

func (s *Scene) VisibleObjects(f math.Frustum) (opaque, transparent []Renderable) {
	opaque = make([]Renderable, len(s.objects))
	transparent = make([]Renderable, len(s.objects))
	var cntOp, cntTr int

	for _, o := range s.objects {
		c, r := o.Geometry().Boundary().Sphere()

		// transform sphere with modelmatrix
		if f.IntersectsSphere(o.MatrixWorld().Transform(c), r*o.MatrixWorld().MaxScaleOnAxis()) {
			if o.Material().Opaque() {
				opaque[cntOp] = o
				cntOp++
			} else {
				transparent[cntTr] = o
				cntTr++
			}
		}
	}

	return opaque[:cntOp], transparent[:cntTr]
}

func (s *Scene) Dispose() {
	for _, o := range s.objects {
		o.Dispose()
	}
}

func (o *Scene) SetPosition(p math.Vector) {
	o.position = p
	o.matrixNeedsUpdate = true
}

func (o *Scene) Position() math.Vector {
	return o.position
}

func (o *Scene) SetUp(u math.Vector) {
	o.up = u.Normalize()
	o.matrixNeedsUpdate = true
}

func (o *Scene) Up() math.Vector {
	return o.up
}

func (o *Scene) LookAt(v math.Vector) {
	o.SetRotation(math.QuaternionFromRotationMatrix(math.LookAt(o.position, v, o.up)))
}

func (o *Scene) SetRotation(r math.Quaternion) {
	o.rotation = r
	o.matrixNeedsUpdate = true
}

func (o *Scene) Rotation() math.Quaternion {
	return o.rotation
}

func (o *Scene) SetScale(s math.Vector) {
	o.scale = s
	o.matrixNeedsUpdate = true
}

func (o *Scene) Scale() math.Vector {
	return o.scale
}

func (o *Scene) Matrix() math.Matrix {
	if o.matrixNeedsUpdate {
		o.matrix = math.ComposeMatrix(o.position, o.rotation, o.scale)

		o.matrixWorldNeedsUpdate = true
		o.matrixNeedsUpdate = false
	}

	return o.matrix
}

func (o *Scene) UpdateMatrixWorld(force bool) {
	m := o.Matrix()

	if o.matrixWorldNeedsUpdate || force {
		if p := o.Parent(); p == nil {
			o.matrixWorld = m
		} else {
			o.matrixWorld = p.MatrixWorld().Mul(m)
		}

		o.matrixWorldNeedsUpdate = false
		force = true
	}

	for _, c := range o.Children() {
		c.UpdateMatrixWorld(force)
	}
}

func (o *Scene) MatrixWorld() math.Matrix {
	return o.matrixWorld
}

func (o *Scene) AddChild(cs ...Object) {
	for _, c := range cs {
		if o == c {
			continue
		}

		if p := c.Parent(); p != nil {
			p.RemoveChild(c)
		}
		c.SetParent(o)

		o.children = append(o.children, c)
		o.AddObject(c)
	}
}

func (o *Scene) RemoveChild(r Object) {
	r.SetParent(nil)

	position := -1
	for i, c := range o.children {
		if r == c {
			position = i
			break
		}
	}

	if position != -1 {
		copy(o.children[position:], o.children[position+1:])
		o.children[len(o.children)-1] = nil
		o.children = o.children[:len(o.children)-1]
	}

	o.RemoveObject(r)
}

func (o *Scene) Children() []Object {
	return o.children
}

func (o *Scene) SetParent(p Object) {
	o.parent = p
}

func (o *Scene) Parent() Object {
	if o.parent == nil {
		return nil
	}

	return o.parent
}
