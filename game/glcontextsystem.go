package game

import (
	"log"
	"runtime"
	"sync"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
)

/*
	SwapBuffers/PollEvents
	SetSize/onResize
	set/getTitle
	onKey/Mouse/Button
	IsKeyDown/IsMouseClick
	SubscribeOnMouseScroll(chan x/y float64)
*/
type glContextSystem struct {
	mChan chan func()
	mDone chan struct{}

	width, height int
	window        *glfw.Window

	keyPressed   map[glfw.Key]bool
	mousePressed map[glfw.MouseButton]bool
	mouseClicked map[glfw.MouseButton]bool
	mx, my, zoom float64

	resize,
	key,
	mousebutton,
	mousemove,
	mousescroll *Observer
}

var (
	ctxInstance *glContextSystem
	ctxOnce     sync.Once
)

type CtxOpts struct {
	Title string
	W, H  int
}

func GlContextSystem(opts *CtxOpts) *glContextSystem {
	ctxOnce.Do(func() {
		if opts == nil {
			log.Fatal("zero options init of system")
		}

		ctxInstance = &glContextSystem{
			mChan: make(chan func()),
			mDone: make(chan struct{}),

			width:  opts.W,
			height: opts.H,

			keyPressed:   map[glfw.Key]bool{},
			mousePressed: map[glfw.MouseButton]bool{},
			mouseClicked: map[glfw.MouseButton]bool{},

			//resize:      NewObserver(ctxInstance.sendSize),
			key:         NewObserver(nil),
			mousebutton: NewObserver(nil),
			mousemove:   NewObserver(nil),
			mousescroll: NewObserver(nil),
		}
		ctxInstance.resize = NewObserver(ctxInstance.sendSize)

		// main thread
		go func() {
			runtime.LockOSThread()
			for mf := range ctxInstance.mChan {
				mf()
				ctxInstance.mDone <- struct{}{}
			}
		}()

		// initialize
		ctxInstance.initGl(opts.Title)
	})

	return ctxInstance
}

// run function on main thread
func (s *glContextSystem) MainThread(f func()) {
	s.mChan <- f
	<-s.mDone
}

func (s *glContextSystem) initGl(title string) {
	s.MainThread(func() {
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
		s.window, err = glfw.CreateWindow(s.width, s.height, title, nil, nil)
		if err != nil {
			log.Fatalf("create window: %v", err)
		}

		s.window.MakeContextCurrent()
		glfw.SwapInterval(1)
		gl.Init()

		// callbacks
		s.window.SetFramebufferSizeCallback(s.onResize)

		s.window.SetKeyCallback(s.onKey)
		s.window.SetCursorPositionCallback(s.onMouseMove)
		s.window.SetScrollCallback(s.onMouseScroll)
		s.window.SetMouseButtonCallback(s.onMouseButton)
		//s.window.SetInputMode(glfw.Cursor, glfw.CursorNormal /*glfw.CursorDisabled*/)

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
		w, h := s.window.GetFramebufferSize()
		s.onResize(s.window, w, h)
	})
}

func (s *glContextSystem) isRunning() bool {
	return !s.window.ShouldClose()
}

func (s *glContextSystem) Update() {
	s.MainThread(func() {
		s.window.SwapBuffers()
		glfw.PollEvents()
	})
}

func (s *glContextSystem) Cleanup() {
	glfw.Terminate()
}

func (s *glContextSystem) onResize(w *glfw.Window, width, height int) {
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

	s.width = width
	s.height = height

	s.resize.Publish(MessageResize{width, height})
}

func (s *glContextSystem) sendSize() <-chan interface{} {
	c := make(chan interface{}, 1)
	defer close(c)

	c <- MessageResize{s.width, s.height}
	return c
}

func (s *glContextSystem) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.window.SetSize(width, height) // TODO: Loops to onResize()?
}

func (s *glContextSystem) Size() (width, height int) {
	return s.width, s.height
}

func (s *glContextSystem) Close() {
	s.window.SetShouldClose(true)
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

func (s *glContextSystem) onKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		s.keyPressed[key] = true
	case glfw.Release:
		delete(s.keyPressed, key)
	}

	s.key.Publish(MessageKey(key))
}

func (s *glContextSystem) IsKeyDown(key Key) bool {
	return s.keyPressed[glfw.Key(key)]
}

func (s *glContextSystem) AnyKeyDown() bool {
	for _, d := range s.keyPressed {
		if d {
			return true
		}
	}
	return false
}

func (s *glContextSystem) onMouseMove(window *glfw.Window, xpos float64, ypos float64) {
	s.mx, s.my = xpos, ypos

	for b := range s.mouseClicked {
		delete(s.mouseClicked, b)
	}

	s.mousemove.Publish(MessageMouseMove{xpos, ypos})
}

func (s *glContextSystem) MousePos() (x, y float64) {
	return s.mx, s.my
}

func (s *glContextSystem) onMouseScroll(w *glfw.Window, xoff float64, yoff float64) {
	s.zoom += yoff
	s.mousescroll.Publish(MessageMouseScroll(yoff))
}

func (s *glContextSystem) MouseScroll() float64 {
	return s.zoom
}

func (s *glContextSystem) onMouseButton(w *glfw.Window, b glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		s.mousePressed[b] = true
		delete(s.mouseClicked, b)
	case glfw.Release:
		delete(s.mousePressed, b)

		if v, ok := s.mouseClicked[b]; ok && !v {
			s.mouseClicked[b] = true
		}
	}

	s.mousebutton.Publish(MessageMouseButton(b))
}

type MouseButton int

const (
	MouseLeft  = MouseButton(glfw.MouseButton1)
	MouseRight = MouseButton(glfw.MouseButton2)
)

func (s *glContextSystem) IsMouseDown(button MouseButton) bool {
	return s.mousePressed[glfw.MouseButton(button)]
}

// mouse up after a down without movement
func (s *glContextSystem) IsMouseClick(button MouseButton) bool {
	return s.mouseClicked[glfw.MouseButton(button)]
}

func (s *glContextSystem) OnResize() *Observer      { return s.resize }
func (s *glContextSystem) OnKey() *Observer         { return s.key }
func (s *glContextSystem) OnMouseButton() *Observer { return s.mousebutton }
func (s *glContextSystem) OnMouseMove() *Observer   { return s.mousemove }
func (s *glContextSystem) OnMouseScroll() *Observer { return s.mousescroll }
