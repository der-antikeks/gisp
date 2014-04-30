package main

import (
	"log"
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

func main() {
	log.Println("init")

	engine := ecs.NewEngine()
	engine.AddSystem(RenderSystem(), 1)
	engine.AddSystem(MovementSystem(), 0)

	newSpaceship(engine)
	newAsteroid(engine)

	log.Println("update")
	log.Println("error:", engine.Update(time.Duration(1)*time.Second))

	log.Println("fin")
}

// entitites

func newSpaceship(engine *ecs.Engine) *ecs.Entity {
	s := engine.CreateEntity("spaceship")

	position := PositionComponent{}
	position.X = 100 / 2
	position.Y = 100 / 2
	position.Rotation = 0

	display := DisplayComponent{}
	display.View = DisplayObject{} // NewImage(w, h);
	s.Set(position, display)

	return s
}

func newAsteroid(engine *ecs.Engine) *ecs.Entity {
	a := engine.CreateEntity("Asteroid01")
	a.Set(
		PositionComponent{X: 10, Y: 10, Rotation: 12},
		VelocityComponent{VelocityX: -1, VelocityY: -5, AngularVelocity: 1},
		DisplayComponent{View: DisplayObject{}},
	)
	return a
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

func RenderSystem() ecs.System {
	s := ecs.CollectionSystem(
		func(delta time.Duration, en *ecs.Entity) error {
			position := en.Get(PositionType).(PositionComponent)
			display := en.Get(DisplayType).(DisplayComponent)

			display.View.X = position.X
			display.View.Y = position.Y
			display.View.Rotation = position.Rotation
			en.Set(display)

			return nil
		},
		[]ecs.ComponentType{PositionType, DisplayType},
	)

	return s
}

func MovementSystem() ecs.System {
	s := ecs.CollectionSystem(
		func(t time.Duration, en *ecs.Entity) error {
			position := en.Get(PositionType).(PositionComponent)
			velocity := en.Get(VelocityType).(VelocityComponent)

			position.X += velocity.VelocityX * t.Seconds()
			position.Y += velocity.VelocityY * t.Seconds()
			position.Rotation += velocity.AngularVelocity * t.Seconds()
			en.Set(position)

			return nil
		},
		[]ecs.ComponentType{PositionType, VelocityType},
	)

	return s
}
