package game

import (
	"log"
	m "math"
	"time"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"
)

type EntityManager struct {
	engine *ecs.Engine

	// TODO: move to separate loader/manager
	geometryCache map[string]Geometry
}

func NewEntityManager(e *ecs.Engine) *EntityManager {
	return &EntityManager{
		engine: e,

		geometryCache: map[string]Geometry{},
	}
}

func (em *EntityManager) Initalize() {
	s := em.engine.Entity()
	if err := em.engine.Set(
		s,
		GameStateComponent{"init", time.Now()},
	); err != nil {
		log.Fatal("could not initialize:", err)
	}
}

func (em *EntityManager) CreateSplashScreen() {
	em.createCube()
	em.createSphere()
}

func (em *EntityManager) CreateMainMenu() {}

func (em *EntityManager) createCube() {
	// Transformation
	trans := Transformation{
		Position: math.Vector{-1, 2, 0},
		Rotation: math.QuaternionFromAxisAngle(math.Vector{1, 0.5, 0}, m.Pi/4.0),
		Scale:    math.Vector{1, 1, 1},
		Up:       math.Vector{0, 1, 0},
	}

	geo := em.getGeometry("cube")
	mat := em.getMaterial("basic")
	mat.SetUniform("diffuse", math.Color{1, 0, 0})
	mat.SetUniform("opacity", 0.8)

	// Entity
	cube := em.engine.Entity()
	if err := em.engine.Set(
		cube,
		trans, geo, mat,
	); err != nil {
		log.Fatal("could not create cube:", err)
	}
}

func (em *EntityManager) createSphere() {
	// Transformation
	trans := Transformation{
		Position: math.Vector{0, 0, 1},
		Rotation: math.Quaternion{},
		Scale:    math.Vector{1, 1, 1},
		Up:       math.Vector{0, 1, 0},
	}

	geo := em.getGeometry("sphere")
	mat := em.getMaterial("basic")
	mat.SetUniform("diffuse", math.Color{0, 1, 0})

	// Entity
	sphere := em.engine.Entity()
	if err := em.engine.Set(
		sphere,
		trans, geo, mat,
	); err != nil {
		log.Fatal("could not create sphere:", err)
	}
}

func (em *EntityManager) getMaterial(id string) Material {
	mat := Material{
		Uniforms: map[string]interface{}{},
		Shader:   GetShader(id),
	}

	return mat
}

func (em *EntityManager) getGeometry(id string) Geometry {
	if g, found := em.geometryCache[id]; found {
		return g
	}

	// Geometry
	geo := Geometry{}

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
	t := Transformation{
		Position: math.Vector{0, 0, -10},
		//Rotation: math.Quaternion{},
		Scale: math.Vector{1, 1, 1},
		Up:    math.Vector{0, 1, 0},
	}
	t.Rotation = math.QuaternionFromRotationMatrix(math.LookAt(t.Position, math.Vector{0, 0, 0}, t.Up))

	c := em.engine.Entity()
	if err := em.engine.Set(
		c,
		Projection{
			Fovy:   fov,
			Aspect: aspect,
			Near:   near,
			Far:    far,
		},
		t,
	); err != nil {
		log.Fatal("could not create camera:", err)
	}
}
