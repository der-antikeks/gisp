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

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
	"github.com/go-gl/mathgl/mgl32"
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

	glfw.WindowHint(glfw.Resizable, 0)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(width, height, "Testing", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	// setup gl
	gl.Init()
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.ClearDepth(1)
	gl.ClearStencil(0)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	gl.Enable(gl.CULL_FACE)
	gl.FrontFace(gl.CCW)
	gl.CullFace(gl.BACK)

	gl.ShadeModel(gl.SMOOTH)
	gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)

	gl.Enable(gl.BLEND)
	gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

	// first pass, computation program
	// vertex, fragment (!)
	computationProgram := LoadShader(`
		#version 330 core

		layout(location = 0) in vec3 vertexPosition;
		layout(location = 1) in vec3 vertexNormal;
		layout(location = 2) in vec2 vertexUV;

		uniform mat4 projectionMatrix;
		uniform mat4 viewMatrix;
		uniform mat4 modelMatrix;
		uniform mat3 normalMatrix;

		out vec4 Position;
		out vec3 Normal;
		out vec2 UV;

		void main() {
			Position = viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
			Normal = normalize((viewMatrix * modelMatrix * vec4(vertexNormal, 0.0)).xyz);
			UV = vertexUV;

			gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
		}
	`, `
		#version 330 core

		in vec4 Position;
		in vec3 Normal;
		in vec2 UV;

		uniform vec3 diffuse;
		uniform float opacity;
		uniform sampler2D diffuseMap;

		layout(location = 0) out vec4 fragmentColor;
		layout(location = 1) out vec4 fragmentPosition;
		layout(location = 2) out vec4 fragmentNormal;

		void main() {
			vec3 materialColor = texture(diffuseMap, UV).rgb;

			fragmentColor = vec4(materialColor * diffuse, opacity);
			fragmentPosition = vec4(Position.xyz, 0);
			fragmentNormal = vec4(Normal.xyz, 0);
		}
	`)
	defer computationProgram.Delete()

	compProjectionUniform := computationProgram.GetUniformLocation("projectionMatrix")
	compViewUniform := computationProgram.GetUniformLocation("viewMatrix")
	compModelUniform := computationProgram.GetUniformLocation("modelMatrix")
	compNormalUniform := computationProgram.GetUniformLocation("normalMatrix")
	compDiffuseUniform := computationProgram.GetUniformLocation("diffuse")
	compOpacityUniform := computationProgram.GetUniformLocation("opacity")
	compDiffuseMapUniform := computationProgram.GetUniformLocation("diffuseMap")

	compPositionAttribute := computationProgram.GetAttribLocation("vertexPosition")
	compNormalAttribute := computationProgram.GetAttribLocation("vertexNormal")
	compUvAttribute := computationProgram.GetAttribLocation("vertexUV")

	// setup mesh buffers
	cubeMesh := NewMeshBuffer(LoadObjFile("assets/cube/cube.obj"))
	defer cubeMesh.Delete()

	// setup texture
	imageTexture := LoadTexture("assets/uvtemplate.png")
	defer imageTexture.Delete()

	// second pass, deferred shading program
	// vertex, fragment (!)
	deferredProgram := LoadShader(`
		#version 330 core

		layout(location = 0) in vec3 vertexPosition;
		layout(location = 1) in vec2 vertexUV;

		uniform mat4 projectionMatrix;
		uniform mat4 viewMatrix;
		uniform mat4 modelMatrix;

		out vec2 UV;

		void main() {
			UV = vertexUV;
			gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
		}
	`, `
		#version 330 core

		in vec2 UV;

		uniform vec3 cameraPosition;
		uniform sampler2D colorMap;
		uniform sampler2D positionMap;
		uniform sampler2D normalMap;

		layout(location = 0) out vec4 fragmentColor;

		void main() {
			vec4 color = texture(colorMap, UV);
			vec3 position = texture(positionMap, UV).xyz;
			vec3 normal = normalize(texture(normalMap, UV).xyz);

			vec3 lightColor = vec3(1, 1, 1);
			vec3 lightPosition = vec3(8, 2, 0);

			vec3 lightDir = normalize(lightPosition - position);

			// ambient, simulates indirect lighting
			vec3 amb = lightColor * vec3(0.1, 0.1, 0.1);

			// diffuse, direct lightning
			float cosTheta = clamp(dot(normal, lightDir), 0.0, 1.0);
			//float cosTheta = max(dot(normal, lightDir), 0.0);
			vec3 diff = lightColor * cosTheta;

			// specular, reflective highlight, like a mirror
			float cosAlpha = clamp(dot(normalize(-position), reflect(-lightDir, normal)), 0.0, 1.0);
			//float cosAlpha = max(dot(normal, normalize(lightDir + normalize(cameraPosition - position))), 0.0);
			vec3 spec = vec3(0.3, 0.3, 0.3) * lightColor * pow(cosAlpha, 5.0);

			fragmentColor = vec4(color.rgb * (amb + diff + spec), color.a);
			//fragmentColor = vec4(color.rgb, color.a);
		}
	`)
	defer deferredProgram.Delete()

	defProjectionUniform := deferredProgram.GetUniformLocation("projectionMatrix")
	defViewUniform := deferredProgram.GetUniformLocation("viewMatrix")
	defModelUniform := deferredProgram.GetUniformLocation("modelMatrix")
	defCameraPositionUniform := deferredProgram.GetUniformLocation("cameraPosition")
	defColorMapUniform := deferredProgram.GetUniformLocation("colorMap")
	defPositionMapUniform := deferredProgram.GetUniformLocation("positionMap")
	defNormalMapUniform := deferredProgram.GetUniformLocation("normalMap")

	defPositionAttribute := deferredProgram.GetAttribLocation("vertexPosition")
	defUvAttribute := deferredProgram.GetAttribLocation("vertexUV")

	// rendertarget
	tw, th := 1024, 1024
	colorTexture, positionTexture, normalTexture, frameBuffer := GenerateMRT(tw, th)

	planeMesh := NewMeshBuffer(GeneratePlane(5, 5*float32(height)/float32(width)))
	defer planeMesh.Delete()

	// main loop
	var (
		lastTime    = time.Now()
		currentTime time.Time
		delta       time.Duration

		angle float32
	)
	for ok := true; ok; ok = (window.GetKey(glfw.KeyEscape) != glfw.Press && !window.ShouldClose()) {
		currentTime = time.Now()
		delta = currentTime.Sub(lastTime)
		lastTime = currentTime

		angle += float32(math.Pi/8.0) * float32(delta.Seconds())
		textureSlots := 0

		// objects
		var objects []mgl32.Mat4
		var x, y, z, a float32
		for x = -4; x <= 4; x += 4 {
			for z = -4; z <= 4; z += 4 {
				a += math.Pi / 4.0

				o := mgl32.Translate3D(x, y, z).Mul4(mgl32.HomogRotate3D(angle+a, (mgl32.Vec3{1, 0.8, 0.5}).Normalize()))
				objects = append(objects, o)
			}
		}

		// camera
		projectionMatrix := mgl32.Perspective(45.0, float32(width)/float32(height), 1.0, 100.0)
		cameraPosition := mgl32.Vec3{0, 5, 10}
		viewMatrix := mgl32.LookAtV(cameraPosition, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
		targetViewMatrix := mgl32.LookAtV(mgl32.Vec3{0, 0, 5}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

		// object material
		diffuseColor := mgl32.Vec3{0.5, 0.8, 1}
		opacity := float32(1.0)

		// render to fbo
		func() {
			frameBuffer.Bind()
			defer frameBuffer.Unbind()

			gl.Viewport(0, 0, tw, th)
			gl.ClearColor(0.0, 0.0, 0.0, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			// use program
			computationProgram.Use()
			defer computationProgram.Unuse()

			// update uniforms
			compProjectionUniform.UniformMatrix4fv(false, projectionMatrix)
			compViewUniform.UniformMatrix4fv(false, viewMatrix)

			compDiffuseUniform.Uniform3f(diffuseColor[0], diffuseColor[1], diffuseColor[2])
			compOpacityUniform.Uniform1f(opacity)

			// bind texture
			gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(textureSlots))
			imageTexture.Bind(gl.TEXTURE_2D)
			defer imageTexture.Unbind(gl.TEXTURE_2D)
			compDiffuseMapUniform.Uniform1i(textureSlots)
			textureSlots++

			// bind attributes
			cubeMesh.VAO.Bind()
			defer cubeMesh.VAO.Unbind()

			cubeMesh.Position.Bind(gl.ARRAY_BUFFER)
			defer cubeMesh.Position.Unbind(gl.ARRAY_BUFFER)
			compPositionAttribute.EnableArray()
			defer compPositionAttribute.DisableArray()
			compPositionAttribute.AttribPointer(3, gl.FLOAT, false, 0, nil)

			cubeMesh.Normal.Bind(gl.ARRAY_BUFFER)
			defer cubeMesh.Normal.Unbind(gl.ARRAY_BUFFER)
			compNormalAttribute.EnableArray()
			defer compNormalAttribute.DisableArray()
			compNormalAttribute.AttribPointer(3, gl.FLOAT, false, 0, nil)

			cubeMesh.UV.Bind(gl.ARRAY_BUFFER)
			defer cubeMesh.UV.Unbind(gl.ARRAY_BUFFER)
			compUvAttribute.EnableArray()
			defer compUvAttribute.DisableArray()
			compUvAttribute.AttribPointer(2, gl.FLOAT, false, 0, nil)

			for _, modelMatrix := range objects {
				compModelUniform.UniformMatrix4fv(false, modelMatrix)

				modelViewMatrix := viewMatrix.Mul4(modelMatrix)
				normalMatrix := mgl32.Mat4Normal(modelViewMatrix)
				compNormalUniform.UniformMatrix3fv(false, normalMatrix)

				// draw elements
				cubeMesh.EBO.Bind(gl.ELEMENT_ARRAY_BUFFER)
				defer cubeMesh.EBO.Unbind(gl.ELEMENT_ARRAY_BUFFER)
				gl.DrawElements(gl.TRIANGLES, cubeMesh.Size, gl.UNSIGNED_SHORT, nil)
			}
		}()

		// render to screen
		func() {
			gl.Viewport(0, 0, width, height)
			gl.ClearColor(0.1, 0.4, 0.1, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			// use program
			deferredProgram.Use()
			defer deferredProgram.Unuse()

			// update uniforms
			defProjectionUniform.UniformMatrix4fv(false, projectionMatrix)
			defViewUniform.UniformMatrix4fv(false, targetViewMatrix)
			defCameraPositionUniform.Uniform3f(cameraPosition[0], cameraPosition[1], cameraPosition[2])

			defModelUniform.UniformMatrix4fv(false, mgl32.HomogRotate3D(angle, (mgl32.Vec3{0, 0, 1}).Normalize()))

			// bind textures
			gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(textureSlots))
			colorTexture.Bind(gl.TEXTURE_2D)
			defer colorTexture.Unbind(gl.TEXTURE_2D)
			defColorMapUniform.Uniform1i(textureSlots)
			textureSlots++

			gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(textureSlots))
			positionTexture.Bind(gl.TEXTURE_2D)
			defer positionTexture.Unbind(gl.TEXTURE_2D)
			defPositionMapUniform.Uniform1i(textureSlots)
			textureSlots++

			gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(textureSlots))
			normalTexture.Bind(gl.TEXTURE_2D)
			defer normalTexture.Unbind(gl.TEXTURE_2D)
			defNormalMapUniform.Uniform1i(textureSlots)
			textureSlots++

			// bind attributes
			planeMesh.VAO.Bind()
			defer planeMesh.VAO.Unbind()

			planeMesh.Position.Bind(gl.ARRAY_BUFFER)
			defer planeMesh.Position.Unbind(gl.ARRAY_BUFFER)
			defPositionAttribute.EnableArray()
			defer defPositionAttribute.DisableArray()
			defPositionAttribute.AttribPointer(3, gl.FLOAT, false, 0, nil)

			planeMesh.UV.Bind(gl.ARRAY_BUFFER)
			defer planeMesh.UV.Unbind(gl.ARRAY_BUFFER)
			defUvAttribute.EnableArray()
			defer defUvAttribute.DisableArray()
			defUvAttribute.AttribPointer(2, gl.FLOAT, false, 0, nil)

			// draw elements
			planeMesh.EBO.Bind(gl.ELEMENT_ARRAY_BUFFER)
			defer planeMesh.EBO.Unbind(gl.ELEMENT_ARRAY_BUFFER)
			gl.DrawElements(gl.TRIANGLES, planeMesh.Size, gl.UNSIGNED_SHORT, nil)

		}()

		// Swap buffers
		window.SwapBuffers()
		glfw.PollEvents()
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

type MeshBuffer struct {
	VAO                  gl.VertexArray
	Position, UV, Normal gl.Buffer // replace with dedicated mbattr (target, size, usage, ...)
	EBO                  gl.Buffer
	Size                 int
}

func (mb *MeshBuffer) Delete() {
	mb.VAO.Delete()
	mb.Position.Delete()
	mb.UV.Delete()
	mb.Normal.Delete()
	mb.EBO.Delete()
	mb.Size = 0
}

func NewMeshBuffer(mesh *Mesh) *MeshBuffer {
	mb := &MeshBuffer{
		VAO:      gl.GenVertexArray(),
		Position: gl.GenBuffer(),
		UV:       gl.GenBuffer(),
		Normal:   gl.GenBuffer(),
		EBO:      gl.GenBuffer(),
	}

	mb.VAO.Bind()

	// vbo's
	mb.Position.Bind(gl.ARRAY_BUFFER)
	size := len(mesh.Positions) * 3 * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, mesh.Positions, gl.STATIC_DRAW)

	mb.UV.Bind(gl.ARRAY_BUFFER)
	size = len(mesh.UVs) * 2 * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, mesh.UVs, gl.STATIC_DRAW)

	mb.Normal.Bind(gl.ARRAY_BUFFER)
	size = len(mesh.Normals) * 3 * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, mesh.Normals, gl.STATIC_DRAW)

	// ebo
	mb.EBO.Bind(gl.ELEMENT_ARRAY_BUFFER)
	size = len(mesh.Indices) * int(glh.Sizeof(gl.UNSIGNED_SHORT))
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, size, mesh.Indices, gl.STATIC_DRAW)

	mb.Size = len(mesh.Indices)

	return mb
}

type Mesh struct {
	Indices   []uint16
	Positions []mgl32.Vec3
	UVs       []mgl32.Vec2
	Normals   []mgl32.Vec3
}

type Vertex struct {
	position mgl32.Vec3
	uv       mgl32.Vec2
	normal   mgl32.Vec3
}

func (v Vertex) Key(precision int) string {
	return fmt.Sprintf("%v_%v_%v_%v_%v_%v_%v_%v",
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
	var (
		positions []mgl32.Vec3
		uvs       []mgl32.Vec2
		normals   []mgl32.Vec3
		vertices  []Vertex
		faces     []Face
	)

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
				positions = append(positions, mgl32.Vec3{
					mustFloat32(fields[1]),
					mustFloat32(fields[2]),
					mustFloat32(fields[3]),
				})

			case "vt": // texture vertices: u, v, [w]
				uvs = append(uvs, mgl32.Vec2{
					mustFloat32(fields[1]),
					1.0 - mustFloat32(fields[2]),
				})

			case "vn": // vertex normals: i, j, k
				normals = append(normals, mgl32.Vec3{
					mustFloat32(fields[1]),
					mustFloat32(fields[2]),
					mustFloat32(fields[3]),
				})

			case "f": // face: v/vt/vn v/vt/vn v/vt/vn

				// quad instead of tri, split up
				// f v/vt/vn v/vt/vn v/vt/vn v/vt/vn
				var fcs [][]string
				if len(fields) == 5 {
					fcs = [][]string{
						[]string{"f", fields[1], fields[2], fields[4]},
						[]string{"f", fields[2], fields[3], fields[4]},
					}
				} else {
					fcs = [][]string{fields}
				}

				for _, fields := range fcs {
					face := make([]Vertex, 3)

					// v/vt/vn
					for i, f := range fields[1:4] {
						a := strings.Split(f, "/")

						// vertex
						face[i].position = positions[mustUint64(a[0])-1]

						// uv
						if len(a) > 1 && a[1] != "" {
							face[i].uv = uvs[mustUint64(a[1])-1]
						}

						// normal
						if len(a) == 3 {
							face[i].normal = normals[mustUint64(a[2])-1]
						}
					}

					offset := len(vertices)
					vertices = append(vertices, face...)
					faces = append(faces, Face{offset, offset + 1, offset + 2})
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
	for i, v := range vertices {
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
	for _, f := range faces {
		a, b, c := changed[f.A], changed[f.B], changed[f.C]
		if a == b || b == c || c == a {
			// degenerated face, remove
			continue
		}

		cleaned = append(cleaned, Face{a, b, c})
	}

	// copy values to buffers
	n := len(unique)
	m := Mesh{
		Indices:   make([]uint16, len(cleaned)*3),
		Positions: make([]mgl32.Vec3, n),
		UVs:       make([]mgl32.Vec2, n),
		Normals:   make([]mgl32.Vec3, n),
	}

	for i, v := range unique {
		m.Positions[i] = v.position
		m.UVs[i] = v.uv
		m.Normals[i] = v.normal
	}

	for i, f := range cleaned {
		m.Indices[i*3] = uint16(f.A)
		m.Indices[i*3+1] = uint16(f.B)
		m.Indices[i*3+2] = uint16(f.C)
	}

	return &m
}

func GeneratePlane(width, height float32) *Mesh {
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

	// normals
	n := mgl32.Vec3{0, 0, 1}

	// copy values to buffers
	return &Mesh{
		Indices:   []uint16{0, 1, 2, 2, 3, 0},
		Positions: []mgl32.Vec3{a, b, c, d},
		UVs:       []mgl32.Vec2{tr, tl, bl, br},
		Normals:   []mgl32.Vec3{n, n, n, n},
	}
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

	// give image to opengl
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
		rgba.Bounds().Dx(), rgba.Bounds().Dy(),
		0, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	// generate mipmaps
	gl.GenerateMipmap(gl.TEXTURE_2D)

	buffer.Unbind(gl.TEXTURE_2D)

	return buffer
}

func GenerateMRT(w, h int) (diffuse, position, normal gl.Texture, fbo gl.Framebuffer) {
	// generate framebuffer
	frameBuffer := gl.GenFramebuffer()
	frameBuffer.Bind()
	defer frameBuffer.Unbind()

	// generate diffuse target
	diffuseBuffer := gl.GenRenderbuffer()
	diffuseBuffer.Bind()
	defer diffuseBuffer.Unbind()
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.RGBA, w, h)
	diffuseBuffer.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER)

	diffuseTexture := gl.GenTexture()
	diffuseTexture.Bind(gl.TEXTURE_2D)
	defer diffuseTexture.Unbind(gl.TEXTURE_2D)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, w, h, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, diffuseTexture, 0)

	// generate position target
	positionBuffer := gl.GenRenderbuffer()
	positionBuffer.Bind()
	defer positionBuffer.Unbind()
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.RGBA32F, w, h)
	positionBuffer.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.RENDERBUFFER)

	positionTexture := gl.GenTexture()
	positionTexture.Bind(gl.TEXTURE_2D)
	defer positionTexture.Unbind(gl.TEXTURE_2D)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, w, h, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, positionTexture, 0)

	// generate normal target
	normalBuffer := gl.GenRenderbuffer()
	normalBuffer.Bind()
	defer normalBuffer.Unbind()
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.RGBA16F, w, h)
	normalBuffer.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT2, gl.RENDERBUFFER)

	normalTexture := gl.GenTexture()
	normalTexture.Bind(gl.TEXTURE_2D)
	defer normalTexture.Unbind(gl.TEXTURE_2D)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, w, h, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT2, gl.TEXTURE_2D, normalTexture, 0)

	// generate depth buffer
	depthBuffer := gl.GenRenderbuffer()
	depthBuffer.Bind()
	defer depthBuffer.Unbind()
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, w, h)
	depthBuffer.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER)

	gl.DrawBuffers(3, []gl.GLenum{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1, gl.COLOR_ATTACHMENT2})

	// check
	if e := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); e != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalf("could not initialize framebuffer: %x", e)
	}

	return diffuseTexture, positionTexture, normalTexture, frameBuffer
}
