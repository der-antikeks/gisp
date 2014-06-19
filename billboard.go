package main

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
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

		out vec2 UV;

		void main() {
			vec3 cameraRight = vec3(viewMatrix[0][0], viewMatrix[1][0], viewMatrix[2][0]);
			vec3 cameraUp = vec3(viewMatrix[0][1], viewMatrix[1][1], viewMatrix[2][1]);
			vec3 center = (modelMatrix * vec4(0, 0, 0, 1)).xyz;

			vec3 vertexPosition_billboard = 
				center
				+ cameraRight * vertexPosition.x
				+ cameraUp * vertexPosition.y;
	
			UV = vertexUV;
			gl_Position = projectionMatrix * viewMatrix * vec4(vertexPosition_billboard, 1.0);
		}
	`, `
		#version 330 core

		in vec2 UV;

		uniform vec3 diffuse;
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
	viewUniform := program.GetUniformLocation("viewMatrix")
	modelUniform := program.GetUniformLocation("modelMatrix")
	diffuseUniform := program.GetUniformLocation("diffuse")
	opacityUniform := program.GetUniformLocation("opacity")
	diffuseMapUniform := program.GetUniformLocation("diffuseMap")

	positionAttribute := program.GetAttribLocation("vertexPosition")
	normalAttribute := program.GetAttribLocation("vertexNormal")
	uvAttribute := program.GetAttribLocation("vertexUV")

	// setup mesh buffers
	mesh := GeneratePlane(2, 2)

	// vao
	vertexArrayObject := gl.GenVertexArray()
	defer vertexArrayObject.Delete()
	vertexArrayObject.Bind()

	// vbo's
	vertexBuffer := gl.GenBuffer()
	defer vertexBuffer.Delete()
	vertexBuffer.Bind(gl.ARRAY_BUFFER)
	size := len(mesh.Vertices) * 3 * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, mesh.Vertices, gl.STATIC_DRAW)

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

	// objects
	var objects []mgl32.Mat4
	var x, y, z float32
	for x = -6; x <= 6; x += 4 {
		objects = append(objects, mgl32.Translate3D(x, y, z))
	}

	// camera
	cameraRotation := mgl32.QuatLookAtV(mgl32.Vec3{0, 0, 10}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cameraSpeed := float32(math.Pi / 1000.0)

	// main loop
	for ok := true; ok; ok = (window.GetKey(glfw.KeyEscape) != glfw.Press && !window.ShouldClose()) {
		textureSlots := 0

		// camera
		projectionMatrix := mgl32.Perspective(45.0, float32(width)/float32(height), 1.0, 100.0)

		cameraRotation = cameraRotation.Mul(mgl32.AnglesToQuat(cameraSpeed, cameraSpeed/10.0, 0, mgl32.XYZ).Inverse())
		position := cameraRotation.Rotate(mgl32.Vec3{0, 0, 10.0})
		viewMatrix := mgl32.Translate3D(position[0], position[1], position[2]).Mul4(cameraRotation.Mat4()).Inv()
		//viewMatrix := mgl32.LookAtV(mgl32.Vec3{0, 0, 10}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

		// object material
		diffuseColor := mgl32.Vec3{0.5, 0.8, 1}
		opacity := float32(1.0)

		// render to screen
		func() {
			gl.Viewport(0, 0, width, height)
			gl.ClearColor(0.1, 0.1, 0.4, 0.0)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			// use program
			program.Use()
			defer program.Unuse()

			// update uniforms
			projectionUniform.UniformMatrix4fv(false, projectionMatrix)
			viewUniform.UniformMatrix4fv(false, viewMatrix)

			diffuseUniform.Uniform3f(diffuseColor[0], diffuseColor[1], diffuseColor[2])
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

			for _, modelMatrix := range objects {
				modelUniform.UniformMatrix4fv(false, modelMatrix)

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

type Mesh struct {
	Indices  []uint16
	Vertices []mgl32.Vec3
	UVs      []mgl32.Vec2
	Normals  []mgl32.Vec3
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
		Indices:  []uint16{0, 1, 2, 2, 3, 0},
		Vertices: []mgl32.Vec3{a, b, c, d},
		UVs:      []mgl32.Vec2{tr, tl, bl, br},
		Normals:  []mgl32.Vec3{n, n, n, n},
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
