package benchmarks

import (
	"math/rand"
	"testing"
)

type Set interface {
	Add(interface{})
	Has(interface{}) bool
	Remove(interface{})

	Iterate() <-chan interface{}

	Intersect(Set) Set
}

func testSet(s Set, t *testing.T) {
	s.Add(1)
	s.Add(3)

	if !s.Has(1) {
		t.Error("add failed")
	}

	if s.Has(2) {
		t.Error("has failed")
	}

	s.Remove(3)
	if s.Has(3) {
		t.Error("remove failed")
	}

	for r := range s.Iterate() {
		if r.(int) != 1 {
			t.Error("interate failed")
		}
	}

	/* TODO
	if s.Intersect(s) != s {
		t.Error("intersect failed")
	}
	*/
}

func benchSetAdd(s Set, b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add(rand.Intn(1000))
	}
}

func benchSetHas(s Set, b *testing.B) {
	for i := 0; i <= 1000; i++ {
		s.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Has(rand.Intn(1000))
	}
}

func benchSetRemove(s Set, b *testing.B) {
	for i := 0; i <= 1000; i++ {
		s.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Remove(rand.Intn(1000))
	}
}

func benchSetIterate(s Set, b *testing.B) {
	for i := 0; i <= 1000; i++ {
		s.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _ = range s.Iterate() {
		}
	}
}

func benchSetIntersect(s Set, o Set, b *testing.B) {
	for i := 0; i <= 1000; i++ {
		s.Add(i)
		o.Add(rand.Intn(1000))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Intersect(o)
	}
}

type MapSet struct {
	data map[interface{}]struct{}
}

func NewMapSet() *MapSet {
	return &MapSet{data: map[interface{}]struct{}{}}
}

func (m *MapSet) Add(i interface{}) {
	m.data[i] = struct{}{}
}

func (m *MapSet) Has(i interface{}) bool {
	_, found := m.data[i]
	return found
}

func (m *MapSet) Remove(i interface{}) {
	delete(m.data, i)
}

func (m *MapSet) Iterate() <-chan interface{} {
	c := make(chan interface{})
	go func() {
		for i := range m.data {
			c <- i
		}
		close(c)
	}()

	return c
}

func (m *MapSet) Intersect(o Set) Set {
	return m
}

func TestMapSet(t *testing.T) {
	testSet(NewMapSet(), t)
}

func BenchmarkMapSet_Add(b *testing.B) {
	rand.Seed(42)
	benchSetAdd(NewMapSet(), b)
}

func BenchmarkMapSet_Has(b *testing.B) {
	rand.Seed(42)
	benchSetHas(NewMapSet(), b)
}

func BenchmarkMapSet_Remove(b *testing.B) {
	rand.Seed(42)
	benchSetRemove(NewMapSet(), b)
}

func BenchmarkMapSet_Iterate(b *testing.B) {
	rand.Seed(42)
	benchSetIterate(NewMapSet(), b)
}

func BenchmarkMapSet_Intersect(b *testing.B) {
	rand.Seed(42)
	benchSetIntersect(NewMapSet(), NewMapSet(), b)
}

type SliceSet struct {
	data []interface{}
}

func NewSliceSet() *SliceSet {
	return &SliceSet{data: []interface{}{}}
}

func (s *SliceSet) Add(i interface{}) {
	s.data = append(s.data, i)
}

func (s *SliceSet) Has(i interface{}) bool {
	for _, v := range s.data {
		if i == v {
			return true
		}
	}

	return false
}

func (s *SliceSet) Remove(i interface{}) {
	for j, v := range s.data {
		if i == v {
			copy(s.data[j:], s.data[j+1:])
			s.data[len(s.data)-1] = nil
			s.data = s.data[:len(s.data)-1]
			return
		}
	}
}

func (s *SliceSet) Iterate() <-chan interface{} {
	c := make(chan interface{})
	go func() {
		for _, i := range s.data {
			c <- i
		}
		close(c)
	}()

	return c
}

func (s *SliceSet) Intersect(o Set) Set { return nil }

func TestSliceSet(t *testing.T) {
	testSet(NewSliceSet(), t)
}

func BenchmarkSliceSet_Add(b *testing.B) {
	rand.Seed(42)
	benchSetAdd(NewSliceSet(), b)
}

func BenchmarkSliceSet_Has(b *testing.B) {
	rand.Seed(42)
	benchSetHas(NewSliceSet(), b)
}

func BenchmarkSliceSet_Remove(b *testing.B) {
	rand.Seed(42)
	benchSetRemove(NewSliceSet(), b)
}

func BenchmarkSliceSet_Iterate(b *testing.B) {
	rand.Seed(42)
	benchSetIterate(NewSliceSet(), b)
}

func BenchmarkSliceSet_Intersect(b *testing.B) {
	rand.Seed(42)
	benchSetIntersect(NewSliceSet(), NewSliceSet(), b)
}
