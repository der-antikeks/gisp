package game

import (
	"fmt"

	"github.com/der-antikeks/mathgl/mgl32"
)

const (
	ProjectionType ComponentType = 1 << iota
	TransformationType
	VelocityType
	GeometryType
	MaterialType

	OrbitControlType

	SceneTreeType

	// old
	MenuType
	MeshType
	PositionType
)

type Projection struct {
	Matrix mgl32.Mat4
	// Width, Height float64		-> update Projection via RenderSystem hooked to MessageResize
	// Rendertarget *Framebuffer	-> nil for screen
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
	Position mgl32.Vec3 // TODO: Vec4?
	Rotation mgl32.Quat
	Scale    mgl32.Vec3
	Up       mgl32.Vec3

	matrix        mgl32.Mat4
	updatedMatrix bool

	Parent   *Transformation // TODO: replace with Entity/engine.Get(parent, TransformationType)
	Children []*Transformation
}

func (c Transformation) Type() ComponentType {
	return TransformationType
}

func Compose(position mgl32.Vec3, rotation mgl32.Quat, scale mgl32.Vec3) mgl32.Mat4 {
	return mgl32.Translate3D(position[0], position[1], position[2]).
		Mul4(rotation.Mat4()).
		Mul4(mgl32.Scale3D(scale[0], scale[1], scale[2]))
}

func (c *Transformation) Matrix() mgl32.Mat4 {
	if !c.updatedMatrix {
		c.matrix = Compose(c.Position, c.Rotation, c.Scale)
		c.updatedMatrix = true
	}
	return c.matrix
}

func (c *Transformation) MatrixWorld() mgl32.Mat4 {
	if c.Parent != nil {
		c.Parent.MatrixWorld().Mul4(c.Matrix())
	}
	return c.Matrix()
}

type Velocity struct {
	Velocity mgl32.Vec3 // distance units(?)/sec
	Angular  mgl32.Vec3 // euler angles in radian/sec
}

func (c Velocity) Type() ComponentType {
	return VelocityType
}

type Geometry struct {
	File string
	mesh *meshbuffer

	Bounding Boundary
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

type SceneTree struct {
	Name string
	leaf *Node
}

func (c SceneTree) Type() ComponentType {
	return SceneTreeType
}
