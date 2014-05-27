package engine

import (
	"github.com/der-antikeks/gisp/math"
)

type Object interface {
	// 3d
	SetPosition(math.Vector)
	Position() math.Vector

	SetUp(math.Vector)
	Up() math.Vector

	LookAt(math.Vector)
	SetRotation(math.Quaternion)
	Rotation() math.Quaternion

	SetScale(math.Vector)
	Scale() math.Vector

	Matrix() math.Matrix
	UpdateMatrixWorld(bool)
	MatrixWorld() math.Matrix

	// relationship
	AddChild(...Object)
	RemoveChild(Object)
	Children() []Object

	SetParent(Object)
	Parent() Object
}

type Group struct {
	Object

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

func NewGroup() *Group {
	return &Group{
		up:    math.Vector{0, 1, 0},
		scale: math.Vector{1, 1, 1},

		matrixNeedsUpdate:      true,
		matrixWorldNeedsUpdate: true,
	}
}

func (o *Group) SetPosition(p math.Vector) {
	o.position = p
	o.matrixNeedsUpdate = true
}

func (o *Group) Position() math.Vector {
	return o.position
}

func (o *Group) SetUp(u math.Vector) {
	o.up = u.Normalize()
	o.matrixNeedsUpdate = true
}

func (o *Group) Up() math.Vector {
	return o.up
}

func (o *Group) LookAt(v math.Vector) {
	o.SetRotation(math.QuaternionFromRotationMatrix(math.LookAt(o.position, v, o.up)))
}

func (o *Group) SetRotation(r math.Quaternion) {
	o.rotation = r
	o.matrixNeedsUpdate = true
}

func (o *Group) Rotation() math.Quaternion {
	return o.rotation
}

func (o *Group) SetScale(s math.Vector) {
	o.scale = s
	o.matrixNeedsUpdate = true
}

func (o *Group) Scale() math.Vector {
	return o.scale
}

func (o *Group) Matrix() math.Matrix {
	if o.matrixNeedsUpdate {
		o.matrix = math.ComposeMatrix(o.position, o.rotation, o.scale)

		o.matrixWorldNeedsUpdate = true
		o.matrixNeedsUpdate = false
	}

	return o.matrix
}

func (o *Group) UpdateMatrixWorld(force bool) {
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

func (o *Group) MatrixWorld() math.Matrix {
	return o.matrixWorld
}

func (o *Group) AddChild(cs ...Object) {
	for _, c := range cs {
		if o == c {
			continue
		}

		if p := c.Parent(); p != nil {
			p.RemoveChild(c)
		}
		c.SetParent(o)

		o.children = append(o.children, c)

		// backward search for root
		var root, parent Object
		for parent = c; parent != nil; parent = parent.Parent() {
			root = parent
		}

		if scene, ok := root.(*Scene); ok {
			scene.AddObject(c)
		}
	}
}

func (o *Group) RemoveChild(r Object) {
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

		// backward search for root
		var root, parent Object
		for parent = r; parent != nil; parent = parent.Parent() {
			root = parent
		}

		if scene, ok := root.(*Scene); ok {
			scene.RemoveObject(r)
		}
	}
}

func (o *Group) Children() []Object {
	return o.children
}

func (o *Group) SetParent(p Object) {
	o.parent = p
}

func (o *Group) Parent() Object {
	if o.parent == nil {
		return nil
	}

	return o.parent
}
