package ecs

import ()

// General purpose object that consists of a name and a set of components.
type Entity struct {
	Name   string // TODO: only for debugging
	engine *Engine

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

	en.Add(components...)
	return en
}

// Attach Component to the Entity
func (en *Entity) Add(components ...Component) {
	for _, c := range components {
		en.components[c.Type()] = c
	}

	if en.engine != nil {
		en.engine.entityAddedComponent(en)
	}
}

// Detach Component from Entity
func (en *Entity) Remove(t ComponentType) {
	if _, found := en.components[t]; !found {
		return
	}

	delete(en.components, t)

	if en.engine != nil {
		en.engine.entityRemovedComponent(en)
	}
}

// Get specific Component of Entity
func (en *Entity) Get(t ComponentType) Component {
	if c, found := en.components[t]; found {
		return c
	}
	return nil
}

// returns state representation where components can be added
func (en *Entity) State(s string) *state {
	if _, ok := en.states[s]; !ok {
		en.states[s] = map[ComponentType]Component{}
	}

	return &state{en, en.states[s]}
}

// return current state of Entity
func (en *Entity) CurrentState() string {
	return en.currentState
}

// Change internal state of Entity, adds/removes registered Components
func (en *Entity) ChangeState(s string) {
	if s == en.currentState {
		return
	}

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
	for _, c := range components {
		s.components[c.Type()] = c
	}
}

// Remove Component from State
func (s *state) Remove(t ComponentType) {
	delete(s.components, t)
}
