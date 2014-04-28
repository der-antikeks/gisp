package game

import (
	m "math"
	"time"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
)

const (
	GameStateType ecs.ComponentType = iota

	PositionType
	VelocityType
	MotionControlType

	MeshType
	ColorType

	ProjectionType
	TransformationType
	GeometryType
	MaterialType
)

type GameStateComponent struct {
	State string
	Since time.Time
}

func (c GameStateComponent) Type() ecs.ComponentType {
	return GameStateType
}

type ColorComponent struct {
	R, G, B float64
}

func (c ColorComponent) Type() ecs.ComponentType {
	return ColorType
}

type Point struct {
	X, Y float64
}

func (p Point) Distance(o Point) float64 {
	dx, dy := o.X-p.X, o.Y-p.Y
	return m.Sqrt(dx*dx + dy*dy)
}

type PositionComponent struct {
	Position Point
	Rotation float64
}

func (c PositionComponent) Type() ecs.ComponentType {
	return PositionType
}

type VelocityComponent struct {
	Velocity Point
	Angular  float64
}

func (c VelocityComponent) Type() ecs.ComponentType {
	return VelocityType
}

type MeshComponent struct {
	Points []Point
	Max    float64
}

func (c MeshComponent) Type() ecs.ComponentType {
	return MeshType
}

type Camera struct {
	Projection math.Matrix
}

type Orthographic struct {
	left, right float64
	bottom, top float64
	near, far   float64

	matrix            math.Matrix
	matrixNeedsUpdate bool
}

type Projection struct {
	Fovy, Aspect, Near, Far float64

	matrix            math.Matrix
	matrixNeedsUpdate bool
}

func (c *Projection) ProjectionMatrix() math.Matrix {
	if c.matrixNeedsUpdate {
		c.matrix = math.NewPerspectiveMatrix(c.Fovy, c.Aspect, c.Near, c.Far)
		c.matrixNeedsUpdate = false
	}
	return c.matrix
}

func (c Projection) Type() ecs.ComponentType {
	return ProjectionType
}

type Transformation struct {
	Position math.Vector
	Rotation math.Quaternion
	Scale    math.Vector
	Up       math.Vector

	matrix            math.Matrix
	matrixNeedsUpdate bool

	Parent   *Transformation
	Children []*Transformation
}

func (c Transformation) Type() ecs.ComponentType {
	return TransformationType
}

func (c *Transformation) Matrix() math.Matrix {
	if c.matrixNeedsUpdate {
		c.matrix = math.ComposeMatrix(c.Position, c.Rotation, c.Scale)
		c.matrixNeedsUpdate = false
	}
	return c.matrix
}

func (c *Transformation) MatrixWorld() math.Matrix {
	if c.Parent != nil {
		c.Parent.MatrixWorld().Mul(c.Matrix())
	}
	return c.Matrix()
}

type Vertex struct {
	position math.Vector
	normal   math.Vector
	uv       math.Vector
}

type Face struct {
	A, B, C int
}

type Geometry struct {
	Vertices []Vertex
	Faces    []Face

	VertexArrayObject gl.VertexArray
	FaceBuffer        gl.Buffer
	PositionBuffer    gl.Buffer
	NormalBuffer      gl.Buffer
	UvBuffer          gl.Buffer
	initialized       bool

	FaceArray     []uint16 // uint32 (4 byte) if points > 65535
	PositionArray []float32
	NormalArray   []float32
	UvArray       []float32
	ColorArray    []float32
	needsUpdate   bool

	Bounding math.Boundary
}

func (c Geometry) Type() ecs.ComponentType {
	return GeometryType
}

type Material struct {
	Opaque  bool
	Program interface{} // shader program
}

func (c Material) Type() ecs.ComponentType {
	return MaterialType
}
