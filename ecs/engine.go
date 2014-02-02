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
func (e *Engine) AddSystem(system *System, priority int) {
	e.systems = append(e.systems, system)
	system.SetEngine(e)
	system.Init()

	// priority
	system.Priority = priority
	e.updatePriority = true

	// late system adding
	for _, entity := range e.entities {
		//fmt.Printf("adding entity %s at %T\n", entity.Name, system)
		system.Add(entity)
	}
}

// Remove System from Engine. Calls Cleaup-Function of System.
func (e *Engine) RemoveSystem(system *System) {
	system.Cleanup()
	system.SetEngine(nil)

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

// ByPriority attaches the methods of sort.Interface to []*System, sorting in increasing order of the System.Priority Field.
type ByPriority []*System

func (a ByPriority) Len() int           { return len(a) }
func (a ByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool { return a[i].Priority < a[j].Priority }

// Update each Systems in order of priority field
func (e *Engine) Update(time time.Duration) error {
	if e.updatePriority {
		sort.Sort(ByPriority(e.systems))
		e.updatePriority = false
	}

	for _, s := range e.systems {
		//fmt.Printf("updating system %s\n", s.Name)
		if err := s.Update(time); err != nil {
			return fmt.Errorf("Error in System %s: %s", s.Name, err)
		}
	}

	return nil
}

// Add Entity to Engine and all registered Systems.
// Calls SetEngine of Entity to get noticed of later changes of the components.
func (e *Engine) AddEntity(entity *Entity) {
	e.entities = append(e.entities, entity)
	entity.SetEngine(e)

	for _, s := range e.systems {
		s.Add(entity)
	}
}

// Remove Entity from Engine and all registered Systems
func (e *Engine) RemoveEntity(entity *Entity) {
	for i, f := range e.entities {
		if f == entity {
			copy(e.entities[i:], e.entities[i+1:])
			e.entities[len(e.entities)-1] = nil
			e.entities = e.entities[:len(e.entities)-1]

			entity.SetEngine(nil)

			for _, s := range e.systems {
				s.Remove(entity)
			}

			return
		}
	}
}

// TODO: update each system with new components

// Called by Entity whose components have changed after Adding
func (e *Engine) EntityUpdated(entity *Entity) {}

//func (e *Engine) EntityComponentAdded(entity *Entity, component interface{})   {}
//func (e *Engine) EntityComponentRemoved(entity *Entity, component interface{}) {}
