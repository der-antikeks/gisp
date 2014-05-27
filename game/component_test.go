package game

import (
	"math/rand"
	"testing"
)

const (
	TestComponentAType ComponentType = 1 << iota
	TestComponentBType
	TestComponentCType
)

type TestComponentA struct{ Data string }

func (c TestComponentA) Type() ComponentType {
	return TestComponentAType
}

type TestComponentB struct{ Data int }

func (c TestComponentB) Type() ComponentType {
	return TestComponentBType
}

type TestComponentC struct{ Data float64 }

func (c TestComponentC) Type() ComponentType {
	return TestComponentCType
}

func TestComponentMap(t *testing.T) {
	testComponentCollection(&ComponentMap{}, t)
}

func TestComponentSlice(t *testing.T) {
	testComponentCollection(&ComponentSlice{}, t)
}

func testComponentCollection(c ComponentCollection, t *testing.T) {
	if r := c.Length(); r != 0 {
		t.Errorf("empty collection has wrong length of %v", r)
	}

	c.Set(TestComponentA{})
	if r := c.Length(); r != 1 {
		t.Errorf("added one component but length is %v instead of 1", r)
	}

	if r := c.Get(TestComponentBType); r != nil {
		t.Errorf("getting unknown component type returns %v instead of nil", r)
	}

	if _, ok := c.Get(TestComponentAType).(TestComponentA); !ok {
		t.Errorf("getting known component returns %T instead of TestComponentA", c.Get(TestComponentAType))
	}

	c.Set(TestComponentB{})
	if r := c.Length(); r != 2 {
		t.Errorf("added another component but length is %v instead of 2", r)
	}

	var cnt int
	c.Iterate(func(Component) bool {
		cnt++
		return true
	})
	if cnt != 2 {
		t.Errorf("complete iteration failed")
	}

	cnt = 0
	c.Iterate(func(Component) bool {
		cnt++
		return false
	})
	if cnt != 1 {
		t.Errorf("partial iteration failed")
	}

	c.Remove(TestComponentAType)
	if r := c.Length(); r != 1 {
		t.Errorf("removed one component but length is %v instead of 1", r)
	}

	if _, ok := c.Get(TestComponentBType).(TestComponentB); !ok {
		t.Errorf("remaining component is %T instead of TestComponentB", c.Get(TestComponentBType))
	}

	if r := c.Get(TestComponentAType); r != nil {
		t.Errorf("removed component is %v instead of nil", r)
	}

	c.Set(TestComponentB{})
	if r := c.Length(); r != 1 {
		t.Errorf("added duplicate component, length is %v instead of 1", r)
	}
}

func BenchmarkComponentMap_Set(b *testing.B) {
	benchmarkComponentCollection_Set(&ComponentMap{}, b)
}

func BenchmarkComponentSlice_Set(b *testing.B) {
	benchmarkComponentCollection_Set(&ComponentSlice{}, b)
}

func benchmarkComponentCollection_Set(c ComponentCollection, b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		switch rand.Intn(3) {
		case 0:
			c.Set(TestComponentA{})
		case 1:
			c.Set(TestComponentB{})
		case 2:
			c.Set(TestComponentC{})
		}
	}
}

func BenchmarkComponentMap_Get(b *testing.B) {
	benchmarkComponentCollection_Get(&ComponentMap{}, b)
}

func BenchmarkComponentSlice_Get(b *testing.B) {
	benchmarkComponentCollection_Get(&ComponentSlice{}, b)
}

func benchmarkComponentCollection_Get(c ComponentCollection, b *testing.B) {
	for i := 0; i < b.N; i++ {
		switch rand.Intn(3) {
		case 0:
			c.Set(TestComponentA{})
		case 1:
			c.Set(TestComponentB{})
		case 2:
			c.Set(TestComponentC{})
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		switch rand.Intn(3) {
		case 0:
			c.Get(TestComponentAType)
		case 1:
			c.Get(TestComponentBType)
		case 2:
			c.Get(TestComponentCType)
		}
	}
}

func BenchmarkComponentMap_Remove(b *testing.B) {
	benchmarkComponentCollection_Remove(&ComponentMap{}, b)
}

func BenchmarkComponentSlice_Remove(b *testing.B) {
	benchmarkComponentCollection_Remove(&ComponentSlice{}, b)
}

func benchmarkComponentCollection_Remove(c ComponentCollection, b *testing.B) {
	for i := 0; i < b.N; i++ {
		switch rand.Intn(3) {
		case 0:
			c.Set(TestComponentA{})
		case 1:
			c.Set(TestComponentB{})
		case 2:
			c.Set(TestComponentC{})
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		switch rand.Intn(3) {
		case 0:
			c.Remove(TestComponentAType)
		case 1:
			c.Remove(TestComponentBType)
		case 2:
			c.Remove(TestComponentCType)
		}
	}
}

func BenchmarkComponentMap_Iterate(b *testing.B) {
	benchmarkComponentCollection_Iterate(&ComponentMap{}, b)
}

func BenchmarkComponentSlice_Iterate(b *testing.B) {
	benchmarkComponentCollection_Iterate(&ComponentSlice{}, b)
}

func benchmarkComponentCollection_Iterate(c ComponentCollection, b *testing.B) {
	for i := 0; i < b.N; i++ {
		switch rand.Intn(3) {
		case 0:
			c.Set(TestComponentA{})
		case 1:
			c.Set(TestComponentB{})
		case 2:
			c.Set(TestComponentC{})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cnt int
		c.Iterate(func(Component) bool {
			cnt++
			return true
		})
	}
}
