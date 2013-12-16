package engine

import (
	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
)

type RenderTarget struct {
	Texture

	frameBuffer   gl.Framebuffer
	textureBuffer gl.Texture
	renderBuffer  gl.Renderbuffer
	initialized   bool

	width, height int
	needsUpdate   bool
}

func NewRenderTarget(w, h int) *RenderTarget {
	return &RenderTarget{
		width:       w,
		height:      h,
		needsUpdate: true,
	}
}

// init frame buffers
func (t *RenderTarget) init() {
	t.frameBuffer = gl.GenFramebuffer()
	t.textureBuffer = gl.GenTexture()
	t.renderBuffer = gl.GenRenderbuffer()

	t.initialized = true
}

// update image and gl parameters
func (t *RenderTarget) update() {
	if !t.initialized {
		t.init()
	}

	// setup texture buffer
	t.textureBuffer.Bind(gl.TEXTURE_2D)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB,
		t.width, t.height,
		0, gl.RGB, gl.UNSIGNED_BYTE, nil) // empty image

	// setup frame buffer
	t.frameBuffer.Bind() // t.frameBuffer.BindTarget(gl.FRAMEBUFFER)

	slot := 0
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0+gl.GLenum(slot), gl.TEXTURE_2D, t.textureBuffer, 0)

	// setup depth buffer
	t.renderBuffer.Bind()
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT16, t.width, t.height)
	t.renderBuffer.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER)

	// configure framebuffer
	//gl.DrawBuffers(1, []gl.GLenum{gl.COLOR_ATTACHMENT0 + gl.GLenum(slot)})
	//gl.Viewport(0, 0, t.width, t.height)

	// check for errors during setup
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("could not initialize framebuffer")
	}

	// release
	t.textureBuffer.Unbind(gl.TEXTURE_2D)
	t.renderBuffer.Unbind()
	t.frameBuffer.Unbind()

	t.needsUpdate = false
}

// cleanup
func (t *RenderTarget) Dispose() {
	if t.frameBuffer != 0 {
		t.renderBuffer.Delete()
		t.textureBuffer.Delete()
		t.frameBuffer.Delete()
	}
}

// bind rendertarget
func (t *RenderTarget) BindFramebuffer() {
	if t.needsUpdate {
		t.update()
	}

	t.frameBuffer.Bind()
	//gl.Viewport(0, 0, t.width, t.height)
}

// bind texture of framebuffer in Texture Unit slot
func (t *RenderTarget) Bind(slot int) {
	if t.needsUpdate {
		t.update()
	}

	t.textureBuffer.Bind(gl.TEXTURE_2D)
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(slot))
}

func (t *RenderTarget) Unbind() {
	if t.initialized {
		t.textureBuffer.Unbind(gl.TEXTURE_2D)
		t.renderBuffer.Unbind()
		t.frameBuffer.Unbind()
	}
}

type RenderPass struct {
	scene  *Scene
	camera Camera
	target *RenderTarget

	clear      bool
	clearColor math.Color
	clearAlpha float64
}

func (rt *RenderPass) SetClear(b bool)            { rt.clear = b }
func (rt *RenderPass) SetClearColor(c math.Color) { rt.clearColor = c }
func (rt *RenderPass) SetClearAlpha(f float64)    { rt.clearAlpha = f }

func NewRenderPass(scene *Scene, camera Camera, target *RenderTarget) *RenderPass {
	return &RenderPass{
		scene:  scene,
		camera: camera,
		target: target,

		clear:      true,
		clearColor: math.Color{0, 0, 0},
		clearAlpha: 1.0,
	}
}

func NewShaderPass(material *Material, target *RenderTarget) *RenderPass {
	// default ortho camera
	camera := NewOrthographicCamera(-1, 1, 1, -1, 0, 1)
	//camera.SetPosition(math.Vector{0, 0, -1})
	//camera.LookAt(math.Vector{0, 0, 0})

	// default plane object
	plane := NewMesh(NewPlaneGeometry(2, 2), material)
	//plane.SetRotation(math.FromAxisAngle(math.Vector{1, 0, 0}, math.Pi))
	plane.SetPosition(math.Vector{0, 0, 0})

	// scene
	scene := NewScene()
	scene.AddChild(plane)

	return &RenderPass{
		scene:  scene,
		camera: camera,
		target: target,

		clear:      false,
		clearColor: math.Color{0, 0, 0},
		clearAlpha: 1.0,
	}
}
