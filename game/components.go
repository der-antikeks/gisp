package game

import (
	"fmt"
	"log"
	"time"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
	"github.com/go-gl/glh"
)

const (
	GameStateType ecs.ComponentType = iota

	ProjectionType
	TransformationType
	VelocityType
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

	matrix        math.Matrix
	updatedMatrix bool
}

func (c *Projection) ProjectionMatrix() math.Matrix {
	if !c.updatedMatrix {
		c.matrix = math.NewPerspectiveMatrix(c.Fovy, c.Aspect, c.Near, c.Far)
		c.updatedMatrix = true
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

	matrix        math.Matrix
	updatedMatrix bool

	Parent   *Transformation
	Children []*Transformation
}

func (c Transformation) Type() ecs.ComponentType {
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
	Position math.Vector
	Rotation math.Quaternion
}

func (c Velocity) Type() ecs.ComponentType {
	return VelocityType
}

// TODO: move geometry to separate loader manager
type Vertex struct {
	position math.Vector
	normal   math.Vector
	uv       math.Vector
}

func (v Vertex) Key(precision int) string {
	return fmt.Sprintf("%v_%v_%v_%v_%v_%v_%v_%v_%v_%v_%v",
		math.Round(v.position[0], precision),
		math.Round(v.position[1], precision),
		math.Round(v.position[2], precision),

		math.Round(v.normal[0], precision),
		math.Round(v.normal[1], precision),
		math.Round(v.normal[2], precision),

		math.Round(v.uv[0], precision),
		math.Round(v.uv[1], precision),
	)
}

type Face struct {
	A, B, C int
}

type Geometry struct {
	Vertices    []Vertex
	Faces       []Face
	initialized bool

	VertexArrayObject gl.VertexArray
	FaceBuffer        gl.Buffer
	PositionBuffer    gl.Buffer
	NormalBuffer      gl.Buffer
	UvBuffer          gl.Buffer

	Bounding math.Boundary
}

func (c Geometry) Type() ecs.ComponentType {
	return GeometryType
}

func (g *Geometry) AddFace(a, b, c Vertex) {
	offset := len(g.Vertices)
	g.Vertices = append(g.Vertices, a, b, c)
	g.Faces = append(g.Faces, Face{offset, offset + 1, offset + 2})
}

func (g *Geometry) MergeVertices() {
	// search and mark duplicate vertices
	lookup := map[string]int{}
	unique := []Vertex{}
	changed := map[int]int{}

	for i, v := range g.Vertices {
		key := v.Key(4)

		if j, found := lookup[key]; !found {
			// new vertex
			lookup[key] = i
			unique = append(unique, v)
			changed[i] = len(unique) - 1
		} else {
			// duplicate vertex
			changed[i] = changed[j]
		}
	}

	// change faces
	cleaned := []Face{}

	for _, f := range g.Faces {
		a, b, c := changed[f.A], changed[f.B], changed[f.C]
		if a == b || b == c || c == a {
			// degenerated face, remove
			continue
		}

		nf := Face{a, b, c}
		cleaned = append(cleaned, nf)
	}

	// replace with cleaned
	g.Vertices = unique
	g.Faces = cleaned
}

func (g *Geometry) ComputeBoundary() {
	g.Bounding = math.NewBoundary()
	for _, v := range g.Vertices {
		g.Bounding.AddPoint(v.position)
	}
}

func (g *Geometry) init() {
	if g.initialized {
		return
	}

	// init vertex buffers
	g.VertexArrayObject = gl.GenVertexArray()
	g.FaceBuffer = gl.GenBuffer()
	g.PositionBuffer = gl.GenBuffer()
	g.NormalBuffer = gl.GenBuffer()
	g.UvBuffer = gl.GenBuffer()

	g.VertexArrayObject.Bind()

	// init mesh buffers
	faceArray := make([]uint16, len(g.Faces)*3) // uint32 (4 byte) if points > 65535

	nvertices := len(g.Vertices)
	positionArray := make([]float32, nvertices*3)
	normalArray := make([]float32, nvertices*3)
	uvArray := make([]float32, nvertices*2)

	// copy values to buffers
	for i, v := range g.Vertices {
		// position
		positionArray[i*3] = float32(v.position[0])
		positionArray[i*3+1] = float32(v.position[1])
		positionArray[i*3+2] = float32(v.position[2])

		// normal
		normalArray[i*3] = float32(v.normal[0])
		normalArray[i*3+1] = float32(v.normal[1])
		normalArray[i*3+2] = float32(v.normal[2])

		// uv
		uvArray[i*2] = float32(v.uv[0])
		uvArray[i*2+1] = float32(v.uv[1])
	}

	for i, f := range g.Faces {
		faceArray[i*3] = uint16(f.A)
		faceArray[i*3+1] = uint16(f.B)
		faceArray[i*3+2] = uint16(f.C)
	}

	// set mesh buffers

	// position
	g.PositionBuffer.Bind(gl.ARRAY_BUFFER)
	size := len(positionArray) * int(glh.Sizeof(gl.FLOAT))              // float32 - gl.FLOAT, float64 - gl.DOUBLE
	gl.BufferData(gl.ARRAY_BUFFER, size, positionArray, gl.STATIC_DRAW) // gl.DYNAMIC_DRAW

	// normal
	g.NormalBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(normalArray) * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, normalArray, gl.STATIC_DRAW)

	// uv
	g.UvBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(uvArray) * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, uvArray, gl.STATIC_DRAW)

	// face
	g.FaceBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	size = len(faceArray) * int(glh.Sizeof(gl.UNSIGNED_SHORT)) // gl.UNSIGNED_SHORT 2, gl.UNSIGNED_INT 4
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, size, faceArray, gl.STATIC_DRAW)

	g.initialized = true
}

func (g *Geometry) cleanup() {
	if g.PositionBuffer != 0 {
		g.PositionBuffer.Delete()
	}

	if g.NormalBuffer != 0 {
		g.NormalBuffer.Delete()
	}

	if g.UvBuffer != 0 {
		g.UvBuffer.Delete()
	}

	if g.FaceBuffer != 0 {
		g.FaceBuffer.Delete()
	}

	if g.VertexArrayObject != 0 {
		g.VertexArrayObject.Delete()
	}
}

// TODO: move material to separate loader
type Texture interface {
	Bind(slot int)
	Unbind()
	Dispose()
}

type Material struct {
	Shader   *shader
	Opaque   bool
	Uniforms map[string]interface{}
}

func (c Material) Type() ecs.ComponentType {
	return MaterialType
}

func (m *Material) SetUniform(name string, value interface{}) {
	if _, allowed := m.Shader.uniforms[name]; !allowed {
		log.Fatalf("uniform %v not allowed for shader", name)
		return
	}
	m.Uniforms[name] = value
}

func (m *Material) UpdateUniforms() {
	var usedTextureUnits int

	for n, vu := range m.Shader.uniforms {
		v, found := m.Uniforms[n]
		if !found {
			v = vu.standard
		}

		switch t := v.(type) {
		case nil: // ignore nil
		default:
			log.Fatalf("%v has unknown type: %T", n, t)

		case Texture:
			t.Bind(usedTextureUnits)
			m.Shader.uniforms[n].location.Uniform1i(usedTextureUnits)

			usedTextureUnits++

		case int:
			m.Shader.uniforms[n].location.Uniform1i(t)
		case float64:
			m.Shader.uniforms[n].location.Uniform1f(float32(t))
		case float32:
			m.Shader.uniforms[n].location.Uniform1f(t)

		case [16]float32:
			m.Shader.uniforms[n].location.UniformMatrix4fv(false, t)
		case [9]float32:
			m.Shader.uniforms[n].location.UniformMatrix3fv(false, t)

		case math.Color:
			m.Shader.uniforms[n].location.Uniform3f(float32(t.R), float32(t.G), float32(t.B))

		case bool:
			if t {
				m.Shader.uniforms[n].location.Uniform1i(1)
			} else {
				m.Shader.uniforms[n].location.Uniform1i(0)
			}
		}
	}
}
