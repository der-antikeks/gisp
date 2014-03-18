package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"

	"github.com/der-antikeks/gisp/ecs"
)

func main() {
	rand.Seed(time.Now().Unix())
	runtime.LockOSThread()

	engine := ecs.NewEngine()

	// managers
	im := NewInputManager()
	em := NewEntityManager(engine)
	w, h := 640, 480
	wm := NewWindowManager(w, h, "Testing", im)
	defer wm.cleanup()

	// systems
	engine.AddSystem(NewGameStateSystem(em, wm), 1)
	engine.AddSystem(NewMenuSystem(im), 2)
	engine.AddSystem(NewRenderSystem(), 10)

	// entities
	em.CreateGame()

	// main loop
	var (
		lastTime     = time.Now()
		currentTime  time.Time
		delta        time.Duration
		renderTicker = time.Tick(time.Duration(1000/70) * time.Millisecond)
	)

	for wm.isRunning() {
		select {
		case <-renderTicker:
			// calc delay
			currentTime = time.Now()
			delta = currentTime.Sub(lastTime)
			lastTime = currentTime

			// update
			if err := engine.Update(delta); err != nil {
				log.Fatal(err)
			}

			// Swap buffers
			wm.update()
		}
	}
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

func (m *WindowManager) update() {
	m.window.SwapBuffers()
	glfw.PollEvents()
}

func (m *WindowManager) cleanup() {
	glfw.Terminate()
}

func (m *WindowManager) onResize(w *glfw.Window, width int, height int) {
	//h := float64(height) / float64(width)
	//znear := 1.0
	//zfar := 1000.0
	//xmax := znear * 0.5

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Viewport(0, 0, width, height)
	//gl.Frustum(-xmax, xmax, -xmax*h, xmax*h, znear, zfar)
	gl.Ortho(-float64(width)/2, float64(width)/2, -float64(height)/2, float64(height)/2, 0, 128)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	//gl.Translated(0.0, 0.0, -20.0)

	m.width = width
	m.height = height
}

func (m *WindowManager) Close() {
	m.window.SetShouldClose(true)
}

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

		if key == glfw.KeyEscape { // TODO: move to game-status-system
			w.SetShouldClose(true)
		}

	case glfw.Release:
		delete(m.keyPressed, key)
	}
}

func (m *InputManager) IsKeyDown(key glfw.Key) bool {
	return m.keyPressed[key]
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

const (
	PositionType ecs.ComponentType = iota
	MeshType
	ColorType

	GameStateType
	MenuType
)

type Point struct {
	X, Y float64
}

func (p Point) Distance(o Point) float64 {
	dx, dy := o.X-p.X, o.Y-p.Y
	return math.Sqrt(dx*dx + dy*dy)
}

type Collidable interface {
	ContainsPoint(Point) bool
	Contains(Collidable) bool
	Intersects(Collidable) bool
}

type Circle struct {
	Center Point
	Radius float64
}

type AABB struct {
	Center, Half Point
}

func (a AABB) ContainsPoint(p Point) bool {
	return a.Center.X-a.Half.X <= p.X &&
		a.Center.X+a.Half.X >= p.X &&
		a.Center.Y-a.Half.Y <= p.Y &&
		a.Center.Y+a.Half.Y >= p.Y
}

func (a AABB) Contains(c Collidable) bool {
	b, ok := c.(AABB)
	if !ok {
		return false
	}

	return a.Center.X-a.Half.X <= b.Center.X-b.Half.X &&
		a.Center.X+a.Half.X >= b.Center.X+b.Half.X &&
		a.Center.Y-a.Half.Y <= b.Center.Y-b.Half.Y &&
		a.Center.Y+a.Half.Y >= b.Center.Y+b.Half.Y
}

func (a AABB) Intersects(c Collidable) bool {
	b, ok := c.(AABB)
	if !ok {
		return false
	}

	return a.Center.X-a.Half.X <= b.Center.X+b.Half.X &&
		a.Center.X+a.Half.X >= b.Center.X-b.Half.X &&
		a.Center.Y-a.Half.Y <= b.Center.Y+b.Half.Y &&
		a.Center.Y+a.Half.Y >= b.Center.Y-b.Half.Y
}

type PositionComponent struct {
	Position Point
	Rotation float64
}

func (c PositionComponent) Type() ecs.ComponentType {
	return PositionType
}

type MeshComponent struct {
	Points   []Point
	Collider Collidable
}

func (c MeshComponent) Type() ecs.ComponentType {
	return MeshType
}

type ColorComponent struct {
	R, G, B float64
}

func (c ColorComponent) Type() ecs.ComponentType {
	return ColorType
}

type GameStateComponent struct {
	State string
}

func (c GameStateComponent) Type() ecs.ComponentType {
	return GameStateType
}

type MenuComponent struct {
	Name string
}

func (c MenuComponent) Type() ecs.ComponentType {
	return MenuType
}

type EntityManager struct {
	engine *ecs.Engine
}

func NewEntityManager(e *ecs.Engine) *EntityManager {
	return &EntityManager{
		engine: e,
	}
}

func (em *EntityManager) CreateGame() {
	s := ecs.NewEntity(
		"game",
		&GameStateComponent{"init"},
	)

	if err := em.engine.AddEntity(s); err != nil {
		log.Fatal(err)
	}
}

func (em *EntityManager) CreateButton(name string, x, y, w, h float64) {
	wh, hh := w*0.5, h*0.5

	s := ecs.NewEntity(
		"Button "+name,
		&MenuComponent{name},
		&PositionComponent{Position: Point{x, y}},

		&MeshComponent{
			Points: []Point{
				Point{-wh, -hh},
				Point{wh, -hh},
				Point{wh, hh},
				Point{-wh, hh},
			},
			Collider: AABB{
				Center: Point{0, 0},
				Half:   Point{wh, hh},
			},
		},
	)

	s.State("normal").Add(&ColorComponent{0, 1, 0})
	s.State("hover").Add(&ColorComponent{0, 0, 1})
	s.State("pressed").Add(&ColorComponent{1, 0, 0})

	s.ChangeState("normal")

	if err := em.engine.AddEntity(s); err != nil {
		log.Fatal(err)
	}
}

func (em *EntityManager) RemoveAllButtons() {
	for _, en := range em.engine.Collection(MenuType, PositionType, MeshType).Entities() {
		em.engine.RemoveEntity(en)
	}
}

func NewGameStateSystem(em *EntityManager, wm *WindowManager) ecs.System {
	return ecs.CollectionSystem(
		func(delta time.Duration, en *ecs.Entity) {
			s := en.Get(GameStateType).(*GameStateComponent)

			switch s.State {
			case "init":
				log.Println("init")
				s.State = "load-menu"

			case "load-menu":
				log.Println("load menu")
				em.RemoveAllButtons()

				em.CreateButton("LoadSubmenu", 0, 100, 100, 20)
				em.CreateButton("B", 0, 70, 100, 20)
				em.CreateButton("C", 0, 40, 100, 20)
				em.CreateButton("D", 0, 10, 100, 20)
				em.CreateButton("E", 0, -20, 100, 20)

				em.CreateButton("Exit", 0, -80, 100, 20)

				s.State = "menu"
			case "menu":

			case "load-submenu":
				log.Println("load submenu")
				em.RemoveAllButtons()

				em.CreateButton("ReturnMenu", 0, 100, 100, 20)
				em.CreateButton("F", 0, 70, 100, 20)

				em.CreateButton("Exit", 0, -80, 100, 20)

				s.State = "submenu"
			case "submenu":

			case "exit":
				log.Println("exit")
				wm.Close()
			default:
			}

		},
		[]ecs.ComponentType{GameStateType},
	)
}

type MenuSystem struct {
	im *InputManager

	buttons   *ecs.Collection
	gamestate *ecs.Collection
}

func NewMenuSystem(im *InputManager) ecs.System {
	return &MenuSystem{
		im: im,
	}
}

func (s *MenuSystem) AddedToEngine(e *ecs.Engine) error {
	s.buttons = e.Collection(MenuType, PositionType, MeshType)
	s.gamestate = e.Collection(GameStateType)
	return nil
}

func (s *MenuSystem) RemovedFromEngine(*ecs.Engine) error {
	s.buttons = nil
	s.gamestate = nil
	return nil
}

func (s *MenuSystem) Update(delta time.Duration) error {
	state := s.gamestate.First().Get(GameStateType).(*GameStateComponent)

	for _, en := range s.buttons.Entities() {
		x, y := s.im.MousePos()
		// TODO: cache old position & compare

		m := en.Get(MenuType).(*MenuComponent)
		p := en.Get(PositionType).(*PositionComponent).Position
		c := en.Get(MeshType).(*MeshComponent).Collider
		mp := Point{(x - 320) - p.X, (240 - y) - p.Y}

		if c.ContainsPoint(mp) {
			if s.im.IsMouseClick(glfw.MouseButton1) {
				log.Println("clicked", m.Name)

				switch m.Name {
				case "Exit":
					state.State = "exit"
				case "LoadSubmenu":
					state.State = "load-submenu"
				case "ReturnMenu":
					state.State = "load-menu"
				}
			}

			if s.im.IsMouseDown(glfw.MouseButton1) {
				en.ChangeState("pressed")
			} else {
				en.ChangeState("hover")
			}
		} else {
			en.ChangeState("normal")
		}
	}

	return nil
}

type RenderSystem struct {
	engine   *ecs.Engine
	drawable *ecs.Collection
}

func NewRenderSystem() ecs.System {
	return &RenderSystem{}
}

func (s *RenderSystem) AddedToEngine(e *ecs.Engine) error {
	s.engine = e
	s.drawable = e.Collection(PositionType, MeshType, ColorType)
	return nil
}

func (s *RenderSystem) RemovedFromEngine(*ecs.Engine) error {
	s.drawable = nil
	s.engine = nil
	return nil
}

func (s *RenderSystem) Update(delta time.Duration) error {
	// init
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.LoadIdentity()

	for _, e := range s.drawable.Entities() {
		p := e.Get(PositionType).(*PositionComponent)
		m := e.Get(MeshType).(*MeshComponent)
		c := e.Get(ColorType).(*ColorComponent)

		gl.LoadIdentity()
		gl.Translated(p.Position.X, p.Position.Y, 0)
		gl.Rotated(p.Rotation, 0, 0, 1)
		gl.Color3d(c.R, c.G, c.B)

		gl.Begin(gl.POLYGON)
		for _, point := range m.Points {
			gl.Vertex3d(point.X, point.Y, 0)
		}
		gl.End()
	}

	return nil
}
