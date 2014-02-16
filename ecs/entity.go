package ecs

import (
	"reflect"
	"sync"
)

// General purpose object that consists of a name and a set of components.
type Entity struct {
	sync.Mutex

	Name   string
	engine *Engine

	// components that define the entity's current state
	components map[reflect.Type]interface{}
}

func NewEntity(name string, components ...interface{}) *Entity {
	e := &Entity{
		Name:       name,
		components: make(map[reflect.Type]interface{}),
	}

	for _, component := range components {
		e.Add(component)
	}

	return e
}

// Set Engine that gets notified of component changes
func (e *Entity) setEngine(engine *Engine) {
	e.engine = engine
}

func componentType(component interface{}) reflect.Type {
	if c, ok := component.(reflect.Type); ok {
		return c
	}

	return reflect.TypeOf(component) // .Elem()
}

// Attach Component
func (e *Entity) Add(component interface{}) {
	t := componentType(component)

	e.Lock()
	e.components[t] = component
	e.Unlock()

	if e.engine != nil {
		e.engine.entityUpdated(e)
	}
}

// Detach Component
func (e *Entity) Remove(component interface{}) {
	t := componentType(component)

	e.Lock()
	delete(e.components, t)
	e.Unlock()

	if e.engine != nil {
		e.engine.entityUpdated(e)
	}
}

// Get Component
func (e *Entity) Get(component interface{}) interface{} {
	t := componentType(component)

	e.Lock()
	defer e.Unlock()

	if r, ok := e.components[t]; ok && r != nil {
		return r
	}
	return nil
}

// Remove all components of Entity
func (e *Entity) reset() {
	e.Lock()
	defer e.Unlock()

	for t := range e.components {
		delete(e.components, t)
	}
}

// HandleMessage(msg Message) echoes to all attached components

/*
// If A and B are sets and every element of A is also an element of B
func (a set) SubsetOf(b set) (set, bool) {
	a.Lock()
	defer a.Unlock()

	b.Lock()
	defer b.Unlock()

	subset := Newset()

	for t := range a.data {
		if b.data[t] == nil {
			return set{}, false
		}

		subset.data[t] = b.data[t]
	}

	return subset, true
}
*/
