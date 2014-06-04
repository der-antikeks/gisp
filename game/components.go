package game

import (
	"fmt"

	"github.com/der-antikeks/gisp/math"
)

const (
	ProjectionType ComponentType = 1 << iota
	TransformationType
	VelocityType
	GeometryType
	MaterialType

	OrbitControlType

	SceneType
	LightType

	// old
	MenuType
	MeshType
	PositionType
)

type Projection struct {
	Matrix        math.Matrix
	Width, Height float64 // update Projection via RenderSystem hooked to MessageResize

	rendertarget *Framebuffer // nil for screen
	priority     int
}

type Framebuffer struct {
	Color math.Color
	Alpha float64
}

func (c Projection) Type() ComponentType {
	return ProjectionType
}

/*
	Entity
		Material
			*shaderprogram <- Program string
				phong - standard mtl
				signed distance field font
				billboard
			Texture <- File string
			Dpacity
			Diffuse Color
		Geometry
			*meshbuffer <- File string
				buffers
			Bounding
		Transformation
			pos,rot,scale...
*/

type Transformation struct {
	Position math.Vector
	Rotation math.Quaternion
	Scale    math.Vector
	Up       math.Vector

	matrix        math.Matrix
	updatedMatrix bool

	Parent   *Transformation // TODO: replace with Entity/engine.Get(parent, TransformationType)
	Children []*Transformation
}

func (c Transformation) Type() ComponentType {
	return TransformationType
}

func (c *Transformation) Matrix() math.Matrix {
	if !c.updatedMatrix {
		c.matrix = math.ComposeMatrix(c.Position, c.Rotation, c.Scale)
		c.updatedMatrix = true
	}
	return c.matrix
}

func (c *Transformation) MatrixWorld() math.Matrix {
	if c.Parent != nil {
		c.Parent.MatrixWorld().Mul(c.Matrix())
	}
	return c.Matrix()
}

type Velocity struct {
	Velocity math.Vector // distance units(?)/sec
	Angular  math.Vector // euler angles in radian/sec
}

func (c Velocity) Type() ComponentType {
	return VelocityType
}

type Geometry struct {
	File string
	mesh *meshbuffer

	Bounding math.Boundary
}

func (c Geometry) Type() ComponentType {
	return GeometryType
}

type Material struct {
	Program string
	program *shaderprogram

	uniforms map[string]interface{}
}

func (m Material) Type() ComponentType {
	return MaterialType
}

func (m Material) opaque() bool {
	if o, f := m.uniforms["opacity"]; f {
		return o.(float64) >= 1.0
	}
	if o, f := m.program.uniforms["opacity"]; f {
		return o.standard.(float64) >= 1.0
	}
	return true
}

func (m *Material) Set(name string, value interface{}) error {
	if _, allowed := m.program.uniforms[name]; !allowed {
		return fmt.Errorf("uniform %v not allowed for shader program", name)
	}
	m.uniforms[name] = value
	return nil
}

func (m *Material) Get(name string) interface{} {
	if v, ok := m.uniforms[name]; ok {
		return v
	}
	return nil
}

type OrbitControl struct {
	MovementSpeed,
	RotationSpeed,
	ZoomSpeed float64

	Min, Max float64
	Target   Entity
}

func (c OrbitControl) Type() ComponentType {
	return OrbitControlType
}

type Scene struct {
	Name string
	leaf *Node
}

func (c Scene) Type() ComponentType {
	return SceneType
}

type Light struct {
	Diffuse math.Color
	Power   float64
}

func (c Light) Type() ComponentType {
	return LightType
}
