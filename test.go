package main

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/der-antikeks/mathgl/mgl32"
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
)

func main() {
	// init
	width, height := 800, 400

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	rand.Seed(time.Now().Unix())
	runtime.LockOSThread()

	// setup glfw
	if !glfw.Init() {
		log.Fatal("Can't open GLFW")
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(width, height, "Testing", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.MakeContextCurrent()

	// setup gl
	gl.Init()
	gl.ClearColor(0.1, 0.1, 0.1, 0.0)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.CULL_FACE)

	// setup shader program
	program := LoadShader(`
		#version 330 core

		in vec3 vertexPosition;
		in vec3 vertexNormal;
		in vec2 vertexUV;

		uniform mat4 projectionMatrix;
		uniform mat4 viewMatrix;
		uniform mat4 modelMatrix;
		uniform mat4 modelViewMatrix;
		uniform mat3 normalMatrix;

		out vec2 UV;

		void main() {
			UV = vertexUV;
			gl_Position = projectionMatrix * modelViewMatrix * vec4(vertexPosition, 1.0);
		}
	`, `
		#version 330 core

		in vec2 UV;

		uniform vec3  diffuse;
		uniform float opacity;
		uniform sampler2D diffuseMap;

		out vec4 fragmentColor;

		void main() {
			vec3 materialColor = texture(diffuseMap, UV).rgb;
			fragmentColor = vec4(materialColor * diffuse, opacity);
		}
	`)
	defer program.Delete()

	projectionUniform := program.GetUniformLocation("projectionMatrix")
	modelViewUniform := program.GetUniformLocation("modelViewMatrix")
	diffuseUniform := program.GetUniformLocation("diffuse")
	opacityUniform := program.GetUniformLocation("opacity")
	diffuseMapUniform := program.GetUniformLocation("diffuseMap")

	positionAttribute := program.GetAttribLocation("vertexPosition")
	normalAttribute := program.GetAttribLocation("vertexNormal")
	uvAttribute := program.GetAttribLocation("vertexUV")

	// setup mesh buffers
	mesh := LoadObjFile("assets/fighter/fighter.obj")

	// vao
	vertexArrayObject := gl.GenVertexArray()
	defer vertexArrayObject.Delete()
	vertexArrayObject.Bind()

	// vbo's
	vertexBuffer := gl.GenBuffer()
	defer vertexBuffer.Delete()
	vertexBuffer.Bind(gl.ARRAY_BUFFER)
	size := len(mesh.Vertices) * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, mesh.Vertices, gl.STATIC_DRAW)

	uvBuffer := gl.GenBuffer()
	defer uvBuffer.Delete()
	uvBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(mesh.UVs) * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, mesh.UVs, gl.STATIC_DRAW)

	normalBuffer := gl.GenBuffer()
	defer normalBuffer.Delete()
	normalBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(mesh.Normals) * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, mesh.Normals, gl.STATIC_DRAW)

	// ebo
	elementBuffer := gl.GenBuffer()
	defer elementBuffer.Delete()
	elementBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	size = len(mesh.Indices) * int(glh.Sizeof(gl.UNSIGNED_SHORT))
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, size, mesh.Indices, gl.STATIC_DRAW)

	// setup texture
	textureBuffer := LoadTexture("assets/fighter/fighter.png")
	defer textureBuffer.Delete()

	// main loop
	var angle float32
	for ok := true; ok; ok = (window.GetKey(glfw.KeyEscape) != glfw.Press && !window.ShouldClose()) {
		angle += float32(math.Pi / 5000.0)
		textureSlots := 0

		func() {
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			// use program
			program.Use()
			defer program.Unuse()

			// update uniforms
			projectionMatrix := mgl32.Perspective(45.0, float32(width)/float32(height), 0.1, 200.0)
			projectionUniform.UniformMatrix4fv(false, projectionMatrix)

			viewMatrix := mgl32.LookAtV(mgl32.Vec3{0, 0, -10}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
			modelMatrix := mgl32.HomogRotate3D(angle, (mgl32.Vec3{1, 0.8, 0.5}).Normalize())
			modelViewMatrix := viewMatrix.Mul4(modelMatrix)
			modelViewUniform.UniformMatrix4fv(false, modelViewMatrix)

			diffuseColor := mgl32.Vec3{0.5, 0.8, 1}
			diffuseUniform.Uniform3f(diffuseColor[0], diffuseColor[1], diffuseColor[2])

			opacity := float32(1.0)
			opacityUniform.Uniform1f(opacity)

			// bind texture
			gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(textureSlots))
			textureBuffer.Bind(gl.TEXTURE_2D)
			defer textureBuffer.Unbind(gl.TEXTURE_2D)
			diffuseMapUniform.Uniform1i(textureSlots)
			textureSlots++

			// bind attributes
			vertexArrayObject.Bind()
			defer vertexArrayObject.Unbind()

			vertexBuffer.Bind(gl.ARRAY_BUFFER)
			defer vertexBuffer.Unbind(gl.ARRAY_BUFFER)
			positionAttribute.EnableArray()
			defer positionAttribute.DisableArray()
			positionAttribute.AttribPointer(3, gl.FLOAT, false, 0, nil)

			normalBuffer.Bind(gl.ARRAY_BUFFER)
			defer normalBuffer.Unbind(gl.ARRAY_BUFFER)
			normalAttribute.EnableArray()
			defer normalAttribute.DisableArray()
			normalAttribute.AttribPointer(3, gl.FLOAT, false, 0, nil)

			uvBuffer.Bind(gl.ARRAY_BUFFER)
			defer uvBuffer.Unbind(gl.ARRAY_BUFFER)
			uvAttribute.EnableArray()
			defer uvAttribute.DisableArray()
			uvAttribute.AttribPointer(2, gl.FLOAT, false, 0, nil)

			// draw elements
			elementBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
			defer elementBuffer.Unbind(gl.ELEMENT_ARRAY_BUFFER)
			gl.DrawElements(gl.TRIANGLES, len(mesh.Indices), gl.UNSIGNED_SHORT, nil)

			// Swap buffers
			window.SwapBuffers()
			glfw.PollEvents()
		}()
	}
}

func LoadShaderFile(vertex, fragment string) gl.Program {
	// load vertex shader
	vdata, err := ioutil.ReadFile(vertex)
	if err != nil {
		log.Fatal("unknown vertex file: ", vertex)
	}

	// load fragment shader
	fdata, err := ioutil.ReadFile(fragment)
	if err != nil {
		log.Fatal("unknown fragment file: ", fragment)
	}

	return LoadShader(string(vdata), string(fdata))
}

func LoadShader(vertex, fragment string) gl.Program {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("load shader panic: ", err)
		}
	}()

	program := gl.CreateProgram()

	// vertex shader
	vshader := gl.CreateShader(gl.VERTEX_SHADER)
	vshader.Source(vertex)
	vshader.Compile()
	if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		log.Fatalf("vertex shader error: %v", vshader.GetInfoLog())
	}
	defer vshader.Delete()

	// fragment shader
	fshader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fshader.Source(fragment)
	fshader.Compile()
	if fshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		log.Fatalf("fragment shader error: %v", fshader.GetInfoLog())
	}
	defer fshader.Delete()

	// program
	program.AttachShader(vshader)
	program.AttachShader(fshader)
	program.Link()
	if program.Get(gl.LINK_STATUS) != gl.TRUE {
		log.Fatalf("linker error: %v", program.GetInfoLog())
	}

	return program
}

type Mesh struct {
	Indices  []uint16
	Vertices []float32 //mgl32.Vec3
	UVs      []float32 //mgl32.Vec2
	Normals  []float32 //mgl32.Vec3
}

type Vertex struct {
	position mgl32.Vec3
	uv       mgl32.Vec2
	normal   mgl32.Vec3
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

func LoadObjFile(path string) *Mesh {
	// open object file, init reader
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	// cache
	var cache struct {
		vertices []mgl32.Vec3
		uvs      []mgl32.Vec2
		normals  []mgl32.Vec3
	}

	var result struct {
		Vertices []Vertex
		Faces    []Face
	}

	// helpers
	mustFloat32 := func(v string) float32 {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			log.Fatal(err)
		}
		return float32(f)
	}
	mustUint64 := func(v string) uint64 {
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		return u
	}

	// parse obj file
	for {
		if line, err := reader.ReadString('\n'); err == nil {
			fields := strings.Split(strings.TrimSpace(line), " ")

			switch strings.ToLower(fields[0]) {
			case "v": // geometric vertices: x, y, z, [w]
				cache.vertices = append(cache.vertices, mgl32.Vec3{
					mustFloat32(fields[1]),
					mustFloat32(fields[2]),
					mustFloat32(fields[3]),
				})

			case "vt": // texture vertices: u, v, [w]
				cache.uvs = append(cache.uvs, mgl32.Vec2{
					mustFloat32(fields[1]),
					1.0 - mustFloat32(fields[2]),
				})

			case "vn": // vertex normals: i, j, k
				cache.normals = append(cache.normals, mgl32.Vec3{
					mustFloat32(fields[1]),
					mustFloat32(fields[2]),
					mustFloat32(fields[3]),
				})

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

					// v/vt/vn
					for i, f := range fields[1:4] {
						a := strings.Split(f, "/")

						// vertex
						face[i].position = cache.vertices[mustUint64(a[0])-1]

						// uv
						if len(a) > 1 && a[1] != "" {
							face[i].uv = cache.uvs[mustUint64(a[1])-1]
						}

						// normal
						if len(a) == 3 {
							face[i].normal = cache.normals[mustUint64(a[2])-1]
						}
					}

					offset := len(result.Vertices)
					result.Vertices = append(result.Vertices, face[0], face[1], face[2])
					result.Faces = append(result.Faces, Face{offset, offset + 1, offset + 2})
				}
			default:
				// ignore
			}
		} else if err == io.EOF {
			break
		} else {
			log.Fatal(err)
		}
	}

	// search and mark duplicate vertices
	lookup := map[string]int{}
	unique := []Vertex{}
	changed := map[int]int{}

	for i, v := range result.Vertices {
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

	for _, f := range result.Faces {
		a, b, c := changed[f.A], changed[f.B], changed[f.C]
		if a == b || b == c || c == a {
			// degenerated face, remove
			continue
		}

		nf := Face{a, b, c}
		cleaned = append(cleaned, nf)
	}

	// replace with cleaned
	result.Vertices = unique
	result.Faces = cleaned

	// copy values to buffers
	n := len(result.Vertices)
	m := Mesh{
		Indices:  make([]uint16, len(result.Faces)*3),
		Vertices: make([]float32, n*3),
		UVs:      make([]float32, n*2),
		Normals:  make([]float32, n*3),
	}

	for i, v := range result.Vertices {
		// position
		m.Vertices[i*3] = float32(v.position[0])
		m.Vertices[i*3+1] = float32(v.position[1])
		m.Vertices[i*3+2] = float32(v.position[2])

		// uv
		m.UVs[i*2] = float32(v.uv[0])
		m.UVs[i*2+1] = float32(v.uv[1])

		// normal
		m.Normals[i*3] = float32(v.normal[0])
		m.Normals[i*3+1] = float32(v.normal[1])
		m.Normals[i*3+2] = float32(v.normal[2])
	}

	for i, f := range result.Faces {
		m.Indices[i*3] = uint16(f.A)
		m.Indices[i*3+1] = uint16(f.B)
		m.Indices[i*3+2] = uint16(f.C)
	}

	return &m
}

func LoadTexture(path string) gl.Texture {
	// load file
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// decode image
	im, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	// convert to rgba
	rgba, ok := im.(*image.RGBA)
	if !ok {
		bounds := im.Bounds()
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, im, image.Pt(0, 0), draw.Src)
	}

	buffer := gl.GenTexture()
	buffer.Bind(gl.TEXTURE_2D)

	// set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // gl.REPEAT
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE) // gl.REPEAT
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // gl.LINEAR_MIPMAP_LINEAR
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// give image(s) to opengl
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
		rgba.Bounds().Dx(), rgba.Bounds().Dy(),
		0, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	// generate mipmaps
	gl.GenerateMipmap(gl.TEXTURE_2D)

	buffer.Unbind(gl.TEXTURE_2D)

	return buffer
}
