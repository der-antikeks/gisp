package ecs

import (
	"fmt"
	"sort"
	"time"
)

// Engine collects and connects Systems with matching Entities
type Engine struct {
	systems          []System
	systemPriorities []SystemPriority
	updatePriority   bool

	entities    map[*Entity][]*Collection
	collections []*Collection

	deleted []*Entity
}

// Creates a new Engine
func NewEngine() *Engine {
	return &Engine{
		systems:          []System{},
		systemPriorities: []SystemPriority{},

		entities:    map[*Entity][]*Collection{},
		collections: []*Collection{},

		deleted: []*Entity{},
	}
}

// Add Entity to Engine and all registered Systems. The Engine is notified of later Component-changes.
func (e *Engine) AddEntity(en *Entity) error {
	if _, found := e.entities[en]; found {
		return fmt.Errorf("entity '%v' already registered", en.Name)
	}

	// set engine of entity
	en.engine = e

	// add entity to entities map
	e.entities[en] = []*Collection{}

	// add entity to matching collections slice
	for _, c := range e.collections {
		if c.accepts(en) {
			c.add(en)
			e.entities[en] = append(e.entities[en], c)
		}
	}

	return nil
}

// Remove Entity from Engine and all registered collections
func (e *Engine) RemoveEntity(en *Entity) {
	e.deleted = append(e.deleted, en)
}

// remove all as deleted marked entities
func (e *Engine) removeDeletedEntities() {
	if len(e.deleted) == 0 {
		return
	}

	for i, en := range e.deleted {
		e.removeEntity(en)
		e.deleted[i] = nil
	}

	e.deleted = e.deleted[:0]
}

// Remove Entity from from Engine and collections
func (e *Engine) removeEntity(en *Entity) {
	if _, found := e.entities[en]; !found {
		return
	}

	// search in all collections of entity
	for _, c := range e.entities[en] {
		c.remove(en)
	}

	// remove entity from entities map
	delete(e.entities, en)

	// remove engine from entity
	en.engine = nil
}

// Called by the Entity whose components are removed after adding it to the Engine
func (e *Engine) entityRemovedComponent(en *Entity, _ Component) {
	for i, c := range e.entities[en] {
		if !c.accepts(en) {
			// component does not accept entity anymore
			// remove component from entites slice
			copy(e.entities[en][i:], e.entities[en][i+1:])
			e.entities[en][len(e.entities[en])-1] = nil
			e.entities[en] = e.entities[en][:len(e.entities[en])-1]

			c.remove(en)
		}
	}
}

// Called by the Entity, if components are added after adding it to the Engine
func (e *Engine) entityAddedComponent(en *Entity, _ Component) {
	// add entity to matching collections slice
	for _, c := range e.collections {
		if c.accepts(en) {
			c.add(en)
			e.entities[en] = append(e.entities[en], c)
		}
	}
}

// Add System to Engine. Already registered Entites are added to the System
func (e *Engine) AddSystem(s System, p SystemPriority) error {
	if err := s.AddedToEngine(e); err != nil {
		return err
	}

	e.systems = append(e.systems, s)
	e.systemPriorities = append(e.systemPriorities, p)
	e.updatePriority = true
	return nil
}

// Remove System from Engine
func (e *Engine) RemoveSystem(s System) {
	if err := s.RemovedFromEngine(e); err == nil {
		for i, f := range e.systems {
			if f == s {
				// found, remove system from slice
				copy(e.systems[i:], e.systems[i+1:])
				e.systems[len(e.systems)-1] = nil
				e.systems = e.systems[:len(e.systems)-1]

				// remove priority
				e.systemPriorities = e.systemPriorities[:i+copy(e.systemPriorities[i:], e.systemPriorities[i+1:])]

				//e.updatePriority = true
				return
			}
		}
	}
}

// Get Collection of Components. Creates new Collection if necessary
func (e *Engine) Collection(types ...ComponentType) *Collection {
	// old Collection
	for _, c := range e.collections {
		if c.equals(types) {
			return c
		}
	}

	// new Collection
	c := NewCollection(types)
	e.collections = append(e.collections, c)

	// add matching entities to Collection slice
	for en := range e.entities {
		if c.accepts(en) {
			c.add(en)
			e.entities[en] = append(e.entities[en], c)
		}
	}

	return c
}

// byPriority attaches the methods of sort.Interface to []System, sorting in increasing order of the System.Priority() method.
type byPriority struct {
	systems    []System
	priorities []SystemPriority
}

func (a byPriority) Len() int { return len(a.systems) }
func (a byPriority) Swap(i, j int) {
	a.systems[i], a.systems[j] = a.systems[j], a.systems[i]
	a.priorities[i], a.priorities[j] = a.priorities[j], a.priorities[i]
}
func (a byPriority) Less(i, j int) bool {
	return a.priorities[i] < a.priorities[j]
}

func (e *Engine) sortSystems() {
	sort.Sort(byPriority{
		systems:    e.systems,
		priorities: e.systemPriorities,
	})
	e.updatePriority = false
}

// Update each Systems in order of priority
func (e *Engine) Update(delta time.Duration) error {
	if e.updatePriority {
		e.sortSystems()
	}

	for _, s := range e.systems {
		e.removeDeletedEntities()
		if err := s.Update(delta); err != nil {
			return err
		}
	}

	return nil
}
