package game

import (
	"fmt"
	"log"
	m "math"
	"sync"

	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
	"github.com/go-gl/glh"
)

/*
	wavefront obj/mtl importer
	http://en.wikipedia.org/wiki/Wavefront_OBJ

	object format
	http://paulbourke.net/dataformats/obj/

	material format
	http://paulbourke.net/dataformats/mtl/

	obj
		o - named object (ignored)
			g - group of polygons

	mtl
		material
		...

var objCache = struct {
	sync.RWMutex
	geometry map[string]Geometry
	material map[string]Material
}{
	geometry: map[string]Geometry{},
	material: map[string]Material{},
}

func LoadObj(path string) (err error, found []string) {
	return nil, nil
}

func loadMtl(path string) (err error, found []string) {
	return nil, nil
}

*/

var meshbufferCache = struct {
	sync.Mutex
	meshbuffers map[string]*meshbuffer
}{
	meshbuffers: map[string]*meshbuffer{},
}

func GetMeshBuffer(name string) *meshbuffer {
	meshbufferCache.Lock()
	defer meshbufferCache.Unlock()

	if mb, found := meshbufferCache.meshbuffers[name]; found {
		return mb
	}

	mb := &meshbuffer{}

	switch name {
	default:
		log.Fatal("unknown geometry name: ", name)
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
		mb.AddFace(
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
		mb.AddFace(
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
		mb.AddFace(
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
		mb.AddFace(
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
		mb.AddFace(
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
		mb.AddFace(
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
		mb.AddFace(
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
		mb.AddFace(
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
		mb.AddFace(
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
		mb.AddFace(
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
		mb.AddFace(
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
		mb.AddFace(
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
					mb.AddFace(
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
					mb.AddFace(
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
					mb.AddFace(
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
					mb.AddFace(
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

	}

	mb.MergeVertices()
	mb.ComputeBoundary()
	mb.FaceCount = len(mb.Faces)
	mb.Init()

	meshbufferCache.meshbuffers[name] = mb
	return mb
}

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

type meshbuffer struct {
	Vertices    []Vertex // TODO: remove unnecessary slices
	Faces       []Face
	initialized bool

	VertexArrayObject gl.VertexArray
	FaceBuffer        gl.Buffer
	PositionBuffer    gl.Buffer
	NormalBuffer      gl.Buffer
	UvBuffer          gl.Buffer

	Bounding  math.Boundary
	FaceCount int
}

func (g *meshbuffer) AddFace(a, b, c Vertex) {
	offset := len(g.Vertices)
	g.Vertices = append(g.Vertices, a, b, c)
	g.Faces = append(g.Faces, Face{offset, offset + 1, offset + 2})
}

func (g *meshbuffer) MergeVertices() {
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

func (g *meshbuffer) ComputeBoundary() {
	g.Bounding = math.NewBoundary()
	for _, v := range g.Vertices {
		g.Bounding.AddPoint(v.position)
	}
}

func (g *meshbuffer) Init() {
	if g.initialized {
		return
	}

	MainThread(func() {
		// init vertex buffers
		g.VertexArrayObject = gl.GenVertexArray() // vao
		g.FaceBuffer = gl.GenBuffer()             // ebo
		g.PositionBuffer = gl.GenBuffer()         // vbo's
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
	})

	g.initialized = true
}

func (g *meshbuffer) Cleanup() {
	MainThread(func() {
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
	})
}
