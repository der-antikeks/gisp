package game

import ()

// ComponentType identifies a specific Component
type ComponentType uint

func (t ComponentType) Uint() uint {
	return uint(t)
}

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
func (a *aspect) subset(types []ComponentType) bool {
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

type ComponentCollection interface {
	Set(Component)
	Get(ComponentType) Component
	Remove(ComponentType)
	Length() int
	Iterate(func(Component) bool)
}

// not safe for concurrent use
// fast set/get/remove, slow iterate
type ComponentMap struct {
	data map[ComponentType]Component
}

func (m *ComponentMap) Set(c Component) {
	if m.data == nil {
		m.data = map[ComponentType]Component{}
	}
	m.data[c.Type()] = c
}

func (m *ComponentMap) Get(t ComponentType) Component {
	if c, ok := m.data[t]; ok {
		return c
	}
	return nil
}

func (m *ComponentMap) Remove(t ComponentType) {
	delete(m.data, t)
}

func (m *ComponentMap) Length() int {
	return len(m.data)
}

func (m *ComponentMap) Iterate(f func(Component) bool) {
	for _, c := range m.data {
		if !f(c) {
			return
		}
	}
}

// not safe for concurrent use
// slow set/get/remove, fast iterate
type ComponentSlice struct {
	data   []Component
	length int
}

func (m *ComponentSlice) Set(c Component) {
	if size := int(c.Type()); size >= len(m.data) {
		if m.data == nil {
			m.data = make([]Component, size+1)
		} else {
			n := make([]Component, size+1)
			copy(n, m.data)
			m.data = n
		}
	}

	if m.Get(c.Type()) != nil {
		return
	}

	m.data[c.Type()] = c
	m.length++
}

func (m *ComponentSlice) Get(t ComponentType) Component {
	if int(t) >= len(m.data) {
		return nil
	}

	if c := m.data[t]; c != nil {
		return c
	}
	return nil
}

func (m *ComponentSlice) Remove(t ComponentType) {
	if int(t) >= len(m.data) {
		return
	}

	m.data[t] = nil
	m.length--
}

func (m *ComponentSlice) Length() int {
	return m.length
}

func (m *ComponentSlice) Iterate(f func(Component) bool) {
	for _, c := range m.data {
		if c == nil {
			continue
		}
		if !f(c) {
			return
		}
	}
}
