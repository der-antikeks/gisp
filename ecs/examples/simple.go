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

	ship := newSpaceship()
	engine.AddEntity(ship)

	as1 := newAsteroid()
	engine.AddEntity(as1)

	log.Println("update")
	log.Println("error:", engine.Update(time.Duration(1)*time.Second))

	log.Println("fin")
}

// entitites

func newSpaceship() *ecs.Entity {
	s := ecs.NewEntity("spaceship")

	position := &PositionComponent{}
	position.X = 100 / 2
	position.Y = 100 / 2
	position.Rotation = 0
	s.Add(position)

	display := &DisplayComponent{}
	display.View = DisplayObject{} // NewImage(w, h);
	s.Add(display)

	return s
}

func newAsteroid() *ecs.Entity {
	return ecs.NewEntity(
		"Asteroid01",
		&PositionComponent{X: 10, Y: 10, Rotation: 12},
		&VelocityComponent{VelocityX: -1, VelocityY: -5, AngularVelocity: 1},
		&DisplayComponent{View: DisplayObject{}},
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

func RenderSystem() ecs.System {
	s := ecs.CollectionSystem(
		func(delta time.Duration, en *ecs.Entity) {
			position := en.Get(PositionType).(*PositionComponent)
			display := en.Get(DisplayType).(*DisplayComponent)

			display.View.X = position.X
			display.View.Y = position.Y
			display.View.Rotation = position.Rotation
		},
		[]ecs.ComponentType{PositionType, DisplayType},
	)

	/*
		s.SetPreUpdateFunc(func() {
			log.Println("initialize render system update loop")
		})

		s.SetInitFunc(func() error {
			log.Println("initializing RenderSystem")
			return nil
		})

		s.SetCleanupFunc(func() error {
			log.Println("cleaning RenderSystem")
			return nil
		})
	*/

	return s
}

func MovementSystem() ecs.System {
	s := ecs.CollectionSystem(
		func(t time.Duration, en *ecs.Entity) {
			position := en.Get(PositionType).(*PositionComponent)
			velocity := en.Get(VelocityType).(*VelocityComponent)

			position.X += velocity.VelocityX * t.Seconds()
			position.Y += velocity.VelocityY * t.Seconds()
			position.Rotation += velocity.AngularVelocity * t.Seconds()
		},
		[]ecs.ComponentType{PositionType, VelocityType},
	)

	/*
		s.SetInitFunc(func() error {
			log.Println("initializing MovementSystem")
			return nil
		})

		s.SetCleanupFunc(func() error {
			log.Println("cleaning MovementSystem")
			return nil
		})
	*/

	return s
}
