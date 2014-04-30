package ecs

import (
	"fmt"
	"testing"
	"time"
)

const (
	TestAType ComponentType = iota
	TestBType
	TestCType
)

type TestAComponent struct{ Value float64 }

func (c TestAComponent) Type() ComponentType { return TestAType }

type TestBComponent struct{ Value float64 }

func (c TestBComponent) Type() ComponentType { return TestBType }

type TestCComponent struct{ Value float64 }

func (c TestCComponent) Type() ComponentType { return TestCType }

type TestSystem struct{}

func (s *TestSystem) AddedToEngine(*Engine) error     { return nil }
func (s *TestSystem) RemovedFromEngine(*Engine) error { return nil }
func (s *TestSystem) Update(time.Duration) error      { return nil }

func TestCollectionObserver(t *testing.T) {
	engine := NewEngine()

	obs := NewCollection([]ComponentType{TestAType})
	s := &TestSystem{}
	var cnt int

	obs.Subscribe(s, func(en *Entity) {
		cnt--
	}, func(en *Entity) {
		cnt--
	})

	n := 10
	for i := 0; i < n; i++ {
		cnt += 2
		en := engine.CreateEntity(fmt.Sprintf("Entity %v", i))
		obs.add(en)
		obs.remove(en)
	}

	if cnt != 0 {
		t.Errorf("missed a notification, remaining: %v", cnt)
	}
}
