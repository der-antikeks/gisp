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

const (
	_ ecs.SystemPriority = iota
	PriorityBeforeRender
	PriorityRender
	PriorityAfterRender
)

var (
	deg2rad = math.Pi / 180.0

	MaxShipSpeed      = 100.0 // pixels per second
	MaxAccelerate     = MaxShipSpeed
	ShipRotationSpeed = 180.0 // deg per second

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
	w, h := 640, 480
	wm := NewWindowManager(w, h, "Testing", im)
	defer wm.cleanup()

	// systems
	maxx, maxy := float64(w)*0.5, float64(h)*0.5
	engine.AddSystem(NewMovementSystem(-maxx, maxx, -maxy, maxy), PriorityBeforeRender)
	engine.AddSystem(NewRenderSystem(), PriorityRender)
	engine.AddSystem(NewCollisionSystem(), PriorityAfterRender)
	engine.AddSystem(NewAsteroidSpawnSystem(em), PriorityBeforeRender)
	engine.AddSystem(NewMotionControlSystem(im), PriorityBeforeRender)
	engine.AddSystem(NewBulletSystem(im, em), PriorityBeforeRender)

	// entities
	em.createSpaceship(0, 0)

	max := float64(h) * 0.5
	min := max
	for i := 0; i < 3; i++ {
		dir := rand.Float64() * 2 * math.Pi
		dist := rand.Float64()*(max-min) + min
		x := dist * math.Cos(dir)
		y := dist * math.Sin(dir)

		em.createAsteroid(x, y, 5)
	}

	//em.CreateGame()

	// main loop
	var (
		lastTime     = time.Now()
		currentTime  time.Time
		delta        time.Duration
		renderTicker = time.Tick(time.Duration(1000/70) * time.Millisecond)

		ratio     = 0.01
		fps       = 70.0
		nextPrint = lastTime
	)

	for wm.isRunning() {
		select {
		case <-renderTicker:
			// calc delay
			currentTime = time.Now()
			delta = currentTime.Sub(lastTime)
			lastTime = currentTime

			// fps test
			fps = fps*(1-ratio) + (1.0/delta.Seconds())*ratio
			if fps >= math.Inf(1) {
				fps = 72.0
			}
			if currentTime.After(nextPrint) {
				nextPrint = currentTime.Add(time.Second / 2.0)
				fmt.Println(fps)
			}

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
	m.window.SetKeyCallback(im.onKey)
	m.window.SetFramebufferSizeCallback(m.onResize)

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

		if key == glfw.KeyEscape { // TODO: move to game-status-system
			w.SetShouldClose(true)
		}

	case glfw.Release:
		delete(m.keyPressed, key)
	}
}

func (m *InputManager) IsDown(key glfw.Key) bool {
	return m.keyPressed[key]
}

type EntityManager struct {
	engine       *ecs.Engine
	asteroidsNum int
	bulletsNum   int
}

func NewEntityManager(e *ecs.Engine) *EntityManager {
	return &EntityManager{
		engine: e,
	}
}

/*
type GameState struct{}
type Hud struct{}

func (em *EntityManager) CreateGame() *ecs.Entity {
	s := ecs.NewEntity(
		"game",
		&GameState{},
		&Hud{},
	)

	if err := em.engine.AddEntity(s); err != nil {
		log.Fatal(err)
	}
}
*/

func (em *EntityManager) createSpaceship(x, y float64) {
	/*
		velocity = rate of change of an object (m/s)
		acceleration = rate of change of a velocity (m/s^2)

		new position = old position + velocity * time
		new velocity = old velocity + acceleration * time

		acceleration = delta velocity / delta time
		velocity = delta position / delta time
	*/

	s := em.engine.CreateEntity("spaceship")
	s.Set(
		ShipStatusComponent{
			Lifes: 5,
		},

		PositionComponent{Point{x, y}, 0},
		VelocityComponent{},

		MotionControlComponent{
			AccelerationSpeed: MaxAccelerate,
			MaxVelocity:       MaxShipSpeed,
			RotationSpeed:     ShipRotationSpeed,

			LeftKey:         glfw.KeyA,
			RightKey:        glfw.KeyD,
			AccelerateKey:   glfw.KeyW,
			DecelerationKey: glfw.KeyS,
		},

		ColorComponent{1, 1, 1},
		MeshComponent{
			Points: []Point{
				Point{-10, -15},
				Point{0, -10},
				Point{10, -15},
				Point{0, 15},
			},
			Max: 15,
		},

		CannonComponent{
			LastBullet:  time.Now(),
			BulletSpeed: BulletSpeed,
			FireKey:     glfw.KeySpace,
		},
	)
}

func (em *EntityManager) createAsteroid(x, y float64, size int) {
	rot := rand.Float64() * 360
	rad := (rot + 90) * deg2rad
	speed := rand.Float64() * MaxAsteroidSpeed

	a := em.engine.CreateEntity(fmt.Sprintf("asteroid%d", em.asteroidsNum))
	a.Set(
		AsteroidStatusComponent{
			Size: size,
		},

		PositionComponent{Point{x, y}, rot},
		VelocityComponent{
			Point{
				speed * math.Cos(rad),
				speed * math.Sin(rad),
			}, MaxAsteroidRotation * rand.Float64(),
		},

		ColorComponent{1, 1, 0},
	)

	em.asteroidsNum++

	mc := MeshComponent{
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

		mc.Max = math.Max(mc.Max, length)
	}

	a.Set(mc)
}

func (em *EntityManager) createBullet(x, y, vx, vy float64) {
	b := em.engine.CreateEntity(fmt.Sprintf("bullet%d", em.bulletsNum))
	b.Set(
		BulletStatusComponent{
			LifeTime: time.Now().Add(2 * time.Second),
		},

		PositionComponent{Point{x, y}, 0},
		VelocityComponent{Point{vx, vy}, 0},

		ColorComponent{1, 0, 0},
		MeshComponent{
			Points: []Point{
				Point{2, 2},
				Point{2, -2},
				Point{-2, -2},
				Point{-2, 2},
			},
			Max: 2,
		},
	)

	em.bulletsNum++
}

// COMPONENTS

const (
	PositionType ecs.ComponentType = iota
	VelocityType
	MotionControlType

	MeshType
	ColorType

	ShipStatusType
	AsteroidStatusType
	BulletStatusType

	CannonType
)

type MotionControlComponent struct {
	AccelerationSpeed,
	MaxVelocity,
	RotationSpeed float64

	LeftKey,
	RightKey,
	AccelerateKey,
	DecelerationKey glfw.Key
}

func (c MotionControlComponent) Type() ecs.ComponentType {
	return MotionControlType
}

type ColorComponent struct {
	R, G, B float64
}

func (c ColorComponent) Type() ecs.ComponentType {
	return ColorType
}

type Point struct {
	X, Y float64
}

func (p Point) Distance(o Point) float64 {
	dx, dy := o.X-p.X, o.Y-p.Y
	return math.Sqrt(dx*dx + dy*dy)
}

type PositionComponent struct {
	Position Point
	Rotation float64
}

func (c PositionComponent) Type() ecs.ComponentType {
	return PositionType
}

type VelocityComponent struct {
	Velocity Point
	Angular  float64
}

func (c VelocityComponent) Type() ecs.ComponentType {
	return VelocityType
}

type MeshComponent struct {
	Points []Point
	Max    float64
}

func (c MeshComponent) Type() ecs.ComponentType {
	return MeshType
}

type ShipStatusComponent struct {
	Lifes int
	Score int
}

func (c ShipStatusComponent) Type() ecs.ComponentType {
	return ShipStatusType
}

type CannonComponent struct {
	LastBullet  time.Time
	BulletSpeed float64

	FireKey glfw.Key
}

func (c CannonComponent) Type() ecs.ComponentType {
	return CannonType
}

type AsteroidStatusComponent struct {
	Destroyed bool
	Size      int
}

func (c AsteroidStatusComponent) Type() ecs.ComponentType {
	return AsteroidStatusType
}

type BulletStatusComponent struct {
	LifeTime time.Time
}

func (c BulletStatusComponent) Type() ecs.ComponentType {
	return BulletStatusType
}

// SYSTEMS

func NewAsteroidSpawnSystem(em *EntityManager) ecs.System {
	return ecs.CollectionSystem(
		func(delta time.Duration, en *ecs.Entity) error {
			p := en.Get(PositionType).(PositionComponent)
			c := en.Get(AsteroidStatusType).(AsteroidStatusComponent)

			if c.Destroyed {
				//fmt.Println("removing dead asteroid", e.Name)

				// spawn new smaller asteroids
				if c.Size > 1 {
					em.createAsteroid(p.Position.X, p.Position.Y, c.Size/2)
					em.createAsteroid(p.Position.X, p.Position.Y, c.Size/2)
					em.createAsteroid(p.Position.X, p.Position.Y, c.Size/2)
				}

				// remove dead asteroid
				em.engine.DeleteEntity(en)
			}

			return nil
		},
		[]ecs.ComponentType{AsteroidStatusType, PositionType},
	)
}

type BulletSystem struct {
	engine  *ecs.Engine
	cannon  ecs.EntityList
	bullets ecs.EntityList

	im *InputManager
	em *EntityManager
}

func NewBulletSystem(im *InputManager, em *EntityManager) *BulletSystem {
	return &BulletSystem{
		im: im,
		em: em,
	}
}

func (s *BulletSystem) AddedToEngine(e *ecs.Engine) error {
	s.engine = e
	s.cannon = e.Collection(PositionType, CannonType)
	s.bullets = e.Collection(BulletStatusType)

	return nil
}

func (s *BulletSystem) RemovedFromEngine(*ecs.Engine) error {
	return nil
}

func (s *BulletSystem) Update(delta time.Duration) error {
	// fire new bullet
	for _, e := range s.cannon.Entities() {
		p := e.Get(PositionType).(PositionComponent)
		c := e.Get(CannonType).(CannonComponent)

		//fmt.Println("controlling", e.Name)

		if s.im.IsDown(c.FireKey) && time.Now().After(c.LastBullet.Add(time.Second/4)) {
			vx := math.Cos(p.Rotation*deg2rad) * c.BulletSpeed
			vy := math.Sin(p.Rotation*deg2rad) * c.BulletSpeed

			s.em.createBullet(p.Position.X, p.Position.Y, vx, vy)
			c.LastBullet = time.Now()
			e.Set(c)
		}

	}

	// remove dead bullets, should be in its own generic lifetime system
	for _, e := range s.bullets.Entities() {
		b := e.Get(BulletStatusType).(BulletStatusComponent)
		if b.LifeTime.Before(time.Now()) {
			s.engine.DeleteEntity(e)
		}
	}

	return nil
}

func NewMotionControlSystem(im *InputManager) ecs.System {
	return ecs.CollectionSystem(
		func(delta time.Duration, en *ecs.Entity) error {

			p := en.Get(PositionType).(PositionComponent)
			m := en.Get(MotionControlType).(MotionControlComponent)
			v := en.Get(VelocityType).(VelocityComponent)

			//fmt.Println("controlling", e.Name)

			if im.IsDown(m.LeftKey) {
				p.Rotation += m.RotationSpeed * delta.Seconds()
			}

			if im.IsDown(m.RightKey) {
				p.Rotation -= m.RotationSpeed * delta.Seconds()
			}

			if im.IsDown(m.AccelerateKey) {
				v.Velocity.X += math.Cos(p.Rotation*deg2rad) * m.AccelerationSpeed * delta.Seconds()
				v.Velocity.Y += math.Sin(p.Rotation*deg2rad) * m.AccelerationSpeed * delta.Seconds()
			}

			if im.IsDown(m.DecelerationKey) {
				v.Velocity.X -= v.Velocity.X * delta.Seconds()
				v.Velocity.Y -= v.Velocity.Y * delta.Seconds()
			}

			speed := math.Sqrt(v.Velocity.X*v.Velocity.X + v.Velocity.Y*v.Velocity.Y)
			if speed > m.MaxVelocity {
				factor := m.MaxVelocity / speed
				v.Velocity.X *= factor
				v.Velocity.Y *= factor
			}

			en.Set(p, v)

			return nil
		},
		[]ecs.ComponentType{PositionType, MotionControlType, VelocityType},
	)
}

type MovementSystem struct {
	engine *ecs.Engine

	minx, maxx,
	miny, maxy float64
}

func NewMovementSystem(minx, maxx, miny, maxy float64) ecs.System {
	return ecs.CollectionSystem(
		func(delta time.Duration, en *ecs.Entity) error {
			p := en.Get(PositionType).(PositionComponent)
			v := en.Get(VelocityType).(VelocityComponent)

			//fmt.Println("moving", e.Name)

			p.Position.X += v.Velocity.X * delta.Seconds()
			p.Position.Y += v.Velocity.Y * delta.Seconds()
			p.Rotation += v.Angular * delta.Seconds()

			// limit position
			if p.Position.X < minx {
				p.Position.X += maxx - minx
			} else if p.Position.X > maxx {
				p.Position.X -= maxx - minx
			}

			if p.Position.Y < miny {
				p.Position.Y += maxy - miny
			} else if p.Position.Y > maxy {
				p.Position.Y -= maxy - miny
			}

			en.Set(p)

			return nil
		},
		[]ecs.ComponentType{PositionType, VelocityType},
	)
}

type RenderSystem struct {
	engine   *ecs.Engine
	drawable ecs.EntityList
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
	return nil
}

func (s *RenderSystem) Update(delta time.Duration) error {
	// init
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.LoadIdentity()

	for _, e := range s.drawable.Entities() {
		p := e.Get(PositionType).(PositionComponent)
		m := e.Get(MeshType).(MeshComponent)
		c := e.Get(ColorType).(ColorComponent)

		//fmt.Println("rendering", e.Name, "at", p)

		gl.LoadIdentity()
		gl.Translated(p.Position.X, p.Position.Y, 0)
		gl.Rotated(p.Rotation-90, 0, 0, 1)
		gl.Color3d(c.R, c.G, c.B)

		gl.Begin(gl.LINE_LOOP)
		for _, point := range m.Points {
			gl.Vertex3d(point.X, point.Y, 0)
		}
		gl.End()
	}

	return nil
}

type CollisionSystem struct {
	engine                    *ecs.Engine
	ships, bullets, asteroids ecs.EntityList
}

func NewCollisionSystem() *CollisionSystem {
	return &CollisionSystem{}
}

func (s *CollisionSystem) AddedToEngine(e *ecs.Engine) error {
	s.engine = e
	s.ships = e.Collection(PositionType, MeshType, ShipStatusType)
	s.bullets = e.Collection(PositionType, MeshType, BulletStatusType)
	s.asteroids = e.Collection(PositionType, MeshType, AsteroidStatusType)

	return nil
}

func (s *CollisionSystem) RemovedFromEngine(*ecs.Engine) error {
	return nil
}

func (s *CollisionSystem) Update(delta time.Duration) error {
	ship := s.ships.First()
	if ship == nil {
		return fmt.Errorf("no ship found for collision system")
	}

	sp := ship.Get(PositionType).(PositionComponent).Position
	sm := ship.Get(MeshType).(MeshComponent).Max

	for _, asteroid := range s.asteroids.Entities() {
		ap := asteroid.Get(PositionType).(PositionComponent).Position
		am := asteroid.Get(MeshType).(MeshComponent).Max

		if sp.Distance(ap) < sm+am {
			//fmt.Println("collision between", ship.Name, "and", asteroid.Name)

			ss := ship.Get(ShipStatusType).(ShipStatusComponent)
			ss.Lifes -= 1
			ship.Set(ss)
		}

		for _, bullet := range s.bullets.Entities() {
			bp := bullet.Get(PositionType).(PositionComponent).Position
			bm := bullet.Get(MeshType).(MeshComponent).Max

			if bp.Distance(ap) < bm+am {
				//fmt.Println("collision between", bullet.Name, "and", asteroid.Name)

				ss := ship.Get(ShipStatusType).(ShipStatusComponent)
				ss.Score += 100
				ship.Set(ss)

				as := asteroid.Get(AsteroidStatusType).(AsteroidStatusComponent)
				as.Destroyed = true
				asteroid.Set(as)

				bs := bullet.Get(BulletStatusType).(BulletStatusComponent)
				bs.LifeTime = time.Time{}
				bullet.Set(bs)
			}
		}
	}

	return nil
}
