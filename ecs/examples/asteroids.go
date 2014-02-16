package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"

	"github.com/der-antikeks/gisp/ecs"
)

var (
	camZ    = -1000.0
	deg2rad = math.Pi / 180.0

	MaxShipSpeed      = 100.0
	MaxAccelerate     = MaxShipSpeed / 10
	ShipRotationSpeed = 4.0

	TimeBetweenBullets = 250 * time.Millisecond
	BulletSpeed        = 2 * MaxShipSpeed
	BulletLifetime     = 5 * time.Second

	MaxAsteroidRotation = 2 * ShipRotationSpeed
	MaxAsteroidSpeed    = MaxShipSpeed
)

func main() {
	rand.Seed(time.Now().Unix())
	runtime.LockOSThread()

	engine := ecs.NewEngine()

	// managers
	im := NewInputManager()
	em := NewEntityManager(engine)

	// add systems
	engine.AddSystem(MotionControlSystem(im), 0)
	engine.AddSystem(MovementSystem(), 1)
	rs, window := RenderSystem(im)
	engine.AddSystem(rs, 10)

	// add entities
	engine.AddEntity(em.CreateSpaceship())
	for i := 0; i < 3; i++ {
		engine.AddEntity(em.CreateAsteroid(rand.Intn(4) + 4))
	}

	engine.AddEntity(em.CreateGame())

	// main loop
	lastTime := time.Now()
	renderTicker := time.Tick(time.Duration(1000/72) * time.Millisecond)

	for !window.ShouldClose() {
		select {
		case <-renderTicker:

			// calc delay
			currentTime := time.Now()
			delta := currentTime.Sub(lastTime)
			lastTime = currentTime

			// update
			engine.Update(delta)

			// Swap buffers
			window.SwapBuffers()
			glfw.PollEvents()
		}
	}
}

// components

type PositionComponent struct {
	X, Y     float64
	Rotation float64
}

// render system

type ColorComponent struct {
	R, G, B float64
}

type Point struct {
	X, Y float64
}

type MeshComponent struct {
	Points []Point
	Max    float64 // collision system
}

func RenderSystem(im *InputManager) (*ecs.System, *glfw.Window) {
	glfw.SetErrorCallback(func(err glfw.ErrorCode, desc string) {
		log.Fatalln("error callback:", err, desc)
	})

	if !glfw.Init() {
		log.Fatalf("failed to initialize glfw")
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, 1)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		log.Fatalf("create window: %v", err)
	}

	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	gl.Init()

	// callbacks
	window.SetKeyCallback(im.onKey)
	window.SetFramebufferSizeCallback(onResize)

	s := ecs.NewSystem("Render", func(position *PositionComponent, mesh *MeshComponent, color *ColorComponent) {
		gl.LoadIdentity()
		gl.Translated(position.X, position.Y, camZ)
		gl.Rotated(position.Rotation, 0, 0, 1)
		gl.Color3d(color.R, color.G, color.B)

		gl.Begin(gl.LINE_LOOP)
		for _, p := range mesh.Points {
			gl.Vertex3d(p.X, p.Y, 0)
		}
		gl.End()
	})

	s.SetPreUpdateFunc(func() {
		log.Println("initialize render system update loop")

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.LoadIdentity()
	})

	s.SetInitFunc(func() error {
		log.Println("initializing RenderSystem")

		// gl init
		gl.ShadeModel(gl.SMOOTH)
		gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)

		gl.ClearColor(0.1, 0.1, 0.1, 0.0)
		gl.ClearDepth(1)
		gl.DepthFunc(gl.LEQUAL)
		gl.Enable(gl.DEPTH_TEST)

		gl.LineWidth(1)
		gl.Enable(gl.LINE_SMOOTH)

		// set size
		w, h := window.GetFramebufferSize()
		onResize(window, w, h)

		return nil
	})

	s.SetCleanupFunc(func() error {
		log.Println("cleaning RenderSystem")
		return nil
	})

	return s, window
}

func onResize(w *glfw.Window, width int, height int) {
	h := float64(height) / float64(width)

	znear := 1.0
	zfar := 1000.0
	xmax := znear * 0.5

	gl.Viewport(0, 0, width, height)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Frustum(-xmax, xmax, -xmax*h, xmax*h, znear, zfar)
	//gl.Ortho(0, float64(width), 0, float64(height), 0, 128)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	//gl.Translated(0.0, 0.0, -20.0)
}

// movemant system

type VelocityComponent struct {
	X, Y    float64
	Angular float64
}

func MovementSystem() *ecs.System {
	s := ecs.NewSystem("Move", func(position *PositionComponent, velocity *VelocityComponent, t time.Duration) {
		position.X += velocity.X * t.Seconds()
		position.Y += velocity.Y * t.Seconds()
		position.Rotation += velocity.Angular * t.Seconds()
	})

	s.SetInitFunc(func() error {
		log.Println("initializing MovementSystem")
		return nil
	})

	s.SetCleanupFunc(func() error {
		log.Println("cleaning MovementSystem")
		return nil
	})

	return s
}

// motion control

type MotionControlComponent struct {
	AccelerationSpeed,
	RotationSpeed float64

	LeftKey,
	RightKey,
	AccelerateKey glfw.Key
}

func MotionControlSystem(im *InputManager) *ecs.System {
	s := ecs.NewSystem("Move", func(
		position *PositionComponent,
		velocity *VelocityComponent,
		control *MotionControlComponent,
		t time.Duration,
	) {

		if im.IsDown(control.LeftKey) {
			position.Rotation -= control.RotationSpeed * t.Seconds()
		}

		if im.IsDown(control.RightKey) {
			position.Rotation += control.RotationSpeed * t.Seconds()
		}

		if im.IsDown(control.AccelerateKey) {
			velocity.X += math.Cos(position.Rotation) * control.AccelerationSpeed * t.Seconds()
			velocity.Y += math.Sin(position.Rotation) * control.AccelerationSpeed * t.Seconds()
		}

	})

	s.SetInitFunc(func() error {
		log.Println("initializing MotionControlSystem")
		return nil
	})

	s.SetCleanupFunc(func() error {
		log.Println("cleaning MotionControlSystem")
		return nil
	})

	return s
}

type InputManager struct {
	keyPressed map[glfw.Key]bool
}

func NewInputManager() *InputManager {
	return &InputManager{
		keyPressed: map[glfw.Key]bool{},
	}
}

func (m *InputManager) onKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		m.keyPressed[key] = true

		if key == glfw.KeyEscape { // TODO: remove
			w.SetShouldClose(true)
		}

	case glfw.Release:
		delete(m.keyPressed, key)
	}
}

func (m *InputManager) IsDown(key glfw.Key) bool {
	return m.keyPressed[key]
}

// entity manager

type EntityManager struct {
	engine       *ecs.Engine
	asteroidsNum int
}

func NewEntityManager(engine *ecs.Engine) *EntityManager {
	return &EntityManager{engine: engine}
}

type GameState struct{}
type Window struct{ Width, Height int }
type Hud struct{}

func (m *EntityManager) CreateGame() *ecs.Entity {
	width, height := 0, 0

	return ecs.NewEntity(
		"game",
		&GameState{},
		&Window{width, height},
		&Hud{},
	)
}

// components: position, velocity, polygon(collision), keymovable
func (m *EntityManager) CreateSpaceship() *ecs.Entity {
	s := ecs.NewEntity(
		"spaceship",
		&PositionComponent{0, 0, 45},
		&VelocityComponent{-1 * MaxShipSpeed, 1 * MaxShipSpeed, 0},

		//&MotionControlComponent{},

		&ColorComponent{1, 1, 1},
		&MeshComponent{
			Points: []Point{
				Point{-10, -15},
				Point{0, -10},
				Point{10, -15},
				Point{0, 15},
			},
			Max: 15,
		},
	)

	return s
}

// components: position, velocity, polygon(collision), size
func (m *EntityManager) CreateAsteroid(size int) *ecs.Entity {
	rot := rand.Float64() * 360
	rad := (rot + 90) * deg2rad
	speed := rand.Float64() * MaxAsteroidSpeed

	a := ecs.NewEntity(
		fmt.Sprintf("asteroid%d", m.asteroidsNum),
		&PositionComponent{0, 0, rot},
		&VelocityComponent{
			speed * math.Cos(rad),
			speed * math.Sin(rad),
			0,
		},

		&ColorComponent{1, 1, 0},
	)

	m.asteroidsNum++

	mc := &MeshComponent{
		Points: make([]Point, 7),
		Max:    0,
	}

	step := (2.0 * math.Pi) / float64(len(mc.Points))
	max := float64(size * 10)
	min := max / 2

	for i := range mc.Points {
		length := (rand.Float64() * (max - min)) + min
		angle := float64(i) * step

		mc.Points[i].X = length * math.Cos(angle)
		mc.Points[i].Y = length * math.Sin(angle)
	}

	a.Add(mc)

	return a
}

// components: position, velocity, polygon(collision), lifetime
func (m *EntityManager) CreateBullet() *ecs.Entity {
	b := ecs.NewEntity(
		"bullet",
		&PositionComponent{0, 0, 0},
		&VelocityComponent{-1 * BulletSpeed, 1 * BulletSpeed, 0},

		&ColorComponent{1, 0, 0},
		&MeshComponent{
			Points: []Point{
				Point{-1, -5},
				Point{1, -5},
				Point{1, 5},
				Point{-1, 5},
			},
			Max: 5,
		},
	)

	return b
}
