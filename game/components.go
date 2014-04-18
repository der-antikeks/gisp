package game

import (
	"math"
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

const (
	GameStateType ecs.ComponentType = iota

	PositionType
	VelocityType
	MotionControlType

	MeshType
	ColorType
)

type GameStateComponent struct {
	State string
	Since time.Time
}

func (c GameStateComponent) Type() ecs.ComponentType {
	return GameStateType
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
