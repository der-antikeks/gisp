package engine

import (
	"fmt"
	m "math"

	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
	"github.com/go-gl/glh"
)

type Vertex struct {
	position math.Vector
	normal   math.Vector
	uv       math.Vector
	color    math.Color
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

		math.Round(v.color.R, precision),
		math.Round(v.color.G, precision),
		math.Round(v.color.B, precision),
	)
}

type Face struct {
	A, B, C int
}

func (f Face) ToLines() []Line {
	return []Line{
		Line{f.A, f.B},
		Line{f.B, f.C},
		Line{f.C, f.A},
	}
}

type Line struct {
	A, B int
}

type Geometry struct {
	// data slices
	vertices []Vertex
	faces    []Face
	lines    []Line

	// gl buffers
	vertexArrayObject gl.VertexArray
	faceBuffer        gl.Buffer
	lineBuffer        gl.Buffer
	positionBuffer    gl.Buffer
	normalBuffer      gl.Buffer
	uvBuffer          gl.Buffer
	colorBuffer       gl.Buffer
	initialized       bool

	faceArray     []uint16 // uint32 (4 byte) if points > 65535
	lineArray     []uint16
	positionArray []float32
	normalArray   []float32
	uvArray       []float32
	colorArray    []float32
	needsUpdate   bool

	hint gl.GLenum //  gl.STATIC_DRAW, gl.DYNAMIC_DRAW

	// boundings
	bounding math.Boundary
}

func NewGeometry() *Geometry {
	return &Geometry{
		initialized: false,
		needsUpdate: true,

		hint: gl.STATIC_DRAW, // gl.DYNAMIC_DRAW,
	}
}

func NewSphereGeometry(radius float64, widthSegments, heightSegments int) *Geometry {
	geo := &Geometry{
		initialized: false,
		needsUpdate: true,

		hint: gl.STATIC_DRAW, // gl.DYNAMIC_DRAW,
	}

	if widthSegments < 3 {
		widthSegments = 3
	}
	if heightSegments < 2 {
		heightSegments = 2
	}

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

	color := math.Color{1, 1, 1}

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
						color:    color,
					}, Vertex{
						position: v3,
						normal:   n3,
						uv:       uv3,
						color:    color,
					}, Vertex{
						position: v4,
						normal:   n4,
						uv:       uv4,
						color:    color,
					})
			} else if m.Abs(v3[1]) == radius {
				geo.AddFace(
					Vertex{
						position: v1,
						normal:   n1,
						uv:       uv1,
						color:    color,
					}, Vertex{
						position: v2,
						normal:   n2,
						uv:       uv2,
						color:    color,
					}, Vertex{
						position: v3,
						normal:   n3,
						uv:       uv3,
						color:    color,
					})
			} else {
				geo.AddFace(
					Vertex{
						position: v1,
						normal:   n1,
						uv:       uv1,
						color:    color,
					}, Vertex{
						position: v2,
						normal:   n2,
						uv:       uv2,
						color:    color,
					}, Vertex{
						position: v4,
						normal:   n4,
						uv:       uv4,
						color:    color,
					})
				geo.AddFace(
					Vertex{
						position: v2,
						normal:   n2,
						uv:       uv2,
						color:    color,
					}, Vertex{
						position: v3,
						normal:   n3,
						uv:       uv3,
						color:    color,
					}, Vertex{
						position: v4,
						normal:   n4,
						uv:       uv4,
						color:    color,
					})
			}
		}
	}

	geo.MergeVertices()
	geo.ComputeBoundary()

	return geo
}

func NewCubeGeometry(size float64) *Geometry {
	geo := &Geometry{
		initialized: false,
		needsUpdate: true,

		hint: gl.STATIC_DRAW, // gl.DYNAMIC_DRAW,
	}

	halfSize := size / 2.0

	/*
		    vertices			uvs

		  h +------+ e
			|\     |\
			| \    | \
			|b +------+ a   tl +------+ tr
		  g +--|---+ f|        |      |
			 \ |    \ |        |      |
			  \|     \|        |      |
			 c +------+ d   bl +------+ br
	*/

	// vertices
	a := math.Vector{halfSize, halfSize, halfSize}
	b := math.Vector{-halfSize, halfSize, halfSize}
	c := math.Vector{-halfSize, -halfSize, halfSize}
	d := math.Vector{halfSize, -halfSize, halfSize}
	e := math.Vector{halfSize, halfSize, -halfSize}
	f := math.Vector{halfSize, -halfSize, -halfSize}
	g := math.Vector{-halfSize, -halfSize, -halfSize}
	h := math.Vector{-halfSize, halfSize, -halfSize}

	white := math.Color{1, 1, 1}

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
			color:    white,
		}, Vertex{ // b
			position: b,
			normal:   normal,
			uv:       tl,
			color:    white,
		}, Vertex{ // c
			position: c,
			normal:   normal,
			uv:       bl,
			color:    white,
		})
	geo.AddFace(
		Vertex{
			position: c,
			normal:   normal,
			uv:       bl,
			color:    white,
		}, Vertex{
			position: d,
			normal:   normal,
			uv:       br,
			color:    white,
		}, Vertex{
			position: a,
			normal:   normal,
			uv:       tr,
			color:    white,
		})

	// back
	normal = math.Vector{0, 0, -1}
	geo.AddFace(
		Vertex{
			position: e,
			normal:   normal,
			uv:       tl,
			color:    white,
		}, Vertex{
			position: f,
			normal:   normal,
			uv:       bl,
			color:    white,
		}, Vertex{
			position: g,
			normal:   normal,
			uv:       br,
			color:    white,
		})
	geo.AddFace(
		Vertex{
			position: g,
			normal:   normal,
			uv:       br,
			color:    white,
		}, Vertex{
			position: h,
			normal:   normal,
			uv:       tr,
			color:    white,
		}, Vertex{
			position: e,
			normal:   normal,
			uv:       tl,
			color:    white,
		})

	// top
	normal = math.Vector{0, 1, 0}
	geo.AddFace(
		Vertex{
			position: e,
			normal:   normal,
			uv:       tr,
			color:    white,
		}, Vertex{
			position: h,
			normal:   normal,
			uv:       tl,
			color:    white,
		}, Vertex{
			position: b,
			normal:   normal,
			uv:       bl,
			color:    white,
		})
	geo.AddFace(
		Vertex{
			position: b,
			normal:   normal,
			uv:       bl,
			color:    white,
		}, Vertex{
			position: a,
			normal:   normal,
			uv:       br,
			color:    white,
		}, Vertex{
			position: e,
			normal:   normal,
			uv:       tr,
			color:    white,
		})

	// bottom
	normal = math.Vector{0, -1, 0}
	geo.AddFace(
		Vertex{
			position: f,
			normal:   normal,
			uv:       br,
			color:    white,
		}, Vertex{
			position: d,
			normal:   normal,
			uv:       tr,
			color:    white,
		}, Vertex{
			position: c,
			normal:   normal,
			uv:       tl,
			color:    white,
		})
	geo.AddFace(
		Vertex{
			position: c,
			normal:   normal,
			uv:       tl,
			color:    white,
		}, Vertex{
			position: g,
			normal:   normal,
			uv:       bl,
			color:    white,
		}, Vertex{
			position: f,
			normal:   normal,
			uv:       br,
			color:    white,
		})

	// left
	normal = math.Vector{-1, 0, 0}
	geo.AddFace(
		Vertex{
			position: b,
			normal:   normal,
			uv:       tr,
			color:    white,
		}, Vertex{
			position: h,
			normal:   normal,
			uv:       tl,
			color:    white,
		}, Vertex{
			position: g,
			normal:   normal,
			uv:       bl,
			color:    white,
		})
	geo.AddFace(
		Vertex{
			position: g,
			normal:   normal,
			uv:       bl,
			color:    white,
		}, Vertex{
			position: c,
			normal:   normal,
			uv:       br,
			color:    white,
		}, Vertex{
			position: b,
			normal:   normal,
			uv:       tr,
			color:    white,
		})

	// right
	normal = math.Vector{1, 0, 0}
	geo.AddFace(
		Vertex{
			position: e,
			normal:   normal,
			uv:       tr,
			color:    white,
		}, Vertex{
			position: a,
			normal:   normal,
			uv:       tl,
			color:    white,
		}, Vertex{
			position: d,
			normal:   normal,
			uv:       bl,
			color:    white,
		})
	geo.AddFace(
		Vertex{
			position: d,
			normal:   normal,
			uv:       bl,
			color:    white,
		}, Vertex{
			position: f,
			normal:   normal,
			uv:       br,
			color:    white,
		}, Vertex{
			position: e,
			normal:   normal,
			uv:       tr,
			color:    white,
		})

	geo.MergeVertices()
	geo.ComputeBoundary()

	return geo
}

func NewPlaneGeometry(width, height float64) *Geometry {
	geo := &Geometry{
		initialized: false,
		needsUpdate: true,

		hint: gl.STATIC_DRAW, // gl.DYNAMIC_DRAW,
	}

	halfWidth := width / 2.0
	halfHeight := height / 2.0

	/*
		    vertices			uvs

			b +------+ a   tl +------+ tr
		      |      |        |      |
			  |      |        |      |
			  |      |        |      |
			c +------+ d   bl +------+ br
	*/

	// vertices
	a := math.Vector{halfWidth, halfHeight, 0}
	b := math.Vector{-halfWidth, halfHeight, 0}
	c := math.Vector{-halfWidth, -halfHeight, 0}
	d := math.Vector{halfWidth, -halfHeight, 0}

	tl := math.Vector{0, 1}
	tr := math.Vector{1, 1}
	bl := math.Vector{0, 0}
	br := math.Vector{1, 0}

	normal := math.Vector{0, 0, 1}
	color := math.Color{1, 1, 1}

	geo.AddFace(
		Vertex{
			position: a,
			normal:   normal,
			uv:       tr,
			color:    color,
		}, Vertex{
			position: b,
			normal:   normal,
			uv:       tl,
			color:    color,
		}, Vertex{
			position: c,
			normal:   normal,
			uv:       bl,
			color:    color,
		})
	geo.AddFace(
		Vertex{
			position: c,
			normal:   normal,
			uv:       bl,
			color:    color,
		}, Vertex{
			position: d,
			normal:   normal,
			uv:       br,
			color:    color,
		}, Vertex{
			position: a,
			normal:   normal,
			uv:       tr,
			color:    color,
		})

	geo.MergeVertices()
	geo.ComputeBoundary()

	return geo
}

func (g *Geometry) AddFace(a, b, c Vertex) {
	offset := len(g.vertices)
	g.vertices = append(g.vertices, a, b, c)
	g.faces = append(g.faces, Face{offset, offset + 1, offset + 2})
	g.lines = append(g.lines, Line{offset, offset + 1}, Line{offset + 1, offset + 2}, Line{offset + 2, offset})
}

func (g *Geometry) MergeVertices() {
	// search and mark duplicate vertices
	lookup := map[string]int{}
	unique := []Vertex{}
	changed := map[int]int{}

	for i, v := range g.vertices {
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
	lines := []Line{}

	for _, f := range g.faces {
		a, b, c := changed[f.A], changed[f.B], changed[f.C]
		if a == b || b == c || c == a {
			// degenerated face, remove
			continue
		}

		nf := Face{a, b, c}
		cleaned = append(cleaned, nf)
		lines = append(lines, nf.ToLines()...)
	}

	// replace with cleaned
	g.vertices = unique
	g.faces = cleaned
	g.lines = lines
}

func (g *Geometry) ComputeBoundary() {
	g.bounding = math.NewBoundary()
	for _, v := range g.vertices {
		g.bounding.AddPoint(v.position)
	}
}

func (g *Geometry) Boundary() math.Boundary {
	return g.bounding
}

// init vertex buffers
func (g *Geometry) init() {
	g.vertexArrayObject = gl.GenVertexArray()
	g.faceBuffer = gl.GenBuffer()
	g.lineBuffer = gl.GenBuffer()
	g.positionBuffer = gl.GenBuffer()
	g.normalBuffer = gl.GenBuffer()
	g.uvBuffer = gl.GenBuffer()
	g.colorBuffer = gl.GenBuffer()

	g.initialized = true
}

func (g *Geometry) update() {
	// split faces(vertices, colors, etc) between different face materials -- if material is face material, every material type has its own buffers
	if !g.initialized {
		g.init()
	}

	g.vertexArrayObject.Bind()

	// init mesh buffers
	g.faceArray = make([]uint16, len(g.faces)*3)
	g.lineArray = make([]uint16, len(g.lines)*2)

	nvertices := len(g.vertices)
	g.positionArray = make([]float32, nvertices*3)
	g.normalArray = make([]float32, nvertices*3)
	g.uvArray = make([]float32, nvertices*2)
	g.colorArray = make([]float32, nvertices*3)

	// copy values to buffers
	for i, v := range g.vertices {
		// position
		g.positionArray[i*3] = float32(v.position[0])
		g.positionArray[i*3+1] = float32(v.position[1])
		g.positionArray[i*3+2] = float32(v.position[2])

		// normal
		g.normalArray[i*3] = float32(v.normal[0])
		g.normalArray[i*3+1] = float32(v.normal[1])
		g.normalArray[i*3+2] = float32(v.normal[2])

		// uv
		g.uvArray[i*2] = float32(v.uv[0])
		g.uvArray[i*2+1] = float32(v.uv[1])

		// color
		g.colorArray[i*3] = float32(v.color.R)
		g.colorArray[i*3+1] = float32(v.color.G)
		g.colorArray[i*3+2] = float32(v.color.B)
	}

	for i, f := range g.faces {
		g.faceArray[i*3] = uint16(f.A)
		g.faceArray[i*3+1] = uint16(f.B)
		g.faceArray[i*3+2] = uint16(f.C)
	}

	for i, l := range g.lines {
		g.lineArray[i*2] = uint16(l.A)
		g.lineArray[i*2+1] = uint16(l.B)
	}

	// set mesh buffers

	// position
	g.positionBuffer.Bind(gl.ARRAY_BUFFER)
	size := len(g.positionArray) * int(glh.Sizeof(gl.FLOAT)) // float32 - gl.FLOAT, float64 - gl.DOUBLE
	gl.BufferData(gl.ARRAY_BUFFER, size, g.positionArray, g.hint)

	// normal
	g.normalBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(g.normalArray) * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, g.normalArray, gl.STATIC_DRAW)

	// uv
	g.uvBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(g.uvArray) * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, g.uvArray, gl.STATIC_DRAW)

	// color
	g.colorBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(g.colorArray) * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, g.colorArray, gl.STATIC_DRAW)

	// face
	g.faceBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	size = len(g.faceArray) * int(glh.Sizeof(gl.UNSIGNED_SHORT)) // gl.UNSIGNED_SHORT 2, gl.UNSIGNED_INT 4
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, size, g.faceArray, gl.STATIC_DRAW)

	// line
	g.lineBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	size = len(g.lineArray) * int(glh.Sizeof(gl.UNSIGNED_SHORT))
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, size, g.lineArray, gl.STATIC_DRAW)

	g.needsUpdate = false
}

func (g *Geometry) Dispose() {
	if g.positionBuffer != 0 {
		g.positionBuffer.Delete()
	}

	if g.uvBuffer != 0 {
		g.uvBuffer.Delete()
	}

	if g.vertexArrayObject != 0 {
		g.vertexArrayObject.Delete()
	}
}

func (g *Geometry) BindVertexArray() {
	if g.needsUpdate {
		g.update()
	}

	g.vertexArrayObject.Bind()
}

func (g *Geometry) BindPositionBuffer() {
	if g.needsUpdate {
		g.update()
	}

	g.positionBuffer.Bind(gl.ARRAY_BUFFER)
}

func (g *Geometry) BindNormalBuffer() {
	if g.needsUpdate {
		g.update()
	}

	g.normalBuffer.Bind(gl.ARRAY_BUFFER)
}

func (g *Geometry) BindUvBuffer() {
	if g.needsUpdate {
		g.update()
	}

	g.uvBuffer.Bind(gl.ARRAY_BUFFER)
}

func (g *Geometry) BindColorBuffer() {
	if g.needsUpdate {
		g.update()
	}

	g.colorBuffer.Bind(gl.ARRAY_BUFFER)
}

func (g *Geometry) BindLineBuffer() {
	if g.needsUpdate {
		g.update()
	}

	g.lineBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
}

func (g *Geometry) LineCount() int {
	if g.needsUpdate {
		g.update()
	}

	return len(g.lineArray)
}

func (g *Geometry) BindFaceBuffer() {
	if g.needsUpdate {
		g.update()
	}

	g.faceBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
}

func (g *Geometry) FaceCount() int {
	if g.needsUpdate {
		g.update()
	}

	return len(g.faceArray)
}

func (g *Geometry) VerticesCount() int {
	return len(g.vertices)
}
