package ecs

import ()

// Collection is a specific set of components
type Collection struct {
	types    []ComponentType
	entities []*Entity

	added   map[System]func(en *Entity)
	removed map[System]func(en *Entity)
}

func NewCollection(types []ComponentType) *Collection {
	return &Collection{
		types:    types,
		entities: []*Entity{},

		added:   map[System]func(en *Entity){},
		removed: map[System]func(en *Entity){},
	}
}

// Entity has the same components as the Collection
func (c *Collection) accepts(en *Entity) bool {
	for _, t := range c.types {
		if en.Get(t) == nil {
			return false
		}
	}
	return true
}

// Collection equals slice of ComponentTypes
func (c *Collection) equals(b []ComponentType) bool {
	if len(c.types) != len(b) {
		return false
	}

	var found bool
	for _, t := range c.types {
		found = false
		for _, t2 := range b {
			if t == t2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// added/removed functions are called upon the occurrence of the respective action by the Collection
func (c *Collection) Subscribe(s System, added, removed func(en *Entity)) {
	c.added[s] = added
	c.removed[s] = removed
}

// added/removed function  of the passed system are no longer called
func (c *Collection) Unsubscribe(s System) {
	delete(c.added, s)
	delete(c.removed, s)
}

// add Entity to Collection without any checking of Components
func (c *Collection) add(en *Entity) {
	c.entities = append(c.entities, en)

	for _, f := range c.added {
		f(en)
	}
}

// remove Entity from Collection
func (c *Collection) remove(en *Entity) {
	for i, f := range c.entities {
		if f == en {
			copy(c.entities[i:], c.entities[i+1:])
			c.entities[len(c.entities)-1] = nil
			c.entities = c.entities[:len(c.entities)-1]

			for _, f := range c.removed {
				f(en)
			}

			return
		}
	}
}

// Return all registered Entities of the Engine, that matches the Collection
func (c *Collection) Entities() []*Entity {
	//return c.entities // invalid nil pointer after removing entity

	ret := make([]*Entity, len(c.entities))
	copy(ret, c.entities)
	return ret
}

// Returns the first matched Entity
func (c *Collection) First() *Entity {
	if len(c.entities) < 1 {
		return nil
	}
	return c.entities[0]
}

/*
func (c *Collection) Last() *Entity {
	l := len(c.entities)
	if l < 1 {
		return nil
	}
	return c.entities[l-1]
}
*/
