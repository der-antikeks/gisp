package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"

	"github.com/der-antikeks/gisp/ecs"
)

const (
	PriorityBeforeRender ecs.SystemPriority = iota
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

	engine := ecs.NewEngine()

	// managers
	im := NewInputManager()
	em := NewEntityManager(engine)
	w, h := 640, 480
	wm := NewWindowManager(w, h, "Testing", im)
	defer wm.cleanup()

	// systems
	maxx, maxy := float64(w)*0.5, float64(h)*0.5
	newMovementSystem(engine, -maxx, maxx, -maxy, maxy)
	newRenderSystem(engine, wm)
	newCollisionSystem(engine)
	newAsteroidSpawnSystem(engine, em)
	newMotionControlSystem(engine, im)
	newBulletSystem(engine, im, em)

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
			engine.Publish(ecs.MessageUpdate{Delta: delta})
		}
	}
}

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

type WindowManager struct {
	width, height int
	window        *glfw.Window
}

func NewWindowManager(w, h int, title string, im *InputManager) *WindowManager {
	m := &WindowManager{
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
	})

	return m
}

func (m *WindowManager) isRunning() bool {
	return !m.window.ShouldClose()
}

func (m *WindowManager) update() {
	MainThread(func() {
		m.window.SwapBuffers()
		glfw.PollEvents()
	})
}

func (m *WindowManager) cleanup() {
	MainThread(func() {
		glfw.Terminate()
	})
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

func (em *EntityManager) CreateGame() ecs.Entity {
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

	s := em.engine.Entity()
	em.engine.Set(
		s,
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

	a := em.engine.Entity()
	em.engine.Set(
		a,
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

	em.engine.Set(a, mc)
}

func (em *EntityManager) createBullet(x, y, vx, vy float64) {
	b := em.engine.Entity()
	em.engine.Set(
		b,
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

func newAsteroidSpawnSystem(engine *ecs.Engine, em *EntityManager) {
	ecs.SingleAspectSystem(
		engine, PriorityBeforeRender,
		func(delta time.Duration, en ecs.Entity) {
			ec, err := engine.Get(en, PositionType)
			if err != nil {
				return
			}
			p := ec.(PositionComponent)
			ec, err = engine.Get(en, AsteroidStatusType)
			if err != nil {
				return
			}
			c := ec.(AsteroidStatusComponent)

			if c.Destroyed {
				//fmt.Println("removing dead asteroid", e.Name)

				// spawn new smaller asteroids
				if c.Size > 1 {
					em.createAsteroid(p.Position.X, p.Position.Y, c.Size/2)
					em.createAsteroid(p.Position.X, p.Position.Y, c.Size/2)
					em.createAsteroid(p.Position.X, p.Position.Y, c.Size/2)
				}

				// remove dead asteroid
				engine.Delete(en)
			}
		},
		[]ecs.ComponentType{AsteroidStatusType, PositionType},
	)
}

func newBulletSystem(engine *ecs.Engine, im *InputManager, em *EntityManager) {
	cannonChan, bulletChan, eventChan := make(chan ecs.Message), make(chan ecs.Message), make(chan ecs.Message)
	engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{PositionType, CannonType},
	}, PriorityBeforeRender, cannonChan)
	engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{BulletStatusType},
	}, PriorityBeforeRender, bulletChan)
	engine.Subscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, PriorityBeforeRender, eventChan)

	cannons := []ecs.Entity{}
	bullets := []ecs.Entity{}

	go func() {
		for {
			select {
			case event := <-cannonChan:
				switch e := event.(type) {
				case ecs.MessageEntityAdd:
					cannons = append(cannons, e.Added)
				case ecs.MessageEntityRemove:
					for i, f := range cannons {
						if f == e.Removed {
							cannons = append(cannons[:i], cannons[i+1:]...)
						}
					}
				}

			case event := <-bulletChan:
				switch e := event.(type) {
				case ecs.MessageEntityAdd:
					bullets = append(bullets, e.Added)
				case ecs.MessageEntityRemove:
					for i, f := range bullets {
						if f == e.Removed {
							bullets = append(bullets[:i], bullets[i+1:]...)
						}
					}
				}

			case event := <-eventChan:
				switch event.(type) {
				case ecs.MessageUpdate:

					// fire new bullet
					for _, e := range cannons {
						ec, err := engine.Get(e, PositionType)
						if err != nil {
							continue
						}
						p := ec.(PositionComponent)
						ec, err = engine.Get(e, CannonType)
						if err != nil {
							continue
						}
						c := ec.(CannonComponent)

						//fmt.Println("controlling", e.Name)

						if im.IsDown(c.FireKey) && time.Now().After(c.LastBullet.Add(time.Second/4)) {
							vx := math.Cos(p.Rotation*deg2rad) * c.BulletSpeed
							vy := math.Sin(p.Rotation*deg2rad) * c.BulletSpeed

							em.createBullet(p.Position.X, p.Position.Y, vx, vy)
							c.LastBullet = time.Now()
							engine.Set(e, c)
						}

					}

					// remove dead bullets, should be in its own generic lifetime system
					for _, e := range bullets {
						ec, err := engine.Get(e, BulletStatusType)
						if err != nil {
							continue
						}
						b := ec.(BulletStatusComponent)
						if b.LifeTime.Before(time.Now()) {
							engine.Delete(e)
						}
					}

				}
			}
		}
	}()
}

func newMotionControlSystem(engine *ecs.Engine, im *InputManager) {
	ecs.SingleAspectSystem(
		engine, PriorityBeforeRender,
		func(delta time.Duration, en ecs.Entity) {

			ec, err := engine.Get(en, PositionType)
			if err != nil {
				return
			}
			p := ec.(PositionComponent)
			ec, err = engine.Get(en, MotionControlType)
			if err != nil {
				return
			}
			m := ec.(MotionControlComponent)
			ec, err = engine.Get(en, VelocityType)
			if err != nil {
				return
			}
			v := ec.(VelocityComponent)

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

			engine.Set(en, p, v)
		},
		[]ecs.ComponentType{PositionType, MotionControlType, VelocityType},
	)
}

func newMovementSystem(engine *ecs.Engine, minx, maxx, miny, maxy float64) {
	ecs.SingleAspectSystem(
		engine, PriorityBeforeRender,
		func(delta time.Duration, en ecs.Entity) {
			ec, err := engine.Get(en, PositionType)
			if err != nil {
				return
			}
			p := ec.(PositionComponent)
			ec, err = engine.Get(en, VelocityType)
			if err != nil {
				return
			}
			v := ec.(VelocityComponent)

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

			engine.Set(en, p)
		},
		[]ecs.ComponentType{PositionType, VelocityType},
	)
}

func newRenderSystem(engine *ecs.Engine, wm *WindowManager) {
	c := make(chan ecs.Message)
	engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{PositionType, MeshType, ColorType},
	}, PriorityRender, c)
	engine.Subscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, PriorityRender, c)

	drawable := []ecs.Entity{}

	go func() {
		for event := range c {
			switch e := event.(type) {
			case ecs.MessageEntityAdd:
				drawable = append(drawable, e.Added)
			case ecs.MessageEntityRemove:
				for i, f := range drawable {
					if f == e.Removed {
						drawable = append(drawable[:i], drawable[i+1:]...)
					}
				}

			case ecs.MessageUpdate:
				// init
				MainThread(func() {
					gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
					gl.LoadIdentity()
				})

				for _, e := range drawable {
					ec, err := engine.Get(e, PositionType)
					if err != nil {
						continue
					}
					p := ec.(PositionComponent)
					ec, err = engine.Get(e, MeshType)
					if err != nil {
						continue
					}
					m := ec.(MeshComponent)
					ec, err = engine.Get(e, ColorType)
					c := ec.(ColorComponent)

					//fmt.Println("rendering", e.Name, "at", p)

					MainThread(func() {
						gl.LoadIdentity()
						gl.Translated(p.Position.X, p.Position.Y, 0)
						gl.Rotated(p.Rotation-90, 0, 0, 1)
						gl.Color3d(c.R, c.G, c.B)

						gl.Begin(gl.LINE_LOOP)
						for _, point := range m.Points {
							gl.Vertex3d(point.X, point.Y, 0)
						}
						gl.End()
					})
				}

				// Swap buffers
				wm.update()
			}
		}
	}()
}

func newCollisionSystem(engine *ecs.Engine) {
	shipChan, bulletChan, asteroidChan := make(chan ecs.Message), make(chan ecs.Message), make(chan ecs.Message)
	engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{PositionType, MeshType, ShipStatusType},
	}, PriorityAfterRender, shipChan)
	engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{PositionType, MeshType, BulletStatusType},
	}, PriorityAfterRender, bulletChan)
	engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{PositionType, MeshType, AsteroidStatusType},
	}, PriorityAfterRender, asteroidChan)

	eventChan := make(chan ecs.Message)
	engine.Subscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, PriorityAfterRender, eventChan)

	var ship ecs.Entity = -1
	var bullets, asteroids ecs.SliceEntityList

	go func() {
		for {
			select {
			case event := <-shipChan:
				switch e := event.(type) {
				case ecs.MessageEntityAdd:
					ship = e.Added
				case ecs.MessageEntityRemove:
					ship = -1
				}

			case event := <-bulletChan:
				switch e := event.(type) {
				case ecs.MessageEntityAdd:
					bullets.Add(e.Added)
				case ecs.MessageEntityRemove:
					bullets.Remove(e.Removed)
				}

			case event := <-asteroidChan:
				switch e := event.(type) {
				case ecs.MessageEntityAdd:
					asteroids.Add(e.Added)
				case ecs.MessageEntityRemove:
					asteroids.Remove(e.Removed)
				}

			case event := <-eventChan:
				switch event.(type) {
				case ecs.MessageUpdate:

					if ship == -1 {
						log.Fatalf("no ship found for collision system")
					}

					ec, err := engine.Get(ship, PositionType)
					if err != nil {
						continue
					}
					sp := ec.(PositionComponent).Position
					ec, err = engine.Get(ship, MeshType)
					if err != nil {
						continue
					}
					sm := ec.(MeshComponent).Max

					for _, asteroid := range asteroids.Entities() {
						ec, err := engine.Get(asteroid, PositionType)
						if err != nil {
							continue
						}
						ap := ec.(PositionComponent).Position
						ec, err = engine.Get(asteroid, MeshType)
						if err != nil {
							continue
						}
						am := ec.(MeshComponent).Max

						if sp.Distance(ap) < sm+am {
							//fmt.Println("collision between", ship.Name, "and", asteroid.Name)

							ec, err := engine.Get(ship, ShipStatusType)
							if err != nil {
								continue
							}
							ss := ec.(ShipStatusComponent)
							ss.Lifes -= 1
							engine.Set(ship, ss)
						}

						for _, bullet := range bullets.Entities() {
							ec, err := engine.Get(bullet, PositionType)
							if err != nil {
								continue
							}
							bp := ec.(PositionComponent).Position
							ec, err = engine.Get(bullet, MeshType)
							if err != nil {
								continue
							}
							bm := ec.(MeshComponent).Max

							if bp.Distance(ap) < bm+am {
								//fmt.Println("collision between", bullet.Name, "and", asteroid.Name)

								ec, err := engine.Get(ship, ShipStatusType)
								if err != nil {
									continue
								}
								ss := ec.(ShipStatusComponent)
								ss.Score += 100
								engine.Set(ship, ss)

								ec, err = engine.Get(asteroid, AsteroidStatusType)
								if err != nil {
									continue
								}
								as := ec.(AsteroidStatusComponent)
								as.Destroyed = true
								engine.Set(asteroid, as)

								ec, err = engine.Get(bullet, BulletStatusType)
								if err != nil {
									continue
								}
								bs := ec.(BulletStatusComponent)
								bs.LifeTime = time.Time{}
								engine.Set(bullet, bs)
							}
						}
					}

				}
			}
		}
	}()
}
