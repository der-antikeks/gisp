package game

import (
	"log"
	"time"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
)

type EntityManager struct {
	engine *ecs.Engine
}

func NewEntityManager(e *ecs.Engine) *EntityManager {
	return &EntityManager{
		engine: e,
	}
}

func (m *EntityManager) Initalize() {
	s := ecs.NewEntity(
		"game",
		&GameStateComponent{"init", time.Now()},
	)

	if err := m.engine.AddEntity(s); err != nil {
		log.Fatal(err)
	}
}

func (m *EntityManager) CreateSplashScreen() {}

func (m *EntityManager) CreateMainMenu() {}

func (em *EntityManager) CreateCube() {
	// Transformation
	trans := &Transformation{
		Position: math.Vector{-2, 2, 0},
		Rotation: math.Quaternion{},
		Scale:    math.Vector{1, 1, 1},
		Up:       math.Vector{0, 1, 0},
	}

	// Geometry
	geo := &Geometry{}
	size := 2.0
	halfSize := size / 2.0

	// vertices
	a := math.Vector{halfSize, halfSize, halfSize}
	b := math.Vector{-halfSize, halfSize, halfSize}
	c := math.Vector{-halfSize, -halfSize, halfSize}
	d := math.Vector{halfSize, -halfSize, halfSize}
	e := math.Vector{halfSize, halfSize, -halfSize}
	f := math.Vector{halfSize, -halfSize, -halfSize}
	g := math.Vector{-halfSize, -halfSize, -halfSize}
	h := math.Vector{-halfSize, halfSize, -halfSize}

	// uvs
	tl := math.Vector{0, 1}
	tr := math.Vector{1, 1}
	bl := math.Vector{0, 0}
	br := math.Vector{1, 0}

	var normal math.Vector

	// front
	normal = math.Vector{0, 0, 1}
	geo.AddFace(
		Vertex{ // a
			position: a,
			normal:   normal,
			uv:       tr,
		}, Vertex{ // b
			position: b,
			normal:   normal,
			uv:       tl,
		}, Vertex{ // c
			position: c,
			normal:   normal,
			uv:       bl,
		})
	geo.AddFace(
		Vertex{
			position: c,
			normal:   normal,
			uv:       bl,
		}, Vertex{
			position: d,
			normal:   normal,
			uv:       br,
		}, Vertex{
			position: a,
			normal:   normal,
			uv:       tr,
		})

	// back
	normal = math.Vector{0, 0, -1}
	geo.AddFace(
		Vertex{
			position: e,
			normal:   normal,
			uv:       tl,
		}, Vertex{
			position: f,
			normal:   normal,
			uv:       bl,
		}, Vertex{
			position: g,
			normal:   normal,
			uv:       br,
		})
	geo.AddFace(
		Vertex{
			position: g,
			normal:   normal,
			uv:       br,
		}, Vertex{
			position: h,
			normal:   normal,
			uv:       tr,
		}, Vertex{
			position: e,
			normal:   normal,
			uv:       tl,
		})

	// top
	normal = math.Vector{0, 1, 0}
	geo.AddFace(
		Vertex{
			position: e,
			normal:   normal,
			uv:       tr,
		}, Vertex{
			position: h,
			normal:   normal,
			uv:       tl,
		}, Vertex{
			position: b,
			normal:   normal,
			uv:       bl,
		})
	geo.AddFace(
		Vertex{
			position: b,
			normal:   normal,
			uv:       bl,
		}, Vertex{
			position: a,
			normal:   normal,
			uv:       br,
		}, Vertex{
			position: e,
			normal:   normal,
			uv:       tr,
		})

	// bottom
	normal = math.Vector{0, -1, 0}
	geo.AddFace(
		Vertex{
			position: f,
			normal:   normal,
			uv:       br,
		}, Vertex{
			position: d,
			normal:   normal,
			uv:       tr,
		}, Vertex{
			position: c,
			normal:   normal,
			uv:       tl,
		})
	geo.AddFace(
		Vertex{
			position: c,
			normal:   normal,
			uv:       tl,
		}, Vertex{
			position: g,
			normal:   normal,
			uv:       bl,
		}, Vertex{
			position: f,
			normal:   normal,
			uv:       br,
		})

	// left
	normal = math.Vector{-1, 0, 0}
	geo.AddFace(
		Vertex{
			position: b,
			normal:   normal,
			uv:       tr,
		}, Vertex{
			position: h,
			normal:   normal,
			uv:       tl,
		}, Vertex{
			position: g,
			normal:   normal,
			uv:       bl,
		})
	geo.AddFace(
		Vertex{
			position: g,
			normal:   normal,
			uv:       bl,
		}, Vertex{
			position: c,
			normal:   normal,
			uv:       br,
		}, Vertex{
			position: b,
			normal:   normal,
			uv:       tr,
		})

	// right
	normal = math.Vector{1, 0, 0}
	geo.AddFace(
		Vertex{
			position: e,
			normal:   normal,
			uv:       tr,
		}, Vertex{
			position: a,
			normal:   normal,
			uv:       tl,
		}, Vertex{
			position: d,
			normal:   normal,
			uv:       bl,
		})
	geo.AddFace(
		Vertex{
			position: d,
			normal:   normal,
			uv:       bl,
		}, Vertex{
			position: f,
			normal:   normal,
			uv:       br,
		}, Vertex{
			position: e,
			normal:   normal,
			uv:       tr,
		})

	geo.MergeVertices()
	geo.ComputeBoundary()

	// Material
	mat := &Material{
		Uniforms:   make(map[string]interface{}),
		Attributes: make(map[string]uint),
	}

	// vertex shader
	vshader := gl.CreateShader(gl.VERTEX_SHADER)
	vshader.Source(`
				#version 330 core

				// Input vertex data, different for all executions of this shader.
				in vec3 vertexPosition;
				in vec3 vertexNormal;
				in vec2 vertexUV;

				// Values that stay constant for the whole mesh.
				uniform mat4 projectionMatrix;
				uniform mat4 viewMatrix;
				uniform mat4 modelMatrix;
				uniform mat4 modelViewMatrix;
				uniform mat3 normalMatrix;

				void main(){
					// Output position of the vertex
					//gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
					gl_Position = projectionMatrix * modelViewMatrix * vec4(vertexPosition, 1.0);
				}`)
	vshader.Compile()
	if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		log.Fatalf("vertex shader error: %v", vshader.GetInfoLog())
	}
	defer vshader.Delete()

	// fragment shader
	fshader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fshader.Source(`
				#version 330 core

				// Output data
				out vec4 fragmentColor;

				void main()
				{
					fragmentColor = vec4(1, 1, 0, 1);
				}`)
	fshader.Compile()
	if fshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		log.Fatalf("fragment shader error: %v", fshader.GetInfoLog())
	}
	defer fshader.Delete()

	// program
	prg := &program{
		program:  gl.CreateProgram(),
		uniforms: make(map[string]gl.UniformLocation),
		attributes: make(map[string]struct {
			location gl.AttribLocation
			enabled  bool
		}),
	}

	prg.program.AttachShader(vshader)
	prg.program.AttachShader(fshader)
	prg.program.Link()
	if prg.program.Get(gl.LINK_STATUS) != gl.TRUE {
		log.Fatalf("linker error: %v", prg.program.GetInfoLog())
	}

	// locations
	uniforms := map[string]interface{}{
		"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
		"viewMatrix":       nil, //[16]float32{},
		"modelMatrix":      nil, //[16]float32{},
		"modelViewMatrix":  nil, //[16]float32{},
		"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()
	}
	for n, _ := range uniforms {
		prg.uniforms[n] = prg.program.GetUniformLocation(n)
	}

	attributes := map[string]uint{
		"vertexPosition": 3,
		"vertexNormal":   3,
		"vertexUV":       2,
	}
	for n, _ := range attributes {
		prg.attributes[n] = struct {
			location gl.AttribLocation
			enabled  bool
		}{
			location: prg.program.GetAttribLocation(n),
			enabled:  false,
		}
	}

	mat.Program = prg

	// default values
	for n, v := range uniforms {
		mat.Uniforms[n] = v
	}

	for n, v := range attributes {
		mat.Attributes[n] = v
	}

	// Entity
	cube := ecs.NewEntity(
		"cube", trans, geo, mat,
	)

	if err := em.engine.AddEntity(cube); err != nil {
		log.Fatal(err)
	}
}

func (em *EntityManager) CreatePerspectiveCamera(fov, aspect, near, far float64) {
	t := &Transformation{
		Position: math.Vector{0, 0, -10},
		//Rotation: math.Quaternion{},
		Scale: math.Vector{1, 1, 1},
		Up:    math.Vector{0, 1, 0},
	}
	t.Rotation = math.QuaternionFromRotationMatrix(math.LookAt(t.Position, math.Vector{0, 0, 0}, t.Up))

	c := ecs.NewEntity(
		"camera",
		&Projection{
			Fovy:   fov,
			Aspect: aspect,
			Near:   near,
			Far:    far,
		}, t,
	)

	if err := em.engine.AddEntity(c); err != nil {
		log.Fatal(err)
	}
}
