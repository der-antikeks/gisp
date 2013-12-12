package engine

import (
	"github.com/der-antikeks/gisp/math"
)

type Camera interface {
	Object

	UpdateProjectionMatrix()
	ProjectionMatrix() math.Matrix
}

type PerspectiveCamera struct {
	Camera

	// projection
	fov    float64
	aspect float64
	near   float64
	far    float64

	projectionNeedsUpdate bool
	projectionMatrix      math.Matrix

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

func NewPerspectiveCamera(fov, aspect, near, far float64) *PerspectiveCamera {
	c := &PerspectiveCamera{
		fov:    fov,
		aspect: aspect,
		near:   near,
		far:    far,

		projectionNeedsUpdate: true,
		projectionMatrix:      math.Identity(),

		up:    math.Vector{0, 1, 0},
		scale: math.Vector{1, 1, 1},

		matrixNeedsUpdate:      true,
		matrixWorldNeedsUpdate: true,
	}

	return c
}

func (c *PerspectiveCamera) SetFov(f float64) {
	c.fov = f
	c.projectionNeedsUpdate = true
}

func (c *PerspectiveCamera) SetAspect(f float64) {
	c.aspect = f
	c.projectionNeedsUpdate = true
}

func (c *PerspectiveCamera) SetNear(f float64) {
	c.near = f
	c.projectionNeedsUpdate = true
}

func (c *PerspectiveCamera) SetFar(f float64) {
	c.far = f
	c.projectionNeedsUpdate = true
}

func (c *PerspectiveCamera) UpdateProjectionMatrix() {
	c.projectionMatrix = math.NewPerspectiveMatrix(c.fov, c.aspect, c.near, c.far)
	c.projectionNeedsUpdate = false
}

func (c *PerspectiveCamera) ProjectionMatrix() math.Matrix {
	if c.projectionNeedsUpdate {
		c.UpdateProjectionMatrix()
	}

	return c.projectionMatrix
}

func (c *PerspectiveCamera) Position() math.Vector {
	return c.position
}

func (c *PerspectiveCamera) SetPosition(v math.Vector) {
	c.position = v
	c.matrixNeedsUpdate = true
}

func (c *PerspectiveCamera) SetUp(v math.Vector) {
	c.up = v.Normalize()
	c.matrixNeedsUpdate = true
}

func (c *PerspectiveCamera) Up() math.Vector {
	return c.up
}

func (c *PerspectiveCamera) LookAt(v math.Vector) {
	c.SetRotation(math.QuaternionFromRotationMatrix(math.LookAt(c.position, v, c.up)))
}

func (c *PerspectiveCamera) SetRotation(r math.Quaternion) {
	c.rotation = r
	c.matrixNeedsUpdate = true
}

func (c *PerspectiveCamera) Rotation() math.Quaternion {
	return c.rotation
}

func (c *PerspectiveCamera) SetScale(s math.Vector) {
	c.scale = s
	c.matrixNeedsUpdate = true
}

func (c *PerspectiveCamera) Scale() math.Vector {
	return c.scale
}

func (c *PerspectiveCamera) Matrix() math.Matrix {
	if c.matrixNeedsUpdate {
		c.matrix = math.ComposeMatrix(c.position, c.rotation, c.scale)

		c.matrixWorldNeedsUpdate = true
		c.matrixNeedsUpdate = false
	}

	return c.matrix
}

func (c *PerspectiveCamera) UpdateMatrixWorld(force bool) {
	m := c.Matrix()

	if c.matrixWorldNeedsUpdate || force {
		if p := c.Parent(); p == nil {
			c.matrixWorld = m
		} else {
			c.matrixWorld = p.MatrixWorld().Mul(m)
		}

		c.matrixWorldNeedsUpdate = false
		force = true
	}

	for _, c := range c.Children() {
		c.UpdateMatrixWorld(force)
	}
}

func (c *PerspectiveCamera) MatrixWorld() math.Matrix {
	return c.matrixWorld
}

func (c *PerspectiveCamera) AddChild(cs ...Object) {
	for _, ch := range cs {
		if c == ch {
			continue
		}

		if p := ch.Parent(); p != nil {
			p.RemoveChild(ch)
		}
		ch.SetParent(c)

		c.children = append(c.children, ch)

		// backward search for root
		var root, parent Object
		for parent = ch; parent != nil; parent = parent.Parent() {
			root = parent
		}

		if scene, ok := root.(*Scene); ok {
			scene.AddObject(ch)
		}
	}
}

func (c *PerspectiveCamera) RemoveChild(r Object) {
	r.SetParent(nil)

	position := -1
	for i, c := range c.children {
		if r == c {
			position = i
			break
		}
	}

	if position != -1 {
		copy(c.children[position:], c.children[position+1:])
		c.children[len(c.children)-1] = nil
		c.children = c.children[:len(c.children)-1]

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

func (c *PerspectiveCamera) Children() []Object {
	return c.children
}

func (c *PerspectiveCamera) SetParent(p Object) {
	c.parent = p
}

func (c *PerspectiveCamera) Parent() Object {
	if c.parent == nil {
		return nil
	}

	return c.parent
}

type OrthographicCamera struct {
	Camera

	// projection
	left, right float64
	bottom, top float64
	near, far   float64

	projectionNeedsUpdate bool
	projectionMatrix      math.Matrix

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

func NewOrthographicCamera(left, right, top, bottom, near, far float64) *OrthographicCamera {
	c := &OrthographicCamera{
		left:   left,
		right:  right,
		top:    top,
		bottom: bottom,
		near:   near,
		far:    far,

		projectionNeedsUpdate: true,
		projectionMatrix:      math.Identity(),

		up:    math.Vector{0, 1, 0},
		scale: math.Vector{1, 1, 1},

		matrixNeedsUpdate:      true,
		matrixWorldNeedsUpdate: true,
	}

	return c
}

func (c *OrthographicCamera) SetLeft(f float64) {
	c.left = f
	c.projectionNeedsUpdate = true
}

func (c *OrthographicCamera) SetRight(f float64) {
	c.right = f
	c.projectionNeedsUpdate = true
}

func (c *OrthographicCamera) SetTop(f float64) {
	c.top = f
	c.projectionNeedsUpdate = true
}

func (c *OrthographicCamera) SetBottom(f float64) {
	c.bottom = f
	c.projectionNeedsUpdate = true
}

func (c *OrthographicCamera) SetNear(f float64) {
	c.near = f
	c.projectionNeedsUpdate = true
}

func (c *OrthographicCamera) SetFar(f float64) {
	c.far = f
	c.projectionNeedsUpdate = true
}

func (c *OrthographicCamera) UpdateProjectionMatrix() {
	c.projectionMatrix = math.NewOrthoMatrix(c.left, c.right, c.bottom, c.top, c.near, c.far)

	c.projectionNeedsUpdate = false
}

func (c *OrthographicCamera) ProjectionMatrix() math.Matrix {
	if c.projectionNeedsUpdate {
		c.UpdateProjectionMatrix()
	}

	return c.projectionMatrix
}

func (c *OrthographicCamera) Position() math.Vector {
	return c.position
}

func (c *OrthographicCamera) SetPosition(v math.Vector) {
	c.position = v
	c.matrixNeedsUpdate = true
}

func (c *OrthographicCamera) SetUp(v math.Vector) {
	c.up = v.Normalize()
	c.matrixNeedsUpdate = true
}

func (c *OrthographicCamera) Up() math.Vector {
	return c.up
}

func (c *OrthographicCamera) LookAt(v math.Vector) {
	c.SetRotation(math.QuaternionFromRotationMatrix(math.LookAt(c.position, v, c.up)))
}

func (c *OrthographicCamera) SetRotation(r math.Quaternion) {
	c.rotation = r
	c.matrixNeedsUpdate = true
}

func (c *OrthographicCamera) Rotation() math.Quaternion {
	return c.rotation
}

func (c *OrthographicCamera) SetScale(s math.Vector) {
	c.scale = s
	c.matrixNeedsUpdate = true
}

func (c *OrthographicCamera) Scale() math.Vector {
	return c.scale
}

func (c *OrthographicCamera) Matrix() math.Matrix {
	if c.matrixNeedsUpdate {
		c.matrix = math.ComposeMatrix(c.position, c.rotation, c.scale)

		c.matrixWorldNeedsUpdate = true
		c.matrixNeedsUpdate = false
	}

	return c.matrix
}

func (c *OrthographicCamera) UpdateMatrixWorld(force bool) {
	m := c.Matrix()

	if c.matrixWorldNeedsUpdate || force {
		if p := c.Parent(); p == nil {
			c.matrixWorld = m
		} else {
			c.matrixWorld = p.MatrixWorld().Mul(m)
		}

		c.matrixWorldNeedsUpdate = false
		force = true
	}

	for _, c := range c.Children() {
		c.UpdateMatrixWorld(force)
	}
}

func (c *OrthographicCamera) MatrixWorld() math.Matrix {
	return c.matrixWorld
}

func (c *OrthographicCamera) AddChild(cs ...Object) {
	for _, ch := range cs {
		if c == ch {
			continue
		}

		if p := ch.Parent(); p != nil {
			p.RemoveChild(ch)
		}
		ch.SetParent(c)

		c.children = append(c.children, ch)

		// backward search for root
		var root, parent Object
		for parent = ch; parent != nil; parent = parent.Parent() {
			root = parent
		}

		if scene, ok := root.(*Scene); ok {
			scene.AddObject(ch)
		}
	}
}

func (c *OrthographicCamera) RemoveChild(r Object) {
	r.SetParent(nil)

	position := -1
	for i, c := range c.children {
		if r == c {
			position = i
			break
		}
	}

	if position != -1 {
		copy(c.children[position:], c.children[position+1:])
		c.children[len(c.children)-1] = nil
		c.children = c.children[:len(c.children)-1]

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

func (c *OrthographicCamera) Children() []Object {
	return c.children
}

func (c *OrthographicCamera) SetParent(p Object) {
	c.parent = p
}

func (c *OrthographicCamera) Parent() Object {
	if c.parent == nil {
		return nil
	}

	return c.parent
}
