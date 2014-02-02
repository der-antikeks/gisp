package ecs

import (
	"fmt"
	"reflect"
	"sync"
)

// set of components
type set struct {
	sync.Mutex
	data map[reflect.Type]interface{}
}

func newset(components ...interface{}) set {
	s := set{
		data: make(map[reflect.Type]interface{}),
	}

	for _, component := range components {
		s.Add(component)
	}

	return s
}

func (s set) String() string {
	r := "set{"
	for t, d := range s.data {
		r += fmt.Sprintf("%s: %v, ", t, d)
	}
	return r + "}"
}

func (s set) ctype(component interface{}) reflect.Type {
	var t reflect.Type
	if c, ok := component.(reflect.Type); ok {
		t = c
	} else {
		t = reflect.TypeOf(component)
		//t = reflect.ValueOf(component).Elem().Type()
	}

	return t
}

func (s set) Add(component interface{}) {
	s.Lock()
	defer s.Unlock()

	t := s.ctype(component)
	if _, ok := s.data[t]; !ok {
		s.data[t] = component
	}
}

func (s set) Remove(component interface{}) {
	s.Lock()
	defer s.Unlock()

	delete(s.data, s.ctype(component))
}

func (s set) Reset() {
	s.Lock()
	defer s.Unlock()

	for t := range s.data {
		delete(s.data, t)
	}
}

func (s set) Get(component interface{}) interface{} {
	s.Lock()
	defer s.Unlock()

	if r, ok := s.data[s.ctype(component)]; ok && r != nil {
		return r
	}
	return nil
}

/*
// If A and B are sets and every element of A is also an element of B
func (a set) SubsetOf(b set) (set, bool) {
	a.Lock()
	defer a.Unlock()

	b.Lock()
	defer b.Unlock()

	subset := Newset()

	for t := range a.data {
		if b.data[t] == nil {
			return set{}, false
		}

		subset.data[t] = b.data[t]
	}

	return subset, true
}
*/
