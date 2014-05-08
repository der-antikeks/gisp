package main

import (
	"log"
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

func main() {
	log.Println("init")

	engine := ecs.NewEngine()
	renderSystem(engine)
	movementSystem(engine)

	newSpaceship(engine)
	newAsteroid(engine)

	log.Println("update")
	engine.Publish(ecs.MessageUpdate{Delta: time.Duration(1) * time.Second})

	log.Println("fin")
}

// entitites

func newSpaceship(engine *ecs.Engine) {
	s := engine.CreateEntity("spaceship")

	position := PositionComponent{}
	position.X = 100 / 2
	position.Y = 100 / 2
	position.Rotation = 0

	display := DisplayComponent{}
	display.View = DisplayObject{} // NewImage(w, h);
	engine.SetComponents(s, position, display)
}

func newAsteroid(engine *ecs.Engine) {
	a := engine.CreateEntity("Asteroid01")
	engine.SetComponents(
		a,
		PositionComponent{X: 10, Y: 10, Rotation: 12},
		VelocityComponent{VelocityX: -1, VelocityY: -5, AngularVelocity: 1},
		DisplayComponent{View: DisplayObject{}},
	)
}

// components
const (
	PositionType ecs.ComponentType = iota
	VelocityType
	DisplayType
)

type PositionComponent struct {
	X, Y     float64
	Rotation float64
}

func (c PositionComponent) Type() ecs.ComponentType {
	return PositionType
}

type VelocityComponent struct {
	VelocityX, VelocityY float64
	AngularVelocity      float64
}

func (c VelocityComponent) Type() ecs.ComponentType {
	return VelocityType
}

type DisplayObject struct {
	X, Y     float64
	Rotation float64
}

type DisplayComponent struct {
	View DisplayObject
}

func (c DisplayComponent) Type() ecs.ComponentType {
	return DisplayType
}

// systems

const (
	PriorityBeforeRender ecs.SystemPriority = iota
	PriorityRender
)

func MustGetComponent(c ecs.Component, err error) ecs.Component {
	if err != nil {
		panic(err)
	}
	return c
}

func renderSystem(engine *ecs.Engine) {
	ecs.SingleAspectSystem(
		engine, PriorityRender,
		func(delta time.Duration, en ecs.Entity) {
			position := MustGetComponent(engine.GetComponent(en, PositionType)).(PositionComponent)
			display := MustGetComponent(engine.GetComponent(en, DisplayType)).(DisplayComponent)

			display.View.X = position.X
			display.View.Y = position.Y
			display.View.Rotation = position.Rotation
			engine.SetComponents(en, display)
		},
		[]ecs.ComponentType{PositionType, DisplayType},
	)
}

func movementSystem(engine *ecs.Engine) {
	ecs.SingleAspectSystem(
		engine, PriorityBeforeRender,
		func(t time.Duration, en ecs.Entity) {
			position := MustGetComponent(engine.GetComponent(en, PositionType)).(PositionComponent)
			velocity := MustGetComponent(engine.GetComponent(en, VelocityType)).(VelocityComponent)

			position.X += velocity.VelocityX * t.Seconds()
			position.Y += velocity.VelocityY * t.Seconds()
			position.Rotation += velocity.AngularVelocity * t.Seconds()
			engine.SetComponents(en, position)
		},
		[]ecs.ComponentType{PositionType, VelocityType},
	)
}
