package ecs

import ()

// Collection is a specific set of components
type Collection struct {
	types    []ComponentType
	entities []*Entity

	EntityAdded   *EntityObserver
	EntityRemoved *EntityObserver
}

func NewCollection(types []ComponentType) *Collection {
	return &Collection{
		types:         types,
		entities:      []*Entity{},
		EntityAdded:   new(EntityObserver),
		EntityRemoved: new(EntityObserver),
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

	for _, t := range c.types {
		var found bool
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

func (c *Collection) add(en *Entity) {
	c.entities = append(c.entities, en)
	c.EntityAdded.Publish(en)
}

func (c *Collection) remove(en *Entity) {
	for i, f := range c.entities {
		if f == en {
			copy(c.entities[i:], c.entities[i+1:])
			c.entities[len(c.entities)-1] = nil
			c.entities = c.entities[:len(c.entities)-1]

			c.EntityRemoved.Publish(en)

			return
		}
	}
}

// Return all registered Entities of the Engine, that matches the Collection
func (c *Collection) Entities() []*Entity {
	return c.entities
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
