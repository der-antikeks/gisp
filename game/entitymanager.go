package game

import (
	"log"
	m "math"
	"time"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
)

type EntityManager struct {
	engine *ecs.Engine

	// TODO: move to separate loader/manager
	materialCache map[string]*Material
	geometryCache map[string]*Geometry
}

func NewEntityManager(e *ecs.Engine) *EntityManager {
	return &EntityManager{
		engine: e,

		materialCache: map[string]*Material{},
		geometryCache: map[string]*Geometry{},
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

func (m *EntityManager) CreateSplashScreen() {
	m.createCube()
	m.createSphere()
}

func (m *EntityManager) CreateMainMenu() {}

func (em *EntityManager) createCube() {
	// Transformation
	trans := &Transformation{
		Position: math.Vector{-2, 2, 0},
		Rotation: math.Quaternion{},
		Scale:    math.Vector{1, 1, 1},
		Up:       math.Vector{0, 1, 0},
	}

	geo := em.getGeometry("cube")
	mat := em.getMaterial("basic")

	// Entity
	cube := ecs.NewEntity(
		"cube", trans, geo, mat,
	)

	if err := em.engine.AddEntity(cube); err != nil {
		log.Fatal(err)
	}
}

func (em *EntityManager) createSphere() {
	// Transformation
	trans := &Transformation{
		Position: math.Vector{2, -2, 5},
		Rotation: math.Quaternion{},
		Scale:    math.Vector{1, 1, 1},
		Up:       math.Vector{0, 1, 0},
	}

	geo := em.getGeometry("sphere")
	mat := em.getMaterial("basic")

	// Entity
	sphere := ecs.NewEntity(
		"sphere", trans, geo, mat,
	)

	if err := em.engine.AddEntity(sphere); err != nil {
		log.Fatal(err)
	}
}

func (em *EntityManager) getMaterial(id string) *Material {
	if mat, found := em.materialCache[id]; found {
		return mat
	}

	mat := &Material{
		Uniforms: map[string]interface{}{},
		//Attributes: map[string]uint{},
		Shader: &shader{
			uniforms: map[string]struct {
				location gl.UniformLocation
				standard interface{}
			}{},
			attributes: map[string]struct {
				location gl.AttribLocation
				size     uint
				enabled  bool
			}{},
		},
	}

	switch id {
	default:
		log.Fatal("unknown material id: ", id)
	case "basic":
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
		mat.Shader.program = gl.CreateProgram()

		mat.Shader.program.AttachShader(vshader)
		mat.Shader.program.AttachShader(fshader)
		mat.Shader.program.Link()
		if mat.Shader.program.Get(gl.LINK_STATUS) != gl.TRUE {
			log.Fatalf("linker error: %v", mat.Shader.program.GetInfoLog())
		}

		// locations
		uniforms := map[string]interface{}{ // name, default value
			"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
			"viewMatrix":       nil, //[16]float32{},
			"modelMatrix":      nil, //[16]float32{},
			"modelViewMatrix":  nil, //[16]float32{},
			"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()
		}
		for n, v := range uniforms {
			mat.Shader.uniforms[n] = struct {
				location gl.UniformLocation
				standard interface{}
			}{
				location: mat.Shader.program.GetUniformLocation(n),
				standard: v,
			}
		}

		attributes := map[string]uint{ // name, size
			"vertexPosition": 3,
			"vertexNormal":   3,
			"vertexUV":       2,
		}
		for n, v := range attributes {
			mat.Shader.attributes[n] = struct {
				location gl.AttribLocation
				size     uint
				enabled  bool
			}{
				location: mat.Shader.program.GetAttribLocation(n),
				size:     v,
				enabled:  false,
			}
		}
	}

	em.materialCache[id] = mat
	return mat
}

func (em *EntityManager) getGeometry(id string) *Geometry {
	if g, found := em.geometryCache[id]; found {
		return g
	}

	// Geometry
	geo := &Geometry{}

	switch id {
	default:
		log.Fatal("unknown geometry id: ", id)
	case "cube":
		// dimensions
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

	case "sphere":
		// dimensions
		radius := 2.0
		widthSegments, heightSegments := 100, 50

		// if widthSegments < 3 {widthSegments = 3}
		// if heightSegments < 2 {heightSegments = 2}

		phiStart, phiLength := 0.0, math.Pi*2
		thetaStart, thetaLength := 0.0, math.Pi

		var vertices, uvs [][]math.Vector

		for y := 0; y <= heightSegments; y++ {
			var verticesRow, uvsRow []math.Vector

			for x := 0; x <= widthSegments; x++ {
				u := float64(x) / float64(widthSegments)
				v := float64(y) / float64(heightSegments)

				vertex := math.Vector{
					-radius * m.Cos(phiStart+u*phiLength) * m.Sin(thetaStart+v*thetaLength),
					radius * m.Cos(thetaStart+v*thetaLength),
					radius * m.Sin(phiStart+u*phiLength) * m.Sin(thetaStart+v*thetaLength),
				}

				verticesRow = append(verticesRow, vertex)
				uvsRow = append(uvsRow, math.Vector{u, 1.0 - v})
			}

			vertices = append(vertices, verticesRow)
			uvs = append(uvs, uvsRow)
		}

		for y := 0; y < heightSegments; y++ {
			for x := 0; x < widthSegments; x++ {
				// vertex id
				v1 := vertices[y][x+1]
				v2 := vertices[y][x]
				v3 := vertices[y+1][x]
				v4 := vertices[y+1][x+1]

				// normals
				n1 := v1.Normalize()
				n2 := v2.Normalize()
				n3 := v3.Normalize()
				n4 := v4.Normalize()

				// uvs
				uv1 := uvs[y][x+1]
				uv2 := uvs[y][x]
				uv3 := uvs[y+1][x]
				uv4 := uvs[y+1][x+1]

				if m.Abs(v1[1]) == radius {
					geo.AddFace(
						Vertex{
							position: v1,
							normal:   n1,
							uv:       uv1,
						}, Vertex{
							position: v3,
							normal:   n3,
							uv:       uv3,
						}, Vertex{
							position: v4,
							normal:   n4,
							uv:       uv4,
						})
				} else if m.Abs(v3[1]) == radius {
					geo.AddFace(
						Vertex{
							position: v1,
							normal:   n1,
							uv:       uv1,
						}, Vertex{
							position: v2,
							normal:   n2,
							uv:       uv2,
						}, Vertex{
							position: v3,
							normal:   n3,
							uv:       uv3,
						})
				} else {
					geo.AddFace(
						Vertex{
							position: v1,
							normal:   n1,
							uv:       uv1,
						}, Vertex{
							position: v2,
							normal:   n2,
							uv:       uv2,
						}, Vertex{
							position: v4,
							normal:   n4,
							uv:       uv4,
						})
					geo.AddFace(
						Vertex{
							position: v2,
							normal:   n2,
							uv:       uv2,
						}, Vertex{
							position: v3,
							normal:   n3,
							uv:       uv3,
						}, Vertex{
							position: v4,
							normal:   n4,
							uv:       uv4,
						})
				}
			}
		}

		geo.MergeVertices()
		geo.ComputeBoundary()
	}

	return geo
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
