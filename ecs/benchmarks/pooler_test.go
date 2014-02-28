package benchmarks

import (
	"math/rand"
	"sync"
	"testing"
)

// INTERFACES

type ComponentType int

type pComponent interface {
	Type() ComponentType
}

type Entity interface {
	Set([]pComponent)
	Get(ComponentType) pComponent
}

type Engine interface {
	NewEntity([]pComponent) Entity
	RemoveEntity(Entity)
	GetEntitiesWithComponents([]ComponentType) []Entity
}

// BENCHMARK

func worker(n int, e Engine, c []pComponent, t []ComponentType, wg *sync.WaitGroup) {
	en := e.NewEntity(c)
	for i := 0; i < n; i++ {
		switch rand.Intn(3) {
		case 0: // add entity
			en = e.NewEntity(c)
		case 1: // remove entity
			e.RemoveEntity(en)
		case 2: // get entities
			e.GetEntitiesWithComponents(t)
		}
	}
	wg.Done()
}

func runBenchmark(n int, e Engine, c []pComponent, t []ComponentType) {
	rand.Seed(12345)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go worker(n/100, e, c, t, &wg)
	}
	wg.Wait()
}

// POINTER IMPLEMENTATION

const (
	_                            = iota
	ComponentTypeA ComponentType = iota
	ComponentTypeB
)

type ComponentA struct {
	Value float64
}

func (c ComponentA) Type() ComponentType {
	return ComponentTypeA
}

type ComponentB struct {
	Value string
}

func (c ComponentB) Type() ComponentType {
	return ComponentTypeB
}

type PointerEntity struct {
	lock       sync.RWMutex
	components map[ComponentType]pComponent
}

func (e *PointerEntity) Set(cs []pComponent) {
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, c := range cs {
		e.components[c.Type()] = c
	}
}

func (e *PointerEntity) Get(t ComponentType) pComponent {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.components[t]
}

type PointerEngine struct {
	lock       sync.RWMutex
	entities   []*PointerEntity
	components map[ComponentType][]*PointerEntity
}

func (e *PointerEngine) NewEntity(c []pComponent) Entity {
	en := &PointerEntity{
		components: map[ComponentType]pComponent{},
	}

	e.entities = append(e.entities, en)
	return en
}

func (e *PointerEngine) RemoveEntity(en Entity) {

}

func (e *PointerEngine) GetEntitiesWithComponents(t []ComponentType) []Entity {
	return nil
}

func BenchmarkPointer(b *testing.B) {
	e := &PointerEngine{
		entities:   []*PointerEntity{},
		components: map[ComponentType][]*PointerEntity{},
	}

	c := []pComponent{ComponentA{1.23}, ComponentB{"b"}}
	t := []ComponentType{ComponentTypeA, ComponentTypeB}

	b.ResetTimer()
	runBenchmark(b.N, e, c, t)
}
