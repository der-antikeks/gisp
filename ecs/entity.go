package ecs

import (
	"sync"
)

// General purpose object that consists of a name and a set of components.
type Entity struct {
	Name   string
	engine *Engine
	lock   sync.RWMutex

	// components that define the entity's current state
	components   map[ComponentType]Component
	states       map[string]map[ComponentType]Component
	currentState string
}

// Creates a new entity attach multiple components if supplied
func NewEntity(name string, components ...Component) *Entity {
	en := &Entity{
		Name:       name,
		components: map[ComponentType]Component{},
		states:     map[string]map[ComponentType]Component{},
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

// returns state representation where components can be added
func (en *Entity) State(s string) *state {
	en.lock.Lock()
	defer en.lock.Unlock()

	if _, ok := en.states[s]; !ok {
		en.states[s] = map[ComponentType]Component{}
	}

	return &state{en, en.states[s]}
}

// Change internal state of Entity, adds/removes registered Components
func (en *Entity) ChangeState(s string) {
	// remove Components of old state
	for t := range en.states[en.currentState] {
		en.Remove(t)
	}

	// add Components for new state
	for _, c := range en.states[s] {
		en.Add(c)
	}

	en.currentState = s
}

type state struct {
	entity     *Entity
	components map[ComponentType]Component
}

// Attach Component to State
func (s *state) Add(components ...Component) {
	s.entity.lock.Lock()
	defer s.entity.lock.Unlock()

	for _, c := range components {
		s.components[c.Type()] = c
	}
}

// Remvoe Component from State
func (s *state) Remove(t ComponentType) {
	s.entity.lock.Lock()
	defer s.entity.lock.Unlock()

	delete(s.components, t)
}
