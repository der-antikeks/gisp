package ecs

import (
	"fmt"
	"reflect"
	"sync"
)

type Set struct {
	sync.Mutex
	data map[reflect.Type]interface{}
}

func NewSet(components ...interface{}) Set {
	s := Set{
		data: make(map[reflect.Type]interface{}),
	}

	for _, component := range components {
		s.Add(component)
	}

	return s
}

func (s Set) String() string {
	r := "Set{"
	for t, d := range s.data {
		r += fmt.Sprintf("%s: %v, ", t, d)
	}
	return r + "}"
}

func (s Set) ctype(component interface{}) reflect.Type {
	var t reflect.Type
	if c, ok := component.(reflect.Type); ok {
		t = c
	} else {
		t = reflect.TypeOf(component)
		//t = reflect.ValueOf(component).Elem().Type()
	}

	return t
}

func (s Set) Add(component interface{}) {
	s.Lock()
	defer s.Unlock()

	t := s.ctype(component)
	if _, ok := s.data[t]; !ok {
		s.data[t] = component
	}
}

func (s Set) Remove(component interface{}) {
	s.Lock()
	defer s.Unlock()

	delete(s.data, s.ctype(component))
}

func (s Set) Reset() {
	s.Lock()
	defer s.Unlock()

	for t := range s.data {
		delete(s.data, t)
	}
}

func (s Set) Get(component interface{}) interface{} {
	s.Lock()
	defer s.Unlock()

	if r, ok := s.data[s.ctype(component)]; ok && r != nil {
		return r
	}
	return nil
}

/*
// If A and B are sets and every element of A is also an element of B
func (a Set) SubsetOf(b Set) (Set, bool) {
	a.Lock()
	defer a.Unlock()

	b.Lock()
	defer b.Unlock()

	subset := NewSet()

	for t := range a.data {
		if b.data[t] == nil {
			return Set{}, false
		}

		subset.data[t] = b.data[t]
	}

	return subset, true
}
*/
