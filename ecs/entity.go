package ecs

import ()

type Entity struct {
	Name   string
	engine *Engine

	// components that define the entity's current state
	Components Set
}

func NewEntity(name string) *Entity {
	return &Entity{
		Name:       name,
		Components: NewSet(),
	}
}

// Set Engine that gets notified of component changes
func (e *Entity) SetEngine(engine *Engine) {
	e.engine = engine
}

// Attach Component
func (e *Entity) Add(component interface{}) {
	e.Components.Add(component)

	if e.engine != nil {
		e.engine.EntityUpdated(e)
	}
}

// Detach Component
func (e *Entity) Remove(component interface{}) {
	e.Components.Remove(component)

	if e.engine != nil {
		e.engine.EntityUpdated(e)
	}
}

// func (e *Entity) Get(component interface{}) interface{} {

// HandleMessage(msg Message) echoes to all attached components
