package engine

import (
	"github.com/der-antikeks/gisp/math"
)

type Renderable interface {
	Object

	Geometry() *Geometry
	SetGeometry(g *Geometry)
	Material() *Material
	SetMaterial(m *Material)
	Dispose()
}

type Mesh struct {
	Renderable

	geometry *Geometry
	material *Material

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

func NewMesh(geo *Geometry, mat *Material) *Mesh {
	return &Mesh{
		geometry: geo,
		material: mat,

		up:    math.Vector{0, 1, 0},
		scale: math.Vector{1, 1, 1},

		matrixNeedsUpdate:      true,
		matrixWorldNeedsUpdate: true,
	}
}

func (m *Mesh) Dispose() {
	m.material.Dispose()
	m.geometry.Dispose()
}

func (m *Mesh) Geometry() *Geometry {
	return m.geometry
}

func (m *Mesh) SetGeometry(g *Geometry) {
	m.geometry = g
}

func (m *Mesh) Material() *Material {
	return m.material
}

func (m *Mesh) SetMaterial(mat *Material) {
	m.material = mat
}

func (o *Mesh) SetPosition(p math.Vector) {
	o.position = p
	o.matrixNeedsUpdate = true
}

func (o *Mesh) Position() math.Vector {
	return o.position
}

func (o *Mesh) SetUp(u math.Vector) {
	o.up = u.Normalize()
	o.matrixNeedsUpdate = true
}

func (o *Mesh) Up() math.Vector {
	return o.up
}

func (o *Mesh) SetRotation(r math.Quaternion) {
	o.rotation = r
	o.matrixNeedsUpdate = true
}

func (o *Mesh) Rotation() math.Quaternion {
	return o.rotation
}

func (o *Mesh) SetScale(s math.Vector) {
	o.scale = s
	o.matrixNeedsUpdate = true
}

func (o *Mesh) Scale() math.Vector {
	return o.scale
}

func (o *Mesh) Matrix() math.Matrix {
	if o.matrixNeedsUpdate {
		o.matrix = math.ComposeMatrix(o.position, o.rotation, o.scale)

		o.matrixWorldNeedsUpdate = true
		o.matrixNeedsUpdate = false
	}

	return o.matrix
}

func (o *Mesh) UpdateMatrixWorld(force bool) {
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

func (o *Mesh) MatrixWorld() math.Matrix {
	return o.matrixWorld
}

func (o *Mesh) AddChild(cs ...Object) {
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

func (o *Mesh) RemoveChild(r Object) {
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

func (o *Mesh) Children() []Object {
	return o.children
}

func (o *Mesh) SetParent(p Object) {
	o.parent = p
}

func (o *Mesh) Parent() Object {
	if o.parent == nil {
		return nil
	}

	return o.parent
}
