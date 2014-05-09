package ecs

import ()

// Aspect is a specific set of component-types
type aspect struct {
	types []ComponentType
}

// Slice of ComponentTypes has all components of the aspect
func (a *aspect) accepts(types []ComponentType) bool {
	if len(a.types) > len(types) {
		return false
	}

	var found bool
	for _, t := range a.types {
		found = false
		for _, t2 := range types {
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

// Aspect equals a slice of ComponentTypes
func (a *aspect) equals(types []ComponentType) bool {
	if len(a.types) != len(types) {
		return false
	}

	var found bool
	for _, t := range a.types {
		found = false
		for _, t2 := range types {
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
