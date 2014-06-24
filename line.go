package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"
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

	gl.LineWidth(2)

	// setup shader program
	program := LoadShader(`
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

		uniform vec3 diffuse;
		uniform float opacity;

		out vec4 fragmentColor;

		void main() {
			fragmentColor = vec4(diffuse, opacity);
		}
	`)
	defer program.Delete()

	projectionUniform := program.GetUniformLocation("projectionMatrix")
	viewUniform := program.GetUniformLocation("viewMatrix")
	modelUniform := program.GetUniformLocation("modelMatrix")
	diffuseUniform := program.GetUniformLocation("diffuse")
	opacityUniform := program.GetUniformLocation("opacity")

	positionAttribute := program.GetAttribLocation("vertexPosition")

	// setup mesh buffers
	indices, positions := GenerateCircle(3.0, 32)

	// vao
	vertexArrayObject := gl.GenVertexArray()
	defer vertexArrayObject.Delete()
	vertexArrayObject.Bind()

	// vbo's
	vertexBuffer := gl.GenBuffer()
	defer vertexBuffer.Delete()
	vertexBuffer.Bind(gl.ARRAY_BUFFER)
	size := len(positions) * 3 * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, positions, gl.STATIC_DRAW)

	// ebo
	elementBuffer := gl.GenBuffer()
	defer elementBuffer.Delete()
	elementBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	size = len(indices) * int(glh.Sizeof(gl.UNSIGNED_SHORT))
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, size, indices, gl.STATIC_DRAW)

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
		diffuseColor := mgl32.Vec3{1, 1, 0}
		opacity := float32(0.8)

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

			// bind attributes
			vertexArrayObject.Bind()
			defer vertexArrayObject.Unbind()

			vertexBuffer.Bind(gl.ARRAY_BUFFER)
			defer vertexBuffer.Unbind(gl.ARRAY_BUFFER)
			positionAttribute.EnableArray()
			defer positionAttribute.DisableArray()
			positionAttribute.AttribPointer(3, gl.FLOAT, false, 0, nil)

			for _, modelMatrix := range objects {
				modelUniform.UniformMatrix4fv(false, modelMatrix)

				// draw elements
				elementBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
				defer elementBuffer.Unbind(gl.ELEMENT_ARRAY_BUFFER)
				gl.DrawElements(gl.LINE_LOOP, len(indices), gl.UNSIGNED_SHORT, nil)
			}
		}()

		// Swap buffers
		window.SwapBuffers()
		glfw.PollEvents()
	}
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

func GenerateCircle(radius float32, segments int) (indices []uint16, positions []mgl32.Vec3) {
	step := math.Pi * 2 / float64(segments)
	for i := 0; i < segments; i++ {
		s, c := math.Sincos(step * float64(i))

		positions = append(positions, mgl32.Vec3{
			float32(c) * radius,
			float32(s) * radius,
			0,
		})
		indices = append(indices, uint16(i))
	}

	return indices, positions
}
