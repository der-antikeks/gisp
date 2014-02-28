package ecs

// Collection is a specific set of components
type Collection struct {
	types  []ComponentType
	engine *Engine
}

// Entity has the same components as the Collection
func (c *Collection) accepts(e *Entity) bool {
	for _, t := range c.types {
		if e.Get(t) == nil {
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

// Return all registered Entities of the Engine, that matches the Collection
func (c *Collection) Entities() []*Entity {
	return c.engine.Collections[c]
}

// Returns the first matched Entity
func (c *Collection) First() *Entity {
	if len(c.engine.Collections[c]) < 1 {
		return nil
	}
	return c.engine.Collections[c][0]
}

/*
func (c *Collection) Last() *Entity {
	l := len(c.engine.Collections[c])
	if l < 1 {
		return nil
	}
	return c.engine.Collections[c][l-1]
}
*/
