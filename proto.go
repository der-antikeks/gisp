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

	// setup shader program
	program := LoadShader(`
		#version 330 core

		in vec3 vertexPosition;
		in vec3 vertexNormal;
		in vec2 vertexUV;

		uniform mat4 projectionMatrix;
		uniform mat4 viewMatrix;
		uniform mat4 modelMatrix;

		uniform vec3 lightPosition;
		uniform vec3 lightDiffuse;
		uniform float lightPower;

		uniform vec3 ambientColor;

		uniform mat4 shadowBiasMVP;

		out vec2 UV;
		out vec3 lightColor;
		out vec4 shadowCoord;

		vec3 adsShading(vec4 position, vec3 norm) {
			vec4 lightPosCam = viewMatrix * vec4(lightPosition, 1.0); // cameraspace
			vec3 lightDir = normalize(vec3(lightPosCam - position));
			vec3 viewDir = normalize(-position.xyz);
			vec3 reflectDir = reflect(-lightDir, norm);
			float distance = length(lightPosition - (modelMatrix * vec4(vertexPosition, 1.0)).xyz);

			// ambient, simulates indirect lighting
			vec3 amb = ambientColor * vec3(0.1, 0.1, 0.1);

			// diffuse, direct lightning
			float cosTheta = clamp(dot(norm, lightDir), 0.0, 1.0);
			vec3 diff = lightDiffuse * lightPower * cosTheta / (distance * distance);

			// specular, reflective highlight, like a mirror
			float cosAlpha = clamp(dot(viewDir, reflectDir), 0.0, 1.0);
			vec3 spec = vec3(0.3, 0.3, 0.3) * lightDiffuse * lightPower * pow(cosAlpha, 5.0) / (distance * distance);

			return amb + diff + spec;
		}

		void main() {
			// Get the position and normal in camera space
			vec3 camNorm = normalize((viewMatrix * modelMatrix * vec4(vertexNormal, 0.0)).xyz);
			vec4 camPosition = viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);

			UV = vertexUV;

			// Evaluate the lighting equation
			lightColor = adsShading(camPosition, camNorm);

			// Output position of the vertex, worldspace
			shadowCoord = shadowBiasMVP * vec4(vertexPosition, 1.0);

			// Output position of the vertex, clipspace
			gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
		}
	`, `
		#version 330 core

		in vec2 UV;
		in vec3 lightColor;
		in vec4 shadowCoord;

		uniform vec3 diffuse;
		uniform float opacity;
		uniform sampler2D diffuseMap;

		uniform sampler2D shadowMap;

		out vec4 fragmentColor;

		vec2 poissonDisk[16] = vec2[](
			vec2(-0.94201624, -0.39906216),
			vec2(0.94558609, -0.76890725),
			vec2(-0.094184101, -0.92938870),
			vec2(0.34495938, 0.29387760),
			vec2(-0.91588581, 0.45771432),
			vec2(-0.81544232, -0.87912464),
			vec2(-0.38277543, 0.27676845),
			vec2(0.97484398, 0.75648379),
			vec2(0.44323325, -0.97511554),
			vec2(0.53742981, -0.47373420),
			vec2(-0.26496911, -0.41893023),
			vec2(0.79197514, 0.19090188),
			vec2(-0.24188840, 0.99706507),
			vec2(-0.81409955, 0.91437590),
			vec2(0.19984126, 0.78641367),
			vec2(0.14383161, -0.14100790)
		);

		float random(vec3 seed, int i) {
			vec4 seed4 = vec4(seed, i);
			float dot_product = dot(seed4, vec4(12.9898, 78.233, 45.164, 94.673));
			return fract(sin(dot_product) * 43758.5453);
		}

		void main() {
			vec3 materialColor = texture(diffuseMap, UV).rgb;

			float bias = 0.005;
			float visibility = 1.0;
			for (int i = 0; i < 4; i++){
				//int index = int(16.0 * random(gl_FragCoord.xyy, i)) % 16;
				int index = i;
				if (texture(shadowMap, shadowCoord.xy + poissonDisk[index]/700.0).z < shadowCoord.z - bias) {
					visibility -= 0.2;
				}
			}

			fragmentColor = vec4(visibility * lightColor * materialColor * diffuse, opacity);
		}
	`)
	defer program.Delete()

	projectionUniform := program.GetUniformLocation("projectionMatrix")
	viewUniform := program.GetUniformLocation("viewMatrix")
	modelUniform := program.GetUniformLocation("modelMatrix")
	diffuseUniform := program.GetUniformLocation("diffuse")
	opacityUniform := program.GetUniformLocation("opacity")
	diffuseMapUniform := program.GetUniformLocation("diffuseMap")
	lightPositionUniform := program.GetUniformLocation("lightPosition")
	lightDiffuseUniform := program.GetUniformLocation("lightDiffuse")
	lightPowerUniform := program.GetUniformLocation("lightPower")
	ambientColorUniform := program.GetUniformLocation("ambientColor")
	shadowBiasMVPUniform := program.GetUniformLocation("shadowBiasMVP")
	shadowMapUniform := program.GetUniformLocation("shadowMap")

	positionAttribute := program.GetAttribLocation("vertexPosition")
	normalAttribute := program.GetAttribLocation("vertexNormal")
	uvAttribute := program.GetAttribLocation("vertexUV")

	// setup mesh buffers
	mesh := LoadObjFile("assets/cube/cube.obj")

	// vao
	vertexArrayObject := gl.GenVertexArray()
	defer vertexArrayObject.Delete()
	vertexArrayObject.Bind()

	// vbo's
	vertexBuffer := gl.GenBuffer()
	defer vertexBuffer.Delete()
	vertexBuffer.Bind(gl.ARRAY_BUFFER)
	size := len(mesh.Positions) * 3 * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, mesh.Positions, gl.STATIC_DRAW)

	uvBuffer := gl.GenBuffer()
	defer uvBuffer.Delete()
	uvBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(mesh.UVs) * 2 * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, mesh.UVs, gl.STATIC_DRAW)

	normalBuffer := gl.GenBuffer()
	defer normalBuffer.Delete()
	normalBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(mesh.Normals) * 3 * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, mesh.Normals, gl.STATIC_DRAW)

	// ebo
	elementBuffer := gl.GenBuffer()
	defer elementBuffer.Delete()
	elementBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	size = len(mesh.Indices) * int(glh.Sizeof(gl.UNSIGNED_SHORT))
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, size, mesh.Indices, gl.STATIC_DRAW)

	// setup texture
	textureBuffer := LoadTexture("assets/uvtemplate.png")
	defer textureBuffer.Delete()

	// shadow map
	sw, sh := 1024, 1024
	depthBuffer, frameBuffer := GenShadowMap(sw, sh)
	defer depthBuffer.Delete()
	defer frameBuffer.Delete()

	shadowProgram := LoadShader(`
		#version 330 core

		in vec3 vertexPosition;

		uniform mat4 projectionMatrix;
		uniform mat4 viewMatrix;
		uniform mat4 modelMatrix;

		void main() {
			gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
		}
	`, `
		#version 330 core

		out float fragmentDepth;

		void main() {
			fragmentDepth = gl_FragCoord.z;
		}
	`)
	defer shadowProgram.Delete()

	shadowProjectionUniform := shadowProgram.GetUniformLocation("projectionMatrix")
	shadowViewUniform := shadowProgram.GetUniformLocation("viewMatrix")
	shadowModelUniform := shadowProgram.GetUniformLocation("modelMatrix")

	shadowPositionAttribute := shadowProgram.GetAttribLocation("vertexPosition")

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
			a += math.Pi / 4.0

			o := mgl32.Translate3D(x, y, z).Mul4(mgl32.HomogRotate3D(angle+a, (mgl32.Vec3{1, 0.8, 0.5}).Normalize()))
			objects = append(objects, o)
		}

		// camera
		projectionMatrix := mgl32.Perspective(45.0, float32(width)/float32(height), 1.0, 100.0)
		viewMatrix := mgl32.LookAtV(mgl32.Vec3{0, 0, 8}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

		// object material
		diffuseColor := mgl32.Vec3{0.5, 0.8, 1}
		opacity := float32(1.0)

		// light
		lightPosition := mgl32.Vec3{10, 2, 0}
		lightInvDir := lightPosition
		lightDiffuse := mgl32.Vec3{1, 1, 1}
		lightPower := float32(50.0)
		ambientColor := mgl32.Vec3{1, 1, 1}

		// shadow
		shadowProjectionMatrix := mgl32.Ortho(-10, 10, -10, 10, -10, 20)
		//shadowProjectionMatrix := mgl32.Perspective(45.0, float32(sw)/float32(sh), 1.0, 100.0)
		shadowViewMatrix := mgl32.LookAtV(lightInvDir, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
		//shadowViewMatrix := mgl32.LookAtV(lightPosition, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
		biasMatrix := mgl32.Mat4{
			0.5, 0.0, 0.0, 0.0,
			0.0, 0.5, 0.0, 0.0,
			0.0, 0.0, 0.5, 0.0,
			0.5, 0.5, 0.5, 1.0,
		}

		// render to shadowmap
		func() {
			frameBuffer.Bind()
			defer frameBuffer.Unbind()
			gl.Viewport(0, 0, sw, sh)
			gl.ClearColor(0.1, 0.1, 0.4, 0.0)
			gl.Clear(gl.DEPTH_BUFFER_BIT)

			// use program
			shadowProgram.Use()
			defer shadowProgram.Unuse()

			// update uniforms
			shadowProjectionUniform.UniformMatrix4fv(false, shadowProjectionMatrix)
			shadowViewUniform.UniformMatrix4fv(false, shadowViewMatrix)

			// bind attributes
			vertexArrayObject.Bind()
			defer vertexArrayObject.Unbind()

			vertexBuffer.Bind(gl.ARRAY_BUFFER)
			defer vertexBuffer.Unbind(gl.ARRAY_BUFFER)
			shadowPositionAttribute.EnableArray()
			defer shadowPositionAttribute.DisableArray()
			shadowPositionAttribute.AttribPointer(3, gl.FLOAT, false, 0, nil)

			for _, modelMatrix := range objects {
				shadowModelUniform.UniformMatrix4fv(false, modelMatrix)

				// draw elements
				elementBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
				defer elementBuffer.Unbind(gl.ELEMENT_ARRAY_BUFFER)
				gl.DrawElements(gl.TRIANGLES, len(mesh.Indices), gl.UNSIGNED_SHORT, nil)
			}
		}()

		// render to screen
		func() {
			gl.Viewport(0, 0, width, height)
			gl.ClearColor(0.1, 0.1, 0.4, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			// use program
			program.Use()
			defer program.Unuse()

			// update uniforms
			projectionUniform.UniformMatrix4fv(false, projectionMatrix)
			viewUniform.UniformMatrix4fv(false, viewMatrix)

			diffuseUniform.Uniform3f(diffuseColor[0], diffuseColor[1], diffuseColor[2])
			opacityUniform.Uniform1f(opacity)

			lightPositionUniform.Uniform3f(lightPosition[0], lightPosition[1], lightPosition[2])
			lightDiffuseUniform.Uniform3f(lightDiffuse[0], lightDiffuse[1], lightDiffuse[2])
			lightPowerUniform.Uniform1f(lightPower)
			ambientColorUniform.Uniform3f(ambientColor[0], ambientColor[1], ambientColor[2])

			// bind texture
			gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(textureSlots))
			textureBuffer.Bind(gl.TEXTURE_2D)
			defer textureBuffer.Unbind(gl.TEXTURE_2D)
			diffuseMapUniform.Uniform1i(textureSlots)
			textureSlots++

			gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(textureSlots))
			depthBuffer.Bind(gl.TEXTURE_2D)
			defer depthBuffer.Unbind(gl.TEXTURE_2D)
			shadowMapUniform.Uniform1i(textureSlots)
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

			for _, modelMatrix := range objects {
				modelUniform.UniformMatrix4fv(false, modelMatrix)

				shadowMVP := shadowProjectionMatrix.Mul4(shadowViewMatrix.Mul4(modelMatrix))
				shadowBiasMVP := biasMatrix.Mul4(shadowMVP)
				shadowBiasMVPUniform.UniformMatrix4fv(false, shadowBiasMVP)

				// draw elements
				elementBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
				defer elementBuffer.Unbind(gl.ELEMENT_ARRAY_BUFFER)
				gl.DrawElements(gl.TRIANGLES, len(mesh.Indices), gl.UNSIGNED_SHORT, nil)
			}
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

/*
type Shader struct {
	program
	[]attrs/location,type
	[]uniforms/location,type,default
}

type MeshBuffer struct {
	vao
	[]vbos/attr target,usage,type->draw name->shader
	ebo/draw-type
}
*/

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

func GenShadowMap(w, h int) (gl.Texture, gl.Framebuffer) {
	// generate depth texture
	depthBuffer := gl.GenTexture()
	depthBuffer.Bind(gl.TEXTURE_2D)
	defer depthBuffer.Unbind(gl.TEXTURE_2D)

	// set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// create storage
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT16,
		w, h,
		0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)

	// generate framebuffer
	frameBuffer := gl.GenFramebuffer()
	frameBuffer.Bind()
	defer frameBuffer.Unbind()

	// configure framebuffer
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, depthBuffer, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	// check
	if e := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); e != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalf("could not initialize framebuffer: %x", e)
	}

	return depthBuffer, frameBuffer
}