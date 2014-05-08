package ecs

import ()

// Aspect is a specific set of component-types
type aspect struct {
	types []ComponentType
}

// Entity has the same components as the aspect
func (a *aspect) accepts(en *entity) bool {
	for _, t := range a.types {
		if en.Get(t) == nil {
			return false
		}
	}
	return true
}

// Aspect equals a slice of ComponentTypes
func (a *aspect) equals(b []ComponentType) bool {
	if len(a.types) != len(b) {
		return false
	}

	var found bool
	for _, t := range a.types {
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
