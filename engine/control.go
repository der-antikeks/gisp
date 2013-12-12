package engine

import (
	//"fmt"
	m "math"
	"time"

	"github.com/der-antikeks/gisp/math"

	glfw "github.com/go-gl/glfw3"
)

type Control interface {
	OnWindowResize(w, h float64)
	OnMouseMove(x, y float64)
	OnMouseScroll(x, y float64)
	OnMouseButton(b glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
	OnKeyPress(key glfw.Key, action glfw.Action, mods glfw.ModifierKey)

	Update(delta time.Duration)
}

type FlyControl struct {
	Control

	camera Camera

	cameraMoveSpeed   float64
	cameraRotateSpeed float64

	// input cache
	horizontalAngle, verticalAngle float64
	MouseX, MouseY, OldX, OldY     float64

	moveForward, moveBack, moveLeft, moveRight, moveUp, moveDown bool
}

func NewFlyControl(camera Camera) *FlyControl {
	return &FlyControl{
		camera: camera,

		cameraMoveSpeed:   5.0,
		cameraRotateSpeed: 0.1,
	}
}

func (c *FlyControl) OnWindowResize(w, h float64) {
	if pcam, ok := c.camera.(*PerspectiveCamera); ok {
		pcam.SetAspect(w / h)
	}
}

func (c *FlyControl) OnMouseMove(x, y float64) {
	c.MouseX = x
	c.MouseY = y
}

func (c *FlyControl) OnMouseScroll(x, y float64) {
	//c.camera.SetFov(45.0 - 5.0*y)
}

func (c *FlyControl) OnMouseButton(b glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {}

func (c *FlyControl) OnKeyPress(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	var pressed bool
	if action == glfw.Press || action == glfw.Repeat {
		pressed = true
	}

	switch key {
	case glfw.KeyUp, glfw.KeyW:
		c.moveForward = pressed
	case glfw.KeyDown, glfw.KeyS:
		c.moveBack = pressed
	case glfw.KeyLeft, glfw.KeyA:
		c.moveLeft = pressed
	case glfw.KeyRight, glfw.KeyD:
		c.moveRight = pressed
	case glfw.KeySpace:
		c.moveUp = pressed
	case glfw.KeyLeftControl:
		c.moveDown = pressed
	}
}

func (c *FlyControl) Update(delta time.Duration) {
	// Orientation
	diffX := c.MouseX - c.OldX
	c.OldX = c.MouseX
	c.horizontalAngle += c.cameraRotateSpeed * delta.Seconds() * -diffX // lon, theta

	diffY := c.MouseY - c.OldY
	c.OldY = c.MouseY
	c.verticalAngle += c.cameraRotateSpeed * delta.Seconds() * -diffY // lat, phi

	// Direction, Spherical coordinates to Cartesian coordinates conversion
	direction := math.Vector{
		m.Cos(c.verticalAngle) * m.Sin(c.horizontalAngle),
		m.Sin(c.verticalAngle),
		m.Cos(c.verticalAngle) * m.Cos(c.horizontalAngle),
	}

	// Right vector
	right := math.Vector{
		m.Sin(c.horizontalAngle - math.Pi/2.0),
		0,
		m.Cos(c.horizontalAngle - math.Pi/2.0),
	}

	// Up vector, perpendicular to both direction and right
	up := right.Cross(direction)
	c.camera.SetUp(up)

	// Position
	actualMoveSpeed := delta.Seconds() * c.cameraMoveSpeed

	if c.moveForward {
		//this.object.translateZ(-actualMoveSpeed)
		c.camera.SetPosition(c.camera.Position().Add(direction.MulScalar(actualMoveSpeed)))
	}
	if c.moveBack {
		//this.object.translateZ(actualMoveSpeed)
		c.camera.SetPosition(c.camera.Position().Sub(direction.MulScalar(actualMoveSpeed)))
	}
	if c.moveLeft {
		//this.object.translateX(-actualMoveSpeed)
		c.camera.SetPosition(c.camera.Position().Sub(right.MulScalar(actualMoveSpeed)))
	}
	if c.moveRight {
		//this.object.translateX(actualMoveSpeed)
		c.camera.SetPosition(c.camera.Position().Add(right.MulScalar(actualMoveSpeed)))
	}
	if c.moveUp {
		//this.object.translateY(actualMoveSpeed)
		c.camera.SetPosition(c.camera.Position().Add(math.Vector{0, 1, 0}.MulScalar(actualMoveSpeed)))
	}
	if c.moveDown {
		//this.object.translateY(-actualMoveSpeed)
		c.camera.SetPosition(c.camera.Position().Sub(math.Vector{0, 1, 0}.MulScalar(actualMoveSpeed)))
	}

	c.camera.LookAt(c.camera.Position().Add(direction))
}

type OrbitControl struct {
	Control

	camera Camera
	target math.Vector

	cameraMoveSpeed   float64
	cameraRotateSpeed float64
	cameraZoomSpeed   float64

	width, height float64
	minDistance   float64
	maxDistance   float64

	// input state
	dragging    bool
	needsUpdate bool

	// input cache
	oldX, oldY     float64
	deltaX, deltaY float64
	zoom           float64
}

func NewOrbitControl(camera Camera) *OrbitControl {
	return &OrbitControl{
		camera: camera,

		cameraMoveSpeed:   5.0,
		cameraRotateSpeed: 1.0,
		cameraZoomSpeed:   1.0,

		width:  640,
		height: 480,

		minDistance: 5.0,
		maxDistance: m.Inf(1),

		dragging:    false,
		needsUpdate: true,

		oldX: -1,
		oldY: -1,
	}
}

func (c *OrbitControl) SetTarget(t math.Vector) {
	c.target = t
	c.needsUpdate = true
}

func (c *OrbitControl) OnWindowResize(w, h float64) {
	c.width, c.height = w, h
	c.needsUpdate = true
}

func (c *OrbitControl) OnMouseScroll(x, y float64) {
	c.zoom -= c.cameraZoomSpeed * y
	c.needsUpdate = true
}

func (c *OrbitControl) OnMouseButton(b glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if b == glfw.MouseButton2 {
		if action == glfw.Press || action == glfw.Repeat {
			c.dragging = true
		} else {
			c.dragging = false
			c.oldX, c.oldY = -1, -1
		}
	}
}

func (c *OrbitControl) OnMouseMove(x, y float64) {
	if c.dragging {
		if c.oldX < 0 || c.oldY < 0 {
			c.oldX, c.oldY = x, y
		}

		c.deltaX += c.oldX - x
		c.deltaY += c.oldY - y
		c.oldX, c.oldY = x, y

		c.needsUpdate = true
	}
}

func (c *OrbitControl) OnKeyPress(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Press && action != glfw.Repeat {
		return
	}

	switch key {
	case glfw.KeyUp, glfw.KeyW:
		c.deltaY -= c.cameraMoveSpeed
		c.needsUpdate = true

	case glfw.KeyDown, glfw.KeyS:
		c.deltaY += c.cameraMoveSpeed
		c.needsUpdate = true

	case glfw.KeyLeft, glfw.KeyA:
		c.deltaX -= c.cameraMoveSpeed
		c.needsUpdate = true

	case glfw.KeyRight, glfw.KeyD:
		c.deltaX += c.cameraMoveSpeed
		c.needsUpdate = true

	case glfw.KeyKpAdd:
		c.zoom -= c.cameraZoomSpeed
		c.needsUpdate = true

	case glfw.KeyKpSubtract:
		c.zoom += c.cameraZoomSpeed
		c.needsUpdate = true

	}
}

const EPS = 0.000001

func (c *OrbitControl) Update(delta time.Duration) {
	if !c.needsUpdate {
		return
	}

	offset := c.camera.Position().Sub(c.target)

	// rotating across whole screen goes 360 degrees around
	// angle from z-axis around y-axis
	theta := m.Atan2(offset[0], offset[2])
	theta += 2.0 * math.Pi * c.deltaX / c.width * c.cameraRotateSpeed

	// rotating up and down along whole screen attempts to go 360, but limited to 180
	// angle from y-axis
	phi := m.Atan2(m.Sqrt(offset[0]*offset[0]+offset[2]*offset[2]), offset[1])
	phi += 2.0 * math.Pi * c.deltaY / c.height * c.cameraRotateSpeed

	// restrict phi to be between desired limits
	minPolarAngle := 0.0
	maxPolarAngle := math.Pi
	phi = m.Max(minPolarAngle, m.Min(maxPolarAngle, phi))

	// restrict phi to be betwee EPS and PI-EPS
	phi = m.Max(EPS, m.Min(math.Pi-EPS, phi))

	// restrict radius to be between desired limits
	radius := offset.Length() + c.zoom
	radius = m.Max(c.minDistance, m.Min(c.maxDistance, radius))

	c.camera.SetPosition(c.target.Add(math.Vector{
		radius * m.Sin(phi) * m.Sin(theta),
		radius * m.Cos(phi),
		radius * m.Sin(phi) * m.Cos(theta),
	}))
	c.camera.LookAt(c.target)

	// set aspect of perspective camera
	if pcam, ok := c.camera.(*PerspectiveCamera); ok {
		pcam.SetAspect(c.width / c.height)
	}

	// reset
	c.needsUpdate = false
	c.deltaX, c.deltaY = 0, 0
	c.zoom = 0
}
