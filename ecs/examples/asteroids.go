package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
)

var (
	camZ    = -1000.0
	objects []Object
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

	glfw.SetErrorCallback(func(err glfw.ErrorCode, desc string) {
		log.Fatalln("error callback:", err, desc)
	})

	if !glfw.Init() {
		log.Fatalln("failed to initialize glfw")
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, 1)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		log.Fatalln("create window:", err)
	}

	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	gl.Init()

	window.SetKeyCallback(onKey)
	window.SetFramebufferSizeCallback(onResize)

	if err := initScene(); err != nil {
		log.Fatalln("init scene:", err)
	}
	defer destroyScene()

	w, h := window.GetFramebufferSize()
	onResize(window, w, h)

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

			// animate
			updateScene(delta)

			// draw
			drawScene()

			// Swap buffers
			window.SwapBuffers()
			glfw.PollEvents()
		}
	}
}

func onKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Press {
		return
	}

	switch glfw.Key(key) {
	case glfw.KeyEscape:
		w.SetShouldClose(true)
	case glfw.KeyW, glfw.KeyUp:
	case glfw.KeyS, glfw.KeyDown:
	case glfw.KeyA, glfw.KeyLeft:
	case glfw.KeyD, glfw.KeyRight:
	case glfw.KeySpace:
	default:
		return
	}
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

func initScene() error {
	// gl init
	gl.ShadeModel(gl.SMOOTH)
	gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)

	gl.ClearColor(0.1, 0.1, 0.1, 0.0)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)
	gl.Enable(gl.DEPTH_TEST)

	gl.LineWidth(1)
	gl.Enable(gl.LINE_SMOOTH)

	// load objects
	objects = append(objects, newSpaceship())

	for i := 0; i < 3; i++ {
		objects = append(objects, newAsteroid(rand.Intn(4)+4))
	}

	return nil
}

// components: position, velocity, polygon(collision), keymovable
func newSpaceship() Object {
	o := Object{
		Position: Position{0, 0, 45},
		Velocity: Velocity{-1 * MaxShipSpeed, 1 * MaxShipSpeed, 0},
		Color:    Color{1, 1, 1},
		Points: []Point{
			Point{-10, -15},
			Point{0, -10},
			Point{10, -15},
			Point{0, 15},
		},
	}

	return o
}

// components: position, velocity, polygon(collision), size
func newAsteroid(size int) Object {
	o := Object{
		Color: Color{1, 1, 0},
	}

	// rotation
	rot := rand.Float64() * 360
	o.Position.Rotation = rot

	// velocity
	rad := (rot + 90) * deg2rad
	speed := rand.Float64() * MaxAsteroidSpeed
	o.Velocity.X = speed * math.Cos(rad)
	o.Velocity.Y = speed * math.Sin(rad)

	// shape
	o.Points = make([]Point, 7)
	step := (2.0 * math.Pi) / float64(len(o.Points))
	max := float64(size * 10)
	min := max / 2

	for i := range o.Points {
		length := (rand.Float64() * (max - min)) + min
		angle := float64(i) * step

		o.Points[i].X = length * math.Cos(angle)
		o.Points[i].Y = length * math.Sin(angle)
	}

	return o
}

// components: position, velocity, polygon(collision), lifetime
func newBullet() Object {
	return Object{
		Position: Position{0, 0, 0},
		Velocity: Velocity{-1 * MaxShipSpeed, 1 * MaxShipSpeed, 0},
		Color:    Color{1, 1, 0},
		Points: []Point{
			Point{-1, -5},
			Point{1, -5},
			Point{1, 5},
			Point{-1, 5},
		},
	}
}

func destroyScene() {
	objects = []Object{}
}

// velocity system
func updateScene(delta time.Duration) {
	for i := range objects {
		objects[i].Position.X += objects[i].Velocity.X * delta.Seconds()
		objects[i].Position.Y += objects[i].Velocity.Y * delta.Seconds()
		objects[i].Position.Rotation += objects[i].Velocity.Angular * delta.Seconds()
	}
}

// entity
type Object struct {
	Position Position
	Velocity Velocity

	Color  Color
	Points []Point
}

// components
type Point struct {
	X, Y float64
}

type Position struct {
	X, Y     float64
	Rotation float64
}

type Velocity struct {
	X, Y    float64
	Angular float64
}

type Color struct {
	R, G, B float64
}

// render system
func drawScene() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.LoadIdentity()

	for _, o := range objects {
		drawPoly(o)
	}
}

func drawPoly(poly Object) {
	gl.LoadIdentity()
	gl.Translated(poly.Position.X, poly.Position.Y, camZ)
	gl.Rotated(poly.Position.Rotation, 0, 0, 1)
	gl.Color3d(poly.Color.R, poly.Color.G, poly.Color.B)

	gl.Begin(gl.LINE_LOOP)
	for _, p := range poly.Points {
		gl.Vertex3d(p.X, p.Y, 0)
	}
	gl.End()
}
