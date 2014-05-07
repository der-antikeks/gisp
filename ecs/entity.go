package ecs

import (
	"sync"
)

// General purpose object that consists of a name and a set of components.
type Entity struct {
	sync.RWMutex

	Name   string // TODO: only for debugging
	engine *Engine

	// components that define the entity's current state
	components map[ComponentType]Component
}

// Add new Components to the Entity or update existing
func (en *Entity) Set(components ...Component) {
	en.Lock()
	var updated bool
	for _, c := range components {
		if !updated {
			if _, found := en.components[c.Type()]; found {
				updated = true
			}
		}

		en.components[c.Type()] = c
	}
	en.Unlock()

	if en.engine != nil {
		if updated {
			en.engine.entityUpdatedComponent(en)
		} else {
			en.engine.entityAddedComponent(en)
		}
	}
}

// Detach Components from Entity
func (en *Entity) Remove(types ...ComponentType) {
	en.Lock()
	for _, t := range types {
		if _, found := en.components[t]; !found {
			return
		}

		delete(en.components, t)
	}
	en.Unlock()

	if en.engine != nil {
		en.engine.entityRemovedComponent(en)
	}
}

// Get specific Component of Entity
func (en *Entity) Get(t ComponentType) Component {
	en.RLock()
	defer en.RUnlock()

	if c, found := en.components[t]; found {
		return c
	}
	return nil
}

type EntityList interface {
	//Add(*Entity)
	//Remove(*Entity)
	Entities() []*Entity
	First() *Entity
}

type SliceEntityList []*Entity

func (l *SliceEntityList) Add(e *Entity) {
	*l = append(*l, e)
}
func (l *SliceEntityList) Remove(e *Entity) {
	a := *l
	for i, f := range a {
		if f == e {
			copy(a[i:], a[i+1:])
			a[len(a)-1] = nil
			*l = a[:len(a)-1]
			return
		}
	}
}
func (l SliceEntityList) Entities() []*Entity {
	return l
}
func (l SliceEntityList) First() *Entity {
	if len(l) < 1 {
		return nil
	}
	return l[0]
}
