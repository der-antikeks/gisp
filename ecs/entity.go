package ecs

import (
	"sync"
)

type Entity int

// General purpose object that consists of a name and a set of components.
type entity struct {
	sync.RWMutex

	Name   string // TODO: only for debugging
	engine *Engine

	// components that define the entity's current state
	components map[ComponentType]Component
}

// Add new Components to the Entity or update existing
func (en *entity) Set(components ...Component) (updated bool) {
	en.Lock()
	defer en.Unlock()

	for _, c := range components {
		if c == nil { // TODO: should not happen but does...
			continue
		}

		if !updated {
			if _, found := en.components[c.Type()]; found {
				updated = true
			}
		}

		en.components[c.Type()] = c
	}
	return
}

// Detach Components from Entity
func (en *entity) Remove(types ...ComponentType) {
	en.Lock()
	defer en.Unlock()
	for _, t := range types {
		if _, found := en.components[t]; !found {
			return
		}

		delete(en.components, t)
	}
}

// Get specific Component of Entity
func (en *entity) Get(t ComponentType) Component {
	en.RLock()
	defer en.RUnlock()

	if c, found := en.components[t]; found && c != nil {
		return c
	}
	return nil
}

type EntityList interface {
	//Add(*Entity)
	//Remove(*Entity)
	Entities() []Entity
	First() Entity
}

type SliceEntityList []Entity

func (l *SliceEntityList) Add(e Entity) {
	*l = append(*l, e)
}
func (l *SliceEntityList) Remove(e Entity) {
	a := *l
	for i, f := range a {
		if f == e {
			*l = append(a[:i], a[i+1:]...)
			return
		}
	}
}
func (l SliceEntityList) Entities() []Entity {
	return l
}
func (l SliceEntityList) First() (Entity, bool) {
	if len(l) < 1 {
		return 0, false
	}
	return l[0], true
}
