package ecs

import ()

// General purpose object that consists of a name and a set of components.
type Entity struct {
	Name   string
	engine *Engine

	// components that define the entity's current state
	components set
}

func NewEntity(name string) *Entity {
	return &Entity{
		Name:       name,
		components: newset(),
	}
}

// Set Engine that gets notified of component changes
func (e *Entity) setEngine(engine *Engine) {
	e.engine = engine
}

// Attach Component
func (e *Entity) Add(component interface{}) {
	e.components.Add(component)

	if e.engine != nil {
		e.engine.entityUpdated(e)
	}
}

// Detach Component
func (e *Entity) Remove(component interface{}) {
	e.components.Remove(component)

	if e.engine != nil {
		e.engine.entityUpdated(e)
	}
}

// Get Component
func (e *Entity) Get(component interface{}) interface{} {
	return e.components.Get(component)
}

// HandleMessage(msg Message) echoes to all attached components
