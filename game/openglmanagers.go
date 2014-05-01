package game

import (
	"log"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
)

func InitOpenGL(w, h int, title string) (*InputManager, *WindowManager) {
	im := NewInputManager()
	wm := NewWindowManager(w, h, title, im)

	return im, wm
}

type WindowManager struct {
	width, height int
	window        *glfw.Window
}

func NewWindowManager(w, h int, title string, im *InputManager) *WindowManager {
	m := &WindowManager{
		width:  w,
		height: h,
	}

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
	gl.ShadeModel(gl.SMOOTH)
	gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)

	gl.ClearColor(0.1, 0.1, 0.1, 0.0)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)
	gl.Enable(gl.DEPTH_TEST)

	gl.LineWidth(1)
	gl.Enable(gl.LINE_SMOOTH)

	// set size
	w, h = m.window.GetFramebufferSize()
	m.onResize(m.window, w, h)

	return m
}

func (m *WindowManager) isRunning() bool {
	return !m.window.ShouldClose()
}

func (m *WindowManager) Update() {
	m.window.SwapBuffers()
	glfw.PollEvents()
}

func (m *WindowManager) Cleanup() {
	glfw.Terminate()
}

func (m *WindowManager) onResize(w *glfw.Window, width int, height int) {
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

type Key glfw.Key

const (
	KeyEscape = glfw.KeyEscape
	KeyEnter  = glfw.KeyEnter
	KeyPause  = glfw.KeyPause
)

type InputManager struct {
	keyPressed   map[glfw.Key]bool
	mousePressed map[glfw.MouseButton]bool
	mouseClicked map[glfw.MouseButton]bool
	mx, my, zoom float64
}

func NewInputManager() *InputManager {
	return &InputManager{
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
}

func (m *InputManager) IsKeyDown(key glfw.Key) bool {
	return m.keyPressed[key]
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
}

func (m *InputManager) MousePos() (x, y float64) {
	return m.mx, m.my
}

func (m *InputManager) onMouseScroll(w *glfw.Window, xoff float64, yoff float64) {
	m.zoom = yoff
}

func (m *InputManager) onMouseButton(w *glfw.Window, b glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		m.mousePressed[b] = true
		m.mouseClicked[b] = false
	case glfw.Release:
		delete(m.mousePressed, b)

		if v, ok := m.mouseClicked[b]; ok && !v {
			m.mouseClicked[b] = true
		}
	}
}

func (m *InputManager) IsMouseDown(button glfw.MouseButton) bool {
	return m.mousePressed[button]
}

// mouse up after a down without movement
func (m *InputManager) IsMouseClick(button glfw.MouseButton) bool {
	s := m.mouseClicked[button]
	m.mouseClicked[button] = false
	return s
}
