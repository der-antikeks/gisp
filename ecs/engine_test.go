package ecs

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type testSortSystem struct {
	priority SystemPriority
	update   func() error
	added    func() error
	removed  func() error
}

func (s *testSortSystem) Priority() SystemPriority {
	return s.priority
}
func (s *testSortSystem) AddedToEngine(*Engine) error {
	if s.added != nil {
		return s.added()
	}
	return fmt.Errorf("Test system was not properly initialized")
}
func (s *testSortSystem) RemovedFromEngine(*Engine) error {
	if s.removed != nil {
		return s.removed()
	}
	return fmt.Errorf("Test system was not properly initialized")
}
func (s *testSortSystem) Update(time.Duration) error {
	if s.update != nil {
		return s.update()
	}
	return fmt.Errorf("Test system was not properly initialized")
}

func TestEngineSortSystems(t *testing.T) {
	newSystem := func(p SystemPriority, o chan SystemPriority) System {
		return &testSortSystem{
			priority: p,
			update: func() error {
				o <- p
				return nil
			},
			added:   func() error { return nil },
			removed: func() error { return nil },
		}
	}

	engine := NewEngine()
	out := make(chan SystemPriority)
	n := 10

	for i := 0; i < n; i++ {
		p := SystemPriority(rand.Intn(100))
		engine.AddSystem(newSystem(p, out), p)
	}

	go engine.Update(0)

	var prev SystemPriority
	for i := 0; i < n; i++ {
		p := <-out
		if p < prev {
			t.Errorf("unsorted: %v < %v", p, prev)
		}
		prev = p
	}
}
