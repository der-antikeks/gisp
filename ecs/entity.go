package ecs

import (
	"sync"
)

// General purpose object that consists of a name and a set of components.
type Entity struct {
	Name   string
	engine *Engine

	// components that define the entity's current state
	components map[ComponentType]Component
	lock       sync.RWMutex
}

// Creates a new entity attach multiple components if supplied
func NewEntity(name string, components ...Component) *Entity {
	en := &Entity{
		Name:       name,
		components: map[ComponentType]Component{},
	}

	for _, c := range components {
		en.Add(c)
	}

	return en
}

// Attach Component to the Entity
func (en *Entity) Add(c Component) {
	en.lock.Lock()
	en.components[c.Type()] = c
	en.lock.Unlock()

	if en.engine != nil {
		en.engine.entityAddedComponent(en, c)
	}
}

// Detach Component from Entity
func (en *Entity) Remove(t ComponentType) {
	en.lock.Lock()
	c := en.components[t]
	delete(en.components, t)
	en.lock.Unlock()

	if en.engine != nil {
		en.engine.entityRemovedComponent(en, c)
	}
}

// Get specific Component of Entity
func (en *Entity) Get(t ComponentType) Component {
	en.lock.RLock()
	c, found := en.components[t]
	en.lock.RUnlock()

	if found {
		return c
	}
	return nil
}

// Remove all components of Entity
func (en *Entity) reset() {
	en.lock.Lock()
	defer en.lock.Unlock()

	for t := range en.components {
		delete(en.components, t)
	}
}

// helper function, calls engine.RemoveEntity(en)
func (en *Entity) Delete() {
	if en.engine != nil {
		en.engine.RemoveEntity(en)
	}
}
