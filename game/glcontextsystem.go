package game

import (
	"log"
	"runtime"
	"sync"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
)

var (
	mOnce sync.Once
	mChan = make(chan func())
	mDone = make(chan struct{})
)

func MainThread(f func()) {
	mOnce.Do(func() {
		go func() {
			runtime.LockOSThread()
			for mf := range mChan {
				mf()
				mDone <- struct{}{}
			}
		}()
	})

	mChan <- f
	<-mDone
}

// NewGlContextSystem()

func InitOpenGL(w, h int, title string, e *Engine) (*InputManager, *WindowManager) {
	im := NewInputManager(e)
	wm := NewWindowManager(w, h, title, im, e)

	return im, wm
}

type WindowManager struct {
	engine        *Engine
	width, height int
	window        *glfw.Window
}

func NewWindowManager(w, h int, title string, im *InputManager, e *Engine) *WindowManager {
	m := &WindowManager{
		engine: e,
		width:  w,
		height: h,
	}

	MainThread(func() {
		// init glfw
		glfw.SetErrorCallback(func(err glfw.ErrorCode, desc string) {
			log.Fatalln("error callback:", err, desc)
		})

		if !glfw.Init() {
			log.Fatalf("failed to initialize glfw")
		}

		glfw.WindowHint(glfw.Resizable, 1)
		glfw.WindowHint(glfw.Samples, 4)

		var err error
		m.window, err = glfw.CreateWindow(w, h, title, nil, nil)
		if err != nil {
			log.Fatalf("create window: %v", err)
		}

		m.window.MakeContextCurrent()
		glfw.SwapInterval(1)
		gl.Init()

		// callbacks
		m.window.SetFramebufferSizeCallback(m.onResize)

		m.window.SetKeyCallback(im.onKey)
		m.window.SetCursorPositionCallback(im.onMouseMove)
		m.window.SetScrollCallback(im.onMouseScroll)
		m.window.SetMouseButtonCallback(im.onMouseButton)
		//m.window.SetInputMode(glfw.Cursor, glfw.CursorNormal /*glfw.CursorDisabled*/)

		// init gl

		// clearing
		gl.ClearColor(0.1, 0.1, 0.1, 0.0)
		gl.ClearDepth(1)
		gl.ClearStencil(0)

		// depth
		gl.DepthFunc(gl.LEQUAL)
		gl.Enable(gl.DEPTH_TEST)

		// cull face
		gl.FrontFace(gl.CCW)
		gl.CullFace(gl.BACK)
		gl.Enable(gl.CULL_FACE)

		// lines
		gl.LineWidth(1)
		gl.Enable(gl.LINE_SMOOTH)

		// aa
		gl.ShadeModel(gl.SMOOTH)
		gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)

		// set size
		w, h = m.window.GetFramebufferSize()
		m.onResize(m.window, w, h)
	})

	return m
}

func (m *WindowManager) isRunning() bool {
	return !m.window.ShouldClose()
}

func (m *WindowManager) Update() {
	MainThread(func() {
		m.window.SwapBuffers()
		glfw.PollEvents()
	})
}

func (m *WindowManager) Cleanup() {
	glfw.Terminate()
}

func (m *WindowManager) onResize(w *glfw.Window, width, height int) {
	//h := float64(height) / float64(width)
	//znear := 1.0
	//zfar := 1000.0
	//xmax := znear * 0.5

	if height < 1 {
		height = 1
	}

	if width < 1 {
		width = 1
	}

	// TODO: set aspect (w / h) of perspective camera

	//	gl.MatrixMode(gl.PROJECTION)
	//	gl.LoadIdentity()
	gl.Viewport(0, 0, width, height)
	//gl.Frustum(-xmax, xmax, -xmax*h, xmax*h, znear, zfar)
	//	gl.Ortho(-float64(width)/2, float64(width)/2, -float64(height)/2, float64(height)/2, 0, 128)
	//	gl.MatrixMode(gl.MODELVIEW)
	//	gl.LoadIdentity()
	//gl.Translated(0.0, 0.0, -20.0)

	m.width = width
	m.height = height

	m.engine.Publish(MessageResize{width, height})
}

func (m *WindowManager) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.window.SetSize(width, height) // TODO: Loops to onResize()?
}

func (m *WindowManager) Size() (width, height int) {
	return m.width, m.height
}

func (m *WindowManager) Close() {
	m.window.SetShouldClose(true)
}

type Key int

const (
	KeyEscape      = Key(glfw.KeyEscape)
	KeyEnter       = Key(glfw.KeyEnter)
	KeyPause       = Key(glfw.KeyPause)
	KeySpace       = Key(glfw.KeySpace)
	KeyLeftControl = Key(glfw.KeyLeftControl)

	KeyUp    = Key(glfw.KeyUp)
	KeyDown  = Key(glfw.KeyDown)
	KeyLeft  = Key(glfw.KeyLeft)
	KeyRight = Key(glfw.KeyRight)

	KeyQ = Key(glfw.KeyQ)
	KeyE = Key(glfw.KeyE)
	KeyW = Key(glfw.KeyW)
	KeyS = Key(glfw.KeyS)
	KeyA = Key(glfw.KeyA)
	KeyD = Key(glfw.KeyD)
)

type InputManager struct {
	engine       *Engine
	keyPressed   map[glfw.Key]bool
	mousePressed map[glfw.MouseButton]bool
	mouseClicked map[glfw.MouseButton]bool
	mx, my, zoom float64
}

func NewInputManager(e *Engine) *InputManager {
	return &InputManager{
		engine:       e,
		keyPressed:   map[glfw.Key]bool{},
		mousePressed: map[glfw.MouseButton]bool{},
		mouseClicked: map[glfw.MouseButton]bool{},
	}
}

func (m *InputManager) onKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		m.keyPressed[key] = true
	case glfw.Release:
		delete(m.keyPressed, key)
	}

	m.engine.Publish(MessageKey(key))
}

func (m *InputManager) IsKeyDown(key Key) bool {
	return m.keyPressed[glfw.Key(key)]
}

func (m *InputManager) AnyKeyDown() bool {
	for _, d := range m.keyPressed {
		if d {
			return true
		}
	}
	return false
}

func (m *InputManager) onMouseMove(window *glfw.Window, xpos float64, ypos float64) {
	m.mx, m.my = xpos, ypos

	for b := range m.mouseClicked {
		delete(m.mouseClicked, b)
	}

	m.engine.Publish(MessageMouseMove{xpos, ypos})
}

func (m *InputManager) MousePos() (x, y float64) {
	return m.mx, m.my
}

func (m *InputManager) onMouseScroll(w *glfw.Window, xoff float64, yoff float64) {
	m.zoom += yoff
	m.engine.Publish(MessageMouseScroll(yoff))
}

func (m *InputManager) MouseScroll() float64 {
	return m.zoom
}

func (m *InputManager) onMouseButton(w *glfw.Window, b glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		m.mousePressed[b] = true
		delete(m.mouseClicked, b)
	case glfw.Release:
		delete(m.mousePressed, b)

		if v, ok := m.mouseClicked[b]; ok && !v {
			m.mouseClicked[b] = true
		}
	}

	m.engine.Publish(MessageMouseButton(b))
}

type MouseButton int

const (
	MouseLeft  = MouseButton(glfw.MouseButton1)
	MouseRight = MouseButton(glfw.MouseButton2)
)

func (m *InputManager) IsMouseDown(button MouseButton) bool {
	return m.mousePressed[glfw.MouseButton(button)]
}

// mouse up after a down without movement
func (m *InputManager) IsMouseClick(button MouseButton) bool {
	s := m.mouseClicked[glfw.MouseButton(button)]
	return s
}
