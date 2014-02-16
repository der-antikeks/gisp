package ecs

import (
	"fmt"
	"sort"
	"time"
)

// Engine collects and connects Systems with matching Entities
type Engine struct {
	entities       []*Entity
	systems        []*System // prioritised list
	updatePriority bool
}

// Creates a new Engine
func NewEngine() *Engine {
	return &Engine{
		entities: []*Entity{},
		systems:  []*System{},
	}
}

// Add System to Engine. Already registered Entites are added to the System
func (e *Engine) AddSystem(system *System, priority int) error {
	e.systems = append(e.systems, system)
	system.setEngine(e)
	if err := system.init(); err != nil {
		return err
	}

	// priority
	system.Priority = priority
	e.updatePriority = true

	// late system adding
	for _, entity := range e.entities {
		//fmt.Printf("adding entity %s at %T\n", entity.Name, system)
		system.add(entity)
	}

	return nil
}

// Remove System from Engine
func (e *Engine) RemoveSystem(system *System) {
	system.cleanup()
	system.setEngine(nil)

	for i, f := range e.systems {
		if f == system {
			// found, remove from slice
			copy(e.systems[i:], e.systems[i+1:])
			e.systems[len(e.systems)-1] = nil
			e.systems = e.systems[:len(e.systems)-1]

			// sort remaining systems
			e.updatePriority = true
			return
		}
	}
}

// byPriority attaches the methods of sort.Interface to []*System, sorting in increasing order of the System.Priority Field.
type byPriority []*System

func (a byPriority) Len() int           { return len(a) }
func (a byPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPriority) Less(i, j int) bool { return a[i].Priority < a[j].Priority }

// Update each Systems in order of priority field
func (e *Engine) Update(time time.Duration) error {
	if e.updatePriority {
		sort.Sort(byPriority(e.systems))
		e.updatePriority = false
	}

	for _, s := range e.systems {
		//fmt.Printf("updating system %s\n", s.Name)
		if err := s.update(time); err != nil {
			return fmt.Errorf("Error in System %s: %s", s.Name, err)
		}
	}

	return nil
}

// Add Entity to Engine and all registered Systems. The Engine is noticed of later changes of the Entity.
func (e *Engine) AddEntity(entity *Entity) {
	e.entities = append(e.entities, entity)
	entity.setEngine(e)

	for _, s := range e.systems {
		s.add(entity)
	}
}

// Remove Entity from Engine and all registered Systems
func (e *Engine) RemoveEntity(entity *Entity) {
	for i, f := range e.entities {
		if f == entity {
			copy(e.entities[i:], e.entities[i+1:])
			e.entities[len(e.entities)-1] = nil
			e.entities = e.entities[:len(e.entities)-1]

			entity.setEngine(nil)

			for _, s := range e.systems {
				s.remove(entity)
			}

			return
		}
	}
}

// TODO: update each system with new components

// Called by Entity whose components have changed after Adding
func (e *Engine) entityUpdated(entity *Entity) {}

//func (e *Engine) EntityComponentAdded(entity *Entity, component interface{})   {}
//func (e *Engine) EntityComponentRemoved(entity *Entity, component interface{}) {}
