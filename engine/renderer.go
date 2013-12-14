package engine

import (
	"fmt"
	"log"

	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
)

type Renderer struct {
	// window
	title  string
	width  int
	height int
	window *glfw.Window

	// state cache
	currentProgram      *Program // if prg of mat.prg != current, prg.use and reset uniforms
	currentMaterial     Material // if mat != current, reset uniforms
	currentCamera       Camera
	currentGeometry     *Geometry
	currentRendertarget *RenderTarget

	// input callbacks
	resizeCallback      func(w, h float64)
	mouseMoveCallback   func(x, y float64)
	mouseScrollCallback func(x, y float64)
	mouseButtonCallback func(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
	keyCallback         func(key glfw.Key, action glfw.Action, mods glfw.ModifierKey)
	controller          Control

	// renderpasses
	passes []*RenderPass
}

func NewRenderer(title string, width, height int) (*Renderer, error) {
	r := &Renderer{
		title:  title,
		width:  width,
		height: height,
	}

	// initialize glfw
	if err := r.initGLFW(); err != nil {
		return nil, err
	}

	// initialize gl
	if err := r.initGL(); err != nil {
		return nil, err
	}

	return r, nil
}

// initialization

func (r *Renderer) errorCallback(err glfw.ErrorCode, desc string) {
	log.Printf("%v: %v\n", err, desc)
}

func (r *Renderer) initGLFW() error {
	glfw.SetErrorCallback(r.errorCallback)
	if !glfw.Init() {
		return fmt.Errorf("Failed to initialize GLFW")
	}

	// create window
	glfw.WindowHint(glfw.Resizable, 1)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(r.width, r.height, r.title, nil, nil)
	if err != nil {
		return err
	}

	// set callbacks
	window.SetFramebufferSizeCallback(r.onResize)
	window.SetKeyCallback(r.onKey)
	window.SetCursorPositionCallback(r.onMouseMove)
	window.SetInputMode(glfw.Cursor, glfw.CursorNormal /*glfw.CursorDisabled*/)
	window.SetScrollCallback(r.onMouseScroll)
	window.SetMouseButtonCallback(r.onMouseButton)

	// change context to created window
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	w, h := window.GetFramebufferSize()
	r.onResize(window, w, h)

	r.window = window
	return nil
}

func (r *Renderer) initGL() error {
	gl.Init()
	if err := glh.CheckGLError(); err != nil {
		return err
	}

	// clearing
	gl.ClearColor(0, 0, 0, 1.0)
	gl.ClearDepth(1)
	gl.ClearStencil(0)

	// depth
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)

	// cull face
	gl.FrontFace(gl.CCW)
	gl.CullFace(gl.BACK)
	gl.Enable(gl.CULL_FACE)

	gl.LineWidth(2)
	/*
		gl.ShadeModel(gl.SMOOTH)

		gl.Enable(gl.MULTISAMPLE)
		gl.Disable(gl.LIGHTING)

		gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)
		gl.ColorMaterial(gl.FRONT_AND_BACK, gl.AMBIENT_AND_DIFFUSE)
	*/

	return nil
}

// cleanup

func (r *Renderer) Unload() {
	for _, p := range r.passes {
		p.scene.Dispose()
	}

	glfw.Terminate()
}

func (r *Renderer) Running() bool {
	return !r.window.ShouldClose()
}

func (r *Renderer) Quit() {
	r.window.SetShouldClose(true)
}

// set parameters

func (r *Renderer) SetTitle(title string) {
	r.title = title
	r.window.SetTitle(title)
}

func (r *Renderer) SetSize(width, height int) {
	r.width = width
	r.height = height
	r.window.SetSize(width, height)
}

func (r *Renderer) SetClearColor(color math.Color, alpha float64) {
	gl.ClearColor(gl.GLclampf(color.R), gl.GLclampf(color.G), gl.GLclampf(color.B), gl.GLclampf(alpha))
}

func (r *Renderer) SetMouseVisible(show bool) {
	if show {
		r.window.SetInputMode(glfw.Cursor, glfw.CursorNormal)
	} else {
		r.window.SetInputMode(glfw.Cursor, glfw.CursorDisabled)
	}
}

// event callbacks

func (r *Renderer) SetController(c Control) {
	r.controller = c
}

func (r *Renderer) SetResizeCallback(f func(w, h float64)) {
	r.resizeCallback = f
}

func (r *Renderer) onResize(window *glfw.Window, w, h int) {
	if h < 1 {
		h = 1
	}

	if w < 1 {
		w = 1
	}

	r.width = w
	r.height = h

	gl.Viewport(0, 0, w, h)

	if r.resizeCallback != nil {
		r.resizeCallback(float64(w), float64(h))
	}

	if r.controller != nil {
		r.controller.OnWindowResize(float64(w), float64(h))
	}
}

func (r *Renderer) SetKeyCallback(f func(key glfw.Key, action glfw.Action, mods glfw.ModifierKey)) {
	r.keyCallback = f
}

func (r *Renderer) onKey(window *glfw.Window, k glfw.Key, s int, action glfw.Action, mods glfw.ModifierKey) {
	if r.keyCallback != nil {
		r.keyCallback(k, action, mods)
	}

	if r.controller != nil {
		r.controller.OnKeyPress(k, action, mods)
	}
}

func (r *Renderer) SetMouseMoveCallback(f func(x, y float64)) {
	r.mouseMoveCallback = f
}

func (r *Renderer) onMouseMove(window *glfw.Window, xpos float64, ypos float64) {
	if r.mouseMoveCallback != nil {
		r.mouseMoveCallback(xpos, ypos)
	}

	if r.controller != nil {
		r.controller.OnMouseMove(xpos, ypos)
	}
}

func (r *Renderer) SetMouseScrollCallback(f func(x, y float64)) {
	r.mouseScrollCallback = f
}

func (r *Renderer) onMouseScroll(w *glfw.Window, xoff float64, yoff float64) {
	if r.mouseScrollCallback != nil {
		r.mouseScrollCallback(xoff, yoff)
	}

	if r.controller != nil {
		r.controller.OnMouseScroll(xoff, yoff)
	}
}

func (r *Renderer) SetMouseButtonCallback(f func(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)) {
	r.mouseButtonCallback = f
}

func (r *Renderer) onMouseButton(w *glfw.Window, b glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if r.mouseButtonCallback != nil {
		r.mouseButtonCallback(b, action, mods)
	}

	if r.controller != nil {
		r.controller.OnMouseButton(b, action, mods)
	}
}

// rendering

func (r *Renderer) AddPass(p *RenderPass) {
	r.passes = append(r.passes, p)
}

func (r *Renderer) Render() {
	for _, p := range r.passes {

		if p.clear {
			r.SetClearColor(p.clearColor, p.clearAlpha)
		}

		if p.target == nil {
			r.RenderScene(p.scene, p.camera, p.clear, nil)
		} else {
			r.RenderScene(p.scene, p.camera, p.clear, p.target)
		}
	}

	r.SwapBuffers()
}

func (r *Renderer) RenderScene(scene *Scene, camera Camera, clear bool, target *RenderTarget) {
	// bind rendertarget
	if r.currentRendertarget != target {
		if target != nil {
			target.BindFramebuffer()
			gl.Viewport(0, 0, target.width, target.height)
		} else {
			r.currentRendertarget.Unbind()
			gl.Viewport(0, 0, r.width, r.height)
		}

		r.currentRendertarget = target
	}

	// clear screen
	if clear {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	} else {
		// TODO: do i need depth for shader passes?
		//gl.Clear(gl.DEPTH_BUFFER_BIT)
	}

	// update scene graph
	scene.UpdateMatrixWorld(false)

	// update camera matrices and frustum
	if camera.Parent() == nil {
		// was not updated with scene graph
		camera.UpdateMatrixWorld(false)
	}

	projScreenMatrix := camera.ProjectionMatrix().Mul(camera.MatrixWorld().Inverse())
	frustum := math.FrustumFromMatrix(projScreenMatrix)

	// filter visible objects
	opaque, transparent := scene.VisibleObjects(frustum)

	// opaque pass (front-to-back order)
	gl.Disable(gl.BLEND)

	for _, o := range opaque {
		r.renderObject(o, camera)
	}

	// transparent pass (back-to-front order)
	gl.Enable(gl.BLEND)
	gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

	for _, o := range transparent {
		r.renderObject(o, camera)
	}
}

func (r *Renderer) SwapBuffers() {
	// Swap buffers
	r.window.SwapBuffers()
	glfw.PollEvents()
}

func (r *Renderer) renderObject(m Renderable, camera Camera) {
	if _, ok := m.Material().(*ShaderMaterial); ok {
		r.renderObjectNew(m, camera)
		return
	}

	material := m.Material()
	var refreshMaterial bool

	if r.currentMaterial != material {
		if r.currentMaterial != nil {
			r.currentMaterial.Unbind()
		}

		r.currentMaterial = material
		refreshMaterial = true
	}

	// use program
	program := material.Program()
	if r.currentProgram != program {
		program.Use()
		r.currentProgram = program

		// different materials may have different programs
		// same material with different values of same uniforms (diffuse, reflection, ...) have same program
		refreshMaterial = true
	}

	if refreshMaterial || r.currentCamera != camera {
		program.Uniform("projectionMatrix").UniformMatrix4fv(false, camera.ProjectionMatrix().Float32())

		if r.currentCamera != camera {
			r.currentCamera = camera
		}
	}

	if refreshMaterial {
		material.UpdateUniforms()
	}

	geometry := m.Geometry()
	if r.currentGeometry != geometry {
		r.currentGeometry = geometry

		program.DisableAttributes()
		geometry.BindVertexArray()

		// vertices
		geometry.BindPositionBuffer()
		program.EnableAttribute("vertexPosition")
		program.Attribute("vertexPosition").AttribPointer(3, gl.FLOAT, false, 0, nil)
		//geometry.positionBuffer.Unbind(gl.ARRAY_BUFFER)

		// normal
		geometry.BindNormalBuffer()
		program.EnableAttribute("vertexNormal")
		program.Attribute("vertexNormal").AttribPointer(3, gl.FLOAT, false, 0, nil)

		// uv
		geometry.BindUvBuffer()
		program.EnableAttribute("vertexUV")
		program.Attribute("vertexUV").AttribPointer(2, gl.FLOAT, false, 0, nil)

		// color
		geometry.BindColorBuffer()
		program.EnableAttribute("vertexColor")
		program.Attribute("vertexColor").AttribPointer(3, gl.FLOAT, false, 0, nil)
	}

	// for each object of same material and geometry

	// Model matrix : an identity matrix (model will be at the origin)
	program.Uniform("modelMatrix").UniformMatrix4fv(false, m.MatrixWorld().Float32())

	// viewMatrix
	viewMatrix := camera.MatrixWorld().Inverse()
	program.Uniform("viewMatrix").UniformMatrix4fv(false, viewMatrix.Float32())

	// modelViewMatrix
	modelViewMatrix := viewMatrix.Mul(m.MatrixWorld())
	program.Uniform("modelViewMatrix").UniformMatrix4fv(false, modelViewMatrix.Float32())

	// normalMatrix
	normalMatrix := modelViewMatrix.Normal()
	program.Uniform("normalMatrix").UniformMatrix3fv(false, normalMatrix.Matrix3Float32())

	// draw triangles
	if material.Wireframe() {
		gl.LineWidth(float32(2))

		geometry.BindLineBuffer()
		gl.DrawElements(gl.LINES, geometry.LineCount(), gl.UNSIGNED_SHORT, nil) // gl.UNSIGNED_INT, UNSIGNED_SHORT
	} else {
		geometry.BindFaceBuffer()
		gl.DrawElements(gl.TRIANGLES, geometry.FaceCount(), gl.UNSIGNED_SHORT, nil /* uintptr(start) */) // gl.UNSIGNED_INT, UNSIGNED_SHORT
	}
}

func (r *Renderer) renderObjectNew(m Renderable, camera Camera) {
	material := m.Material().(*ShaderMaterial)
	var refreshMaterial bool

	if r.currentMaterial != material {
		if r.currentMaterial != nil {
			r.currentMaterial.Unbind()
		}

		r.currentMaterial = material
		refreshMaterial = true
	}

	// use program
	if material.UseProgram() {
		refreshMaterial = true
	}
	/*
		program := material.Program()
		if r.currentProgram != program {
			program.Use()
			r.currentProgram = program

			// different materials may have different programs
			// same material with different values of same uniforms (diffuse, reflection, ...) have same program
			refreshMaterial = true
		}
	*/

	if refreshMaterial || r.currentCamera != camera {
		//program.Uniform("projectionMatrix").UniformMatrix4fv(false, camera.ProjectionMatrix().Float32())
		material.UpdateUniform("projectionMatrix", camera.ProjectionMatrix().Float32())

		if r.currentCamera != camera {
			r.currentCamera = camera
		}
	}

	if refreshMaterial {
		material.UpdateUniforms()
	}

	geometry := m.Geometry()
	if r.currentGeometry != geometry {
		r.currentGeometry = geometry

		//program.DisableAttributes()
		material.DisableAttributes()
		geometry.BindVertexArray()

		// vertices
		geometry.BindPositionBuffer()
		//program.EnableAttribute("vertexPosition")
		//program.Attribute("vertexPosition").AttribPointer(3, gl.FLOAT, false, 0, nil)
		material.EnableAttribute("vertexPosition")
		//geometry.positionBuffer.Unbind(gl.ARRAY_BUFFER)

		// normal
		geometry.BindNormalBuffer()
		//program.EnableAttribute("vertexNormal")
		//program.Attribute("vertexNormal").AttribPointer(3, gl.FLOAT, false, 0, nil)
		material.EnableAttribute("vertexNormal")

		// uv
		geometry.BindUvBuffer()
		//program.EnableAttribute("vertexUV")
		//program.Attribute("vertexUV").AttribPointer(2, gl.FLOAT, false, 0, nil)
		material.EnableAttribute("vertexUV")

		// color
		geometry.BindColorBuffer()
		//program.EnableAttribute("vertexColor")
		//program.Attribute("vertexColor").AttribPointer(3, gl.FLOAT, false, 0, nil)
		material.EnableAttribute("vertexColor")
	}

	// for each object of same material and geometry

	// Model matrix : an identity matrix (model will be at the origin)
	//program.Uniform("modelMatrix").UniformMatrix4fv(false, m.MatrixWorld().Float32())
	material.UpdateUniform("modelMatrix", m.MatrixWorld().Float32())

	// viewMatrix
	viewMatrix := camera.MatrixWorld().Inverse()
	//program.Uniform("viewMatrix").UniformMatrix4fv(false, viewMatrix.Float32())
	material.UpdateUniform("viewMatrix", viewMatrix.Float32())

	// modelViewMatrix
	modelViewMatrix := viewMatrix.Mul(m.MatrixWorld())
	//program.Uniform("modelViewMatrix").UniformMatrix4fv(false, modelViewMatrix.Float32())
	material.UpdateUniform("modelViewMatrix", modelViewMatrix.Float32())

	// normalMatrix
	normalMatrix := modelViewMatrix.Normal()
	//program.Uniform("normalMatrix").UniformMatrix3fv(false, normalMatrix.Matrix3Float32())
	material.UpdateUniform("normalMatrix", normalMatrix.Float32())

	// draw triangles
	if material.Wireframe() {
		gl.LineWidth(float32(2))

		geometry.BindLineBuffer()
		gl.DrawElements(gl.LINES, geometry.LineCount(), gl.UNSIGNED_SHORT, nil) // gl.UNSIGNED_INT, UNSIGNED_SHORT
	} else {
		geometry.BindFaceBuffer()
		gl.DrawElements(gl.TRIANGLES, geometry.FaceCount(), gl.UNSIGNED_SHORT, nil /* uintptr(start) */) // gl.UNSIGNED_INT, UNSIGNED_SHORT
	}
}
