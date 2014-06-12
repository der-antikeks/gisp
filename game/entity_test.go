package game

import (
	"math/rand"
	"testing"
)

const (
	maxv = 1e6
)

func TestEntitySlice(t *testing.T) {
	testEntityCollection(&EntitySlice{}, t)
}

func TestEntityMap(t *testing.T) {
	testEntityCollection(&EntityMap{}, t)
}

func TestEntityBitset(t *testing.T) {
	testEntityCollection(&EntityBitset{}, t)
}

func testEntityCollection(c EntityCollection, t *testing.T) {
	if r := c.Length(); r != 0 {
		t.Errorf("empty collection has wrong length of %v", r)
	}

	c.Set(Entity(42))
	if r := c.Length(); r != 1 {
		t.Errorf("added one entity but length is %v instead of 1", r)
	}

	if r := c.Get(Entity(815)); r != false {
		t.Errorf("getting unknown entity returns %v instead of false", r)
	}

	if r := c.Get(Entity(42)); r != true {
		t.Errorf("getting known entity returns %v instead of true", r)
	}

	c.Set(Entity(815))
	if r := c.Length(); r != 2 {
		t.Errorf("added another entity but length is %v instead of 2", r)
	}

	var s uint
	c.Iterate(func(e Entity) bool {
		s += e.Uint()
		return true
	})
	if s != 815+42 {
		t.Errorf("complete iteration failed")
	}

	c.Iterate(func(e Entity) bool {
		s = e.Uint()
		return false
	})
	if s != 815 && s != 42 {
		t.Errorf("partial iteration failed")
	}

	c.Remove(Entity(42))
	if r := c.Length(); r != 1 {
		t.Errorf("removed one entity but length is %v instead of 1", r)
	}

	if r := c.Get(Entity(815)); r != true {
		t.Errorf("remaining entity is %v instead of true", r)
	}

	if r := c.Get(Entity(42)); r != false {
		t.Errorf("removed entity is %v instead of false", r)
	}

	c.Set(Entity(815))
	if r := c.Length(); r != 1 {
		t.Errorf("added duplicate entity, length is %v instead of 1", r)
	}
}

func BenchmarkEntitySlice_Set(b *testing.B) {
	benchmarkEntityCollection_Set(&EntitySlice{}, b)
}

func BenchmarkEntityMap_Set(b *testing.B) {
	benchmarkEntityCollection_Set(&EntityMap{}, b)
}

func BenchmarkEntityBitset_Set(b *testing.B) {
	benchmarkEntityCollection_Set(&EntityBitset{}, b)
}

func benchmarkEntityCollection_Set(c EntityCollection, b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(Entity(rand.Uint32() % maxv))
	}
}

func BenchmarkEntitySlice_Get(b *testing.B) {
	benchmarkEntityCollection_Get(&EntitySlice{}, b)
}

func BenchmarkEntityMap_Get(b *testing.B) {
	benchmarkEntityCollection_Get(&EntityMap{}, b)
}

func BenchmarkEntityBitset_Get(b *testing.B) {
	benchmarkEntityCollection_Get(&EntityBitset{}, b)
}

func benchmarkEntityCollection_Get(c EntityCollection, b *testing.B) {
	for i := 0; i < b.N; i++ {
		c.Set(Entity(rand.Uint32() % maxv))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(Entity(rand.Uint32() % maxv))
	}
}

func BenchmarkEntitySlice_Remove(b *testing.B) {
	benchmarkEntityCollection_Remove(&EntitySlice{}, b)
}

func BenchmarkEntityMap_Remove(b *testing.B) {
	benchmarkEntityCollection_Remove(&EntityMap{}, b)
}

func BenchmarkEntityBitset_Remove(b *testing.B) {
	benchmarkEntityCollection_Remove(&EntityBitset{}, b)
}

func benchmarkEntityCollection_Remove(c EntityCollection, b *testing.B) {
	for i := 0; i < b.N; i++ {
		c.Set(Entity(rand.Uint32() % maxv))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Remove(Entity(rand.Uint32() % maxv))
	}
}

func BenchmarkEntitySlice_Iterate(b *testing.B) {
	benchmarkEntityCollection_Iterate(&EntitySlice{}, b)
}

func BenchmarkEntityMap_Iterate(b *testing.B) {
	benchmarkEntityCollection_Iterate(&EntityMap{}, b)
}

func BenchmarkEntityBitset_Iterate(b *testing.B) {
	benchmarkEntityCollection_Iterate(&EntityBitset{}, b)
}

func benchmarkEntityCollection_Iterate(c EntityCollection, b *testing.B) {
	for i := 0; i < b.N; i++ {
		c.Set(Entity(rand.Uint32() % maxv))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sum uint
		c.Iterate(func(e Entity) bool {
			sum += e.Uint()
			return true
		})
	}
}
