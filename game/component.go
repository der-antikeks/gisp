package game

// General purpose object that identifies a set of components.
type Entity int

// ComponentType identifies a specific Component
type ComponentType int

// Component is a set of data needed for a specific purpose
type Component interface {
	Type() ComponentType
}

type Property struct {
	Name  string
	Value interface{}
}

type SerializableComponent interface {
	Component
	Load(<-chan Property) error
	Save(chan<- Property) error
}

// Aspect is a specific set of component-types
type aspect struct {
	types []ComponentType
}

// Slice of ComponentTypes has all components of the aspect
func (a *aspect) accepts(types []ComponentType) bool {
	if a == nil && types != nil {
		return false
	}

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
	if a == nil && types != nil {
		return false
	}

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
