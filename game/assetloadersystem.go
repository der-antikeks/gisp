package game

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/der-antikeks/mathgl/mgl32"

	"github.com/go-gl/gl"
	"github.com/go-gl/glh"

	// TODO: freetype license or outsource as a separate sdf generator
	"code.google.com/p/freetype-go/freetype"
)

/*
	geometry, material, texture, shader

	LoadGeometry(name)
*/
type assetLoaderSystem struct {
	lock sync.Mutex
	path string

	meshbuffers    map[string]*meshbuffer
	shaderPrograms map[string]*shaderprogram
	textures       map[string]*Texture
	fonts          map[string]*Font
}

var (
	assetInstance *assetLoaderSystem
	assetOnce     sync.Once
)

type AssetOpts struct {
	Path string
}

func AssetLoaderSystem(opts *AssetOpts) *assetLoaderSystem {
	assetOnce.Do(func() {
		if opts == nil {
			log.Fatal("zero options init of system")
		}

		assetInstance = &assetLoaderSystem{
			path: opts.Path,

			meshbuffers:    map[string]*meshbuffer{},
			shaderPrograms: map[string]*shaderprogram{},
			textures:       map[string]*Texture{},
			fonts:          map[string]*Font{},
		}
	})

	return assetInstance
}

func (ls *assetLoaderSystem) LoadOBJ(name string) (*meshbuffer, Boundary) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	if mb, found := ls.meshbuffers[name]; found {
		return mb, mb.Bounding
	}

	// open object file, init reader
	path := ls.path + "/" + name

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	// starting mesh
	mb := &meshbuffer{}

	// cache
	var (
		vertices []mgl32.Vec3
		normals  []mgl32.Vec3
		uvs      []mgl32.Vec2
	)

	/*
		wavefront obj/mtl importer
		http://en.wikipedia.org/wiki/Wavefront_OBJ

		object format
		http://paulbourke.net/dataformats/obj/

		material format
		http://paulbourke.net/dataformats/mtl/
	*/

	for {
		if line, err := reader.ReadString('\n'); err == nil {
			fields := strings.Split(strings.TrimSpace(line), " ")

			switch strings.ToLower(fields[0]) {

			// Vertex data

			case "v": // geometric vertices: x, y, z, [w]
				x, err := strconv.ParseFloat(fields[1], 32)
				if err != nil {
					log.Fatal(err)
				}

				y, err := strconv.ParseFloat(fields[2], 32)
				if err != nil {
					log.Fatal(err)
				}

				z, err := strconv.ParseFloat(fields[3], 32)
				if err != nil {
					log.Fatal(err)
				}

				vertices = append(vertices, mgl32.Vec3{float32(x), float32(y), float32(z)})

			case "vt": // texture vertices: u, v, [w]
				u, err := strconv.ParseFloat(fields[1], 32)
				if err != nil {
					log.Fatal(err)
				}

				v, err := strconv.ParseFloat(fields[2], 32)
				if err != nil {
					log.Fatal(err)
				}

				uvs = append(uvs, mgl32.Vec2{float32(u), 1.0 - float32(v)})

			case "vn": // vertex normals: i, j, k
				x, err := strconv.ParseFloat(fields[1], 32)
				if err != nil {
					log.Fatal(err)
				}

				y, err := strconv.ParseFloat(fields[2], 32)
				if err != nil {
					log.Fatal(err)
				}

				z, err := strconv.ParseFloat(fields[3], 32)
				if err != nil {
					log.Fatal(err)
				}

				normals = append(normals, mgl32.Vec3{float32(x), float32(y), float32(z)})

			case "vp": // parameter space vertices
			case "cstype": // curve or surface type
			case "deg": // degree
			case "bmat": // basis matrix
			case "step": // step size

			// Elements

			case "f": // face: v/vt/vn v/vt/vn v/vt/vn

				// quad instead of tri, split up
				// f v/vt/vn v/vt/vn v/vt/vn v/vt/vn
				var faces [][]string
				if len(fields) == 5 {
					faces = [][]string{
						[]string{"f", fields[1], fields[2], fields[4]},
						[]string{"f", fields[2], fields[3], fields[4]},
					}
				} else {
					faces = [][]string{fields}
				}

				for _, fields := range faces {
					var face [3]Vertex
					var v uint64

					for i, f := range fields[1:4] {
						a := strings.Split(f, "/")

						// vertex
						if v, err = strconv.ParseUint(a[0], 10, 64); err != nil {
							log.Fatal(err)
						}
						face[i].position = vertices[v-1]

						// uv
						if len(a) > 1 && a[1] != "" {
							if v, err = strconv.ParseUint(a[1], 10, 64); err != nil {
								log.Fatal(err)
							}
							face[i].uv = uvs[v-1]
						}

						// normal
						if len(a) == 3 {
							if v, err = strconv.ParseUint(a[2], 10, 64); err != nil {
								log.Fatal(err)
							}
							face[i].normal = normals[v-1]
						}
					}

					mb.AddFace(face[0], face[1], face[2])
				}

			case "p": // point
			case "l": // line
			case "curv": // curve
			case "curv2": // 2d curve
			case "surf": // surface

			// Free-form curve/surface body statements

			case "parm", "trim", "hole", "scrv", "sp", "end", "con":

			// Grouping

			case "g": // group name
			case "s": // smoothing group
			case "mg": // merging group
			case "o": // object name

			// Display/render attributes

			case "usemtl": // material name
			case "mtllib": // material library
			case "bevel", "c_interp", "d_interp", "lod",
				"shadow_obj", "trace_obj", "ctech", "stech":

			case "#": // comment
			case "": // empty line
			default:
				log.Fatalf("unknown object line type: %s", line)
			}
		} else if err == io.EOF {
			break
		} else {
			log.Fatal(err)
		}
	}

	mb.MergeVertices()
	mb.ComputeBoundary()
	mb.FaceCount = len(mb.Faces)
	GlContextSystem(nil).MainThread(func() {
		mb.Init()
	})

	ls.meshbuffers[name] = mb
	return mb, mb.Bounding
}

func (ls *assetLoaderSystem) SpherePrimitive(radius float64, widthSegments, heightSegments int) (*meshbuffer, Boundary) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	name := fmt.Sprintf("%s_%g_%d_%d", "sphere", radius, widthSegments, heightSegments)
	if mb, found := ls.meshbuffers[name]; found {
		return mb, mb.Bounding
	}

	mb := &meshbuffer{}

	// dimensions
	if widthSegments < 3 {
		widthSegments = 3
	}
	if heightSegments < 2 {
		heightSegments = 2
	}

	phiStart, phiLength := 0.0, math.Pi*2
	thetaStart, thetaLength := 0.0, math.Pi

	var vertices [][]mgl32.Vec3
	var uvs [][]mgl32.Vec2

	for y := 0; y <= heightSegments; y++ {
		var verticesRow []mgl32.Vec3
		var uvsRow []mgl32.Vec2

		for x := 0; x <= widthSegments; x++ {
			u := float32(x) / float32(widthSegments)
			v := float32(y) / float32(heightSegments)

			vertex := mgl32.Vec3{
				float32(-radius * math.Cos(phiStart+float64(u)*phiLength) * math.Sin(thetaStart+float64(v)*thetaLength)),
				float32(radius * math.Cos(thetaStart+float64(v)*thetaLength)),
				float32(radius * math.Sin(phiStart+float64(u)*phiLength) * math.Sin(thetaStart+float64(v)*thetaLength)),
			}

			verticesRow = append(verticesRow, vertex)
			uvsRow = append(uvsRow, mgl32.Vec2{u, 1.0 - v})
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

			if math.Abs(float64(v1[1])) == radius {
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
			} else if math.Abs(float64(v3[1])) == radius {
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

	mb.MergeVertices()
	mb.ComputeBoundary()
	mb.FaceCount = len(mb.Faces)
	GlContextSystem(nil).MainThread(func() {
		mb.Init()
	})

	ls.meshbuffers[name] = mb
	return mb, mb.Bounding
}

func (ls *assetLoaderSystem) CubePrimitive(size float32) (*meshbuffer, Boundary) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	name := fmt.Sprintf("%s_%g", "cube", size)
	if mb, found := ls.meshbuffers[name]; found {
		return mb, mb.Bounding
	}

	mb := &meshbuffer{}

	// dimensions
	halfSize := size / 2.0

	// vertices
	a := mgl32.Vec3{halfSize, halfSize, halfSize}
	b := mgl32.Vec3{-halfSize, halfSize, halfSize}
	c := mgl32.Vec3{-halfSize, -halfSize, halfSize}
	d := mgl32.Vec3{halfSize, -halfSize, halfSize}
	e := mgl32.Vec3{halfSize, halfSize, -halfSize}
	f := mgl32.Vec3{halfSize, -halfSize, -halfSize}
	g := mgl32.Vec3{-halfSize, -halfSize, -halfSize}
	h := mgl32.Vec3{-halfSize, halfSize, -halfSize}

	// uvs
	tl := mgl32.Vec2{0, 1}
	tr := mgl32.Vec2{1, 1}
	bl := mgl32.Vec2{0, 0}
	br := mgl32.Vec2{1, 0}

	var normal mgl32.Vec3

	// front
	normal = mgl32.Vec3{0, 0, 1}
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
	normal = mgl32.Vec3{0, 0, -1}
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
	normal = mgl32.Vec3{0, 1, 0}
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
	normal = mgl32.Vec3{0, -1, 0}
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
	normal = mgl32.Vec3{-1, 0, 0}
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
	normal = mgl32.Vec3{1, 0, 0}
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

	mb.MergeVertices()
	mb.ComputeBoundary()
	mb.FaceCount = len(mb.Faces)
	GlContextSystem(nil).MainThread(func() {
		mb.Init()
	})

	ls.meshbuffers[name] = mb
	return mb, mb.Bounding
}

func (ls *assetLoaderSystem) PlanePrimitive(width, height float32) (*meshbuffer, Boundary) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	name := fmt.Sprintf("%s_%g_%g", "plane", width, height)
	if mb, found := ls.meshbuffers[name]; found {
		return mb, mb.Bounding
	}

	mb := &meshbuffer{}

	// dimensions
	halfWidth := width / 2.0
	halfHeight := height / 2.0

	// vertices
	a := mgl32.Vec3{halfWidth, halfHeight, 0}
	b := mgl32.Vec3{-halfWidth, halfHeight, 0}
	c := mgl32.Vec3{-halfWidth, -halfHeight, 0}
	d := mgl32.Vec3{halfWidth, -halfHeight, 0}

	// uvs
	tl := mgl32.Vec2{0, 1}
	tr := mgl32.Vec2{1, 1}
	bl := mgl32.Vec2{0, 0}
	br := mgl32.Vec2{1, 0}

	normal := mgl32.Vec3{0, 0, 1}

	mb.AddFace(
		Vertex{
			position: a,
			normal:   normal,
			uv:       tr,
		}, Vertex{
			position: b,
			normal:   normal,
			uv:       tl,
		}, Vertex{
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

	mb.MergeVertices()
	mb.ComputeBoundary()
	mb.FaceCount = len(mb.Faces)
	GlContextSystem(nil).MainThread(func() {
		mb.Init()
	})

	ls.meshbuffers[name] = mb
	return mb, mb.Bounding
}

type Vertex struct {
	position mgl32.Vec3
	normal   mgl32.Vec3
	uv       mgl32.Vec2
}

func (v Vertex) Key(precision int) string {
	return fmt.Sprintf("%v_%v_%v_%v_%v_%v_%v_%v_%v_%v_%v",
		mgl32.Round(v.position[0], precision),
		mgl32.Round(v.position[1], precision),
		mgl32.Round(v.position[2], precision),

		mgl32.Round(v.normal[0], precision),
		mgl32.Round(v.normal[1], precision),
		mgl32.Round(v.normal[2], precision),

		mgl32.Round(v.uv[0], precision),
		mgl32.Round(v.uv[1], precision),
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

	Bounding  Boundary
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
	g.Bounding = NewBoundary()
	for _, v := range g.Vertices {
		g.Bounding.AddPoint(v.position)
	}
}

func (g *meshbuffer) Init() {
	if g.initialized {
		return
	}

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

	g.Vertices = nil
	g.Faces = nil
	g.initialized = true
}

func (g *meshbuffer) Cleanup() {
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

type shaderprogram struct {
	program gl.Program
	enabled bool

	uniforms map[string]struct {
		location gl.UniformLocation
		standard interface{}
	}
	attributes map[string]struct {
		location gl.AttribLocation
		size     uint
		enabled  bool
	}
}

func (ls *assetLoaderSystem) LoadShader(name string) *shaderprogram {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	if s, found := ls.shaderPrograms[name]; found {
		return s
	}

	path := ls.path + "/shaders/" + name

	// load vertex shader
	vdata, err := ioutil.ReadFile(path + ".vertex")
	if err != nil {
		log.Fatal("unknown shader program: ", name)
	}

	// load fragment shader
	fdata, err := ioutil.ReadFile(path + ".fragment")
	if err != nil {
		log.Fatal("unknown shader program: ", name)
	}

	// standard shader uniforms and attributes
	uniforms := map[string]interface{}{
		"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
		"viewMatrix":       nil, //[16]float32{},
		"modelMatrix":      nil, //[16]float32{},
		"modelViewMatrix":  nil, //[16]float32{},
		"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()
	}
	attributes := map[string]uint{
		"vertexPosition": 3,
		"vertexNormal":   3,
		"vertexUV":       2,
	}

	// material specific uniforms
	switch name {
	default:
		log.Fatal("unknown shader program: ", name)
	case "basic":
		uniforms["opacity"] = 1.0
		uniforms["diffuse"] = mgl32.Vec3{1, 1, 1}

	case "phong":
		uniforms["diffuseMap"] = nil // *Texture
		uniforms["opacity"] = 1.0
		uniforms["diffuse"] = mgl32.Vec3{1, 1, 1}

		uniforms["ambient"] = mgl32.Vec3{1, 1, 1}
		uniforms["emissive"] = mgl32.Vec3{1, 1, 1}
		uniforms["specular"] = mgl32.Vec3{1, 1, 1}
	case "flat":
		uniforms["lightDiffuse"] = mgl32.Vec3{1, 1, 1}
		uniforms["lightPosition"] = mgl32.Vec3{0, 0, 0}
		uniforms["lightPower"] = 50.0

		uniforms["ambientColor"] = mgl32.Vec3{1, 1, 1}

		uniforms["diffuseMap"] = nil // *Texture
		uniforms["opacity"] = 1.0

	case "billboard":
		uniforms["opacity"] = 1.0
		uniforms["diffuse"] = mgl32.Vec3{1, 1, 1}

		uniforms["diffuseMap"] = nil // *Texture

	case "fixedbb":
		uniforms["opacity"] = 1.0
		uniforms["diffuse"] = mgl32.Vec3{1, 1, 1}
		uniforms["diffuseMap"] = nil // *Texture

		uniforms["size"] = mgl32.Vec2{1, 1}

	case "sdf":
		uniforms["distanceFieldMap"] = nil // *Texture
		uniforms["diffuse"] = mgl32.Vec3{1, 1, 1}
		uniforms["opacity"] = 0.25
	}

	s := &shaderprogram{
		uniforms: map[string]struct {
			location gl.UniformLocation
			standard interface{}
		}{},
		attributes: map[string]struct {
			location gl.AttribLocation
			size     uint
			enabled  bool
		}{},
	}

	GlContextSystem(nil).MainThread(func() {
		s.program = gl.CreateProgram()

		// vertex shader
		vshader := gl.CreateShader(gl.VERTEX_SHADER)
		vshader.Source(string(vdata))
		vshader.Compile()
		if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
			log.Fatalf("vertex shader error: %v", vshader.GetInfoLog())
		}
		defer vshader.Delete()

		// fragment shader
		fshader := gl.CreateShader(gl.FRAGMENT_SHADER)
		fshader.Source(string(fdata))
		fshader.Compile()
		if fshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
			log.Fatalf("fragment shader error: %v", fshader.GetInfoLog())
		}
		defer fshader.Delete()

		// program
		s.program.AttachShader(vshader)
		s.program.AttachShader(fshader)
		s.program.Link()
		if s.program.Get(gl.LINK_STATUS) != gl.TRUE {
			log.Fatalf("linker error: %v", s.program.GetInfoLog())
		}

		defer func() {
			if err := recover(); err != nil {
				log.Fatal("Recovered: ", err)
			}
		}()

		// locations
		for n, v := range uniforms {
			s.uniforms[n] = struct {
				location gl.UniformLocation
				standard interface{}
			}{
				location: s.program.GetUniformLocation(n),
				standard: v,
			}
		}

		for n, v := range attributes {
			s.attributes[n] = struct {
				location gl.AttribLocation
				size     uint
				enabled  bool
			}{
				location: s.program.GetAttribLocation(n),
				size:     v,
				enabled:  false,
			}
		}
	})

	ls.shaderPrograms[name] = s
	return s
}

func (s *shaderprogram) DisableAttributes() {
	for n, a := range s.attributes {
		if a.enabled {
			a.location.DisableArray()
			a.enabled = false
			s.attributes[n] = a
		}
	}
}

func (s *shaderprogram) EnableAttribute(name string) {
	a, ok := s.attributes[name]
	if !ok {
		log.Fatal("unknown attribute: ", name)
	}

	if !a.enabled {
		a.location.EnableArray()
		a.enabled = true
		s.attributes[name] = a
	}

	a.location.AttribPointer(a.size, gl.FLOAT, false, 0, nil)
}

func (s *shaderprogram) Cleanup() {
	// TODO: do something
}

type Texture struct {
	buffer gl.Texture
	w, h   int // TODO: needed?
}

// cleanup
func (t Texture) Cleanup() {
	if t.buffer != 0 {
		t.buffer.Delete()
	}
}

func (ls *assetLoaderSystem) LoadTexture(name string) (*Texture, error) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	if t, found := ls.textures[name]; found {
		return t, nil
	}

	path := ls.path + "/" + name

	// load file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// decode image
	im, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := im.Bounds()

	// convert to rgba
	rgba, ok := im.(*image.RGBA)
	if !ok {
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, im, image.Pt(0, 0), draw.Src)
	}

	// create texture
	t := &Texture{
	//w: bounds.Dx(),
	//h: bounds.Dy(),
	}

	GlContextSystem(nil).MainThread(func() {
		t.buffer = gl.GenTexture()
		t.buffer.Bind(gl.TEXTURE_2D)

		// set texture parameters
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // gl.LINEAR_MIPMAP_LINEAR
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		// give image(s) to opengl
		gl.TexImage2D(gl.TEXTURE_2D, 0 /*level*/, gl.RGBA,
			rgba.Bounds().Dx(), rgba.Bounds().Dy(),
			0, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

		// generate mipmaps
		gl.GenerateMipmap(gl.TEXTURE_2D)

		t.buffer.Unbind(gl.TEXTURE_2D)
	})

	ls.textures[name] = t
	return t, nil
}

type Glyph struct {
	x, y    int
	w, h    int
	advance int
}

type Font struct {
	sdf   *Texture
	w, h  int
	chars map[rune]Glyph
}

// cleanup
func (f *Font) Cleanup() {
	if f.sdf.buffer != 0 {
		f.sdf.buffer.Delete()
	}
}

func NextHighestPowerOfTwo(n int) int {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}

func (ls *assetLoaderSystem) LoadSDFFont(name string, size float64, low, high int) (*Font, error) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	lname := fmt.Sprintf("%s_%g_%d_%d", name, size, low, high)
	if t, found := ls.fonts[lname]; found {
		return t, nil
	}

	path := ls.path + "/" + name
	dpi := 72.0 // screen resolution in dots per inch
	//size := 32.0          // font size in points
	//low, high := 32, 127  // lower, upper rune limits
	spread := 4           // signed distance radius
	padding := spread + 1 // padding between glyphs

	// read font data
	fontBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}

	// create image
	glyphCounts := high - low + 1
	glyphsPerRow := int(math.Ceil(math.Sqrt(float64(glyphCounts))))
	glyphsPerCol := (glyphCounts / glyphsPerRow) + 1

	glyphBounds := font.Bounds(int32(size))
	glyphWidth := int(glyphBounds.XMax-glyphBounds.XMin) + padding*2
	glyphHeight := int(glyphBounds.YMax-glyphBounds.YMin) + padding*2

	rgba := image.NewRGBA(image.Rect(0, 0,
		NextHighestPowerOfTwo(glyphWidth*glyphsPerRow),
		NextHighestPowerOfTwo(glyphHeight*glyphsPerCol)))
	//draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)

	// initialize context
	ctx := freetype.NewContext()
	ctx.SetDPI(dpi)
	ctx.SetFont(font)
	ctx.SetFontSize(size)
	ctx.SetClip(rgba.Bounds())
	ctx.SetDst(rgba)
	ctx.SetSrc(image.White)

	// draw runes
	glyphs := make(map[rune]Glyph)
	var glyphNum, glyphX, glyphY int

	for ch := low; ch <= high; ch++ {
		glyphX = (glyphNum % glyphsPerRow) * glyphWidth
		glyphY = (glyphNum / glyphsPerRow) * glyphHeight

		metric := font.HMetric(int32(size), font.Index(rune(ch)))
		advance := int(metric.AdvanceWidth) + padding

		glyphs[rune(ch)] = Glyph{glyphX, glyphY, glyphWidth, glyphHeight, advance}

		pt := freetype.Pt(glyphX+padding, glyphY+padding+int(ctx.PointToFix32(size)>>8))
		if _, err = ctx.DrawString(string(ch), pt); err != nil {
			return nil, err
		}

		glyphNum++
	}

	// generate signed distance field
	sdf := image.NewRGBA(rgba.Bounds())

	for _, glyph := range glyphs {
		// create mask
		mask := make([][]bool, glyph.h)

		for y := 0; y < glyph.h; y++ {
			if mask[y] == nil {
				mask[y] = make([]bool, glyph.w)
			}

			for x := 0; x < glyph.w; x++ {
				r, g, b, a := rgba.At(glyph.x+x, glyph.y+y).RGBA()
				mask[y][x] = (r >= 0x7fff || g >= 0x7fff || b >= 0x7fff) && (a >= 0x7fff)
			}
		}

		// find signed distance
		width := len(mask[0])
		height := len(mask)

		for centerY := 0; centerY < glyph.h; centerY++ {
			for centerX := 0; centerX < glyph.w; centerX++ {
				base := mask[centerY][centerX]

				startX, endX := int(math.Max(0, float64(centerX-spread))), int(math.Min(float64(centerX+spread), float64(width-1)))
				startY, endY := int(math.Max(0, float64(centerY-spread))), int(math.Min(float64(centerY+spread), float64(height-1)))

				closestSquareDist := spread * spread
				for y := startY; y <= endY; y++ {
					for x := startX; x <= endX; x++ {
						if base != mask[y][x] {
							squareDist := (centerX-x)*(centerX-x) + (centerY-y)*(centerY-y)
							if squareDist < closestSquareDist {
								closestSquareDist = squareDist
							}
						}
					}
				}

				dist := math.Min(math.Sqrt(float64(closestSquareDist)), float64(spread))
				if !base {
					dist *= -1
				}

				// distance to color
				cu := uint8(math.Min(1, math.Max(0, 0.5+0.5*(dist/float64(spread)))) * 0xff)
				c := color.RGBA{cu, cu, cu, cu} // premultiplied alpha white

				sdf.Set(glyph.x+centerX, glyph.y+centerY, c)
			}
		}
	}

	bounds := sdf.Bounds().Size()

	// generate texture
	tex := &Texture{}

	GlContextSystem(nil).MainThread(func() {
		tex.buffer = gl.GenTexture()
		tex.buffer.Bind(gl.TEXTURE_2D)

		// set texture parameters
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR) // gl.LINEAR_MIPMAP_LINEAR
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		// give image(s) to opengl
		gl.TexImage2D(gl.TEXTURE_2D, 0 /*level*/, gl.RGBA,
			sdf.Bounds().Dx(), sdf.Bounds().Dy(),
			0, gl.RGBA, gl.UNSIGNED_BYTE, sdf.Pix)

		// generate mipmaps
		gl.GenerateMipmap(gl.TEXTURE_2D)

		tex.buffer.Unbind(gl.TEXTURE_2D)
	})

	f := &Font{
		sdf:   tex,
		w:     bounds.X,
		h:     bounds.Y,
		chars: glyphs,
	}

	ls.fonts[lname] = f
	return f, nil
}

func (ls *assetLoaderSystem) CreateString(f *Font, s string) (*meshbuffer, Boundary) {
	mb := &meshbuffer{}

	normal := mgl32.Vec3{0, 0, 1}
	var size float32 = 1.0
	var sx float32 = 0.0

	for _, c := range s {
		glyph, found := f.chars[c]
		if !found {
			continue
		}

		x := float32(glyph.x) / float32(f.w)
		y := float32(glyph.y) / float32(f.h)
		w := float32(glyph.advance) / float32(f.w)
		h := float32(glyph.h) / float32(f.h)

		a := mgl32.Vec3{sx + size*w, size * h, 0}
		b := mgl32.Vec3{sx, size * h, 0}
		c := mgl32.Vec3{sx, 0, 0}
		d := mgl32.Vec3{sx + size*w, 0, 0}
		sx += size * w

		tl := mgl32.Vec2{x, y}
		tr := mgl32.Vec2{x + w, y}
		bl := mgl32.Vec2{x, y + h}
		br := mgl32.Vec2{x + w, y + h}

		mb.AddFace(
			Vertex{
				position: a,
				normal:   normal,
				uv:       tr,
			}, Vertex{
				position: b,
				normal:   normal,
				uv:       tl,
			}, Vertex{
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
	}

	mb.MergeVertices()
	mb.ComputeBoundary()
	mb.FaceCount = len(mb.Faces)
	GlContextSystem(nil).MainThread(func() {
		mb.Init()
	})

	return mb, mb.Bounding
}

// TODO: Framebuffers

type Framebuffer struct {
	Color mgl32.Vec3
	Alpha float64
	Clear bool

	buffer *Texture
}

func (ls *assetLoaderSystem) NewFramebuffer(w, h int) *Texture {
	t := &Texture{
		buffer: gl.GenTexture(),
		w:      w,
		h:      h,
	}

	t.buffer.Bind(gl.TEXTURE_2D)
	{
		// set texture parameters
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // gl.NEAREST
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR) // gl.NEAREST

		// create storage
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
			t.w, t.h,
			0, gl.RGBA, gl.UNSIGNED_BYTE, nil)

		// generate mipmaps
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}
	t.buffer.Unbind(gl.TEXTURE_2D)

	return t
}

func (ls *assetLoaderSystem) Cleanup() {
	// TODO: unload textures, buffers, programs and empty caches

	for _, m := range ls.meshbuffers {
		GlContextSystem(nil).MainThread(func() {
			m.Cleanup()
		})
	}

	for _, s := range ls.shaderPrograms {
		GlContextSystem(nil).MainThread(func() {
			s.Cleanup()
		})
	}

	for _, t := range ls.textures {
		GlContextSystem(nil).MainThread(func() {
			t.Cleanup()
		})
	}

	for _, f := range ls.fonts {
		GlContextSystem(nil).MainThread(func() {
			f.Cleanup()
		})
	}
}
