package ecs

import (
	"time"
)

type singleAspectSystem struct {
	engine *Engine
	events chan Event

	types    []ComponentType
	entities []*Entity
}

// Creates a System with a single Components-Aspect
// the supplied update function is invoked for all entites of this aspect
// after receiving an UpdateEvent from the Engine
func SingleAspectSystem(e *Engine, prio int, update func(time.Duration, *Entity), types []ComponentType) *singleAspectSystem {
	c := make(chan Event)
	s := &singleAspectSystem{
		engine: e,
		events: c,
		types:  types,
	}

	go func() {
		s.Restart(prio)

		for event := range c {
			switch e := event.(type) {
			case EntityAddEvent:
				s.entities = append(s.entities, e.Added)
			case EntityRemoveEvent:
				for i, f := range s.entities {
					if f == e.Removed {
						copy(s.entities[i:], s.entities[i+1:])
						s.entities[len(s.entities)-1] = nil
						s.entities = s.entities[:len(s.entities)-1]
					}
				}

			case UpdateEvent:
				for _, en := range s.entities {
					update(e.Delta, en)
				}
			}
		}
	}()

	return s
}

func (s *singleAspectSystem) Restart(prio int) {
	s.engine.SubscribeEvent(s.events, prio)
	s.engine.SubscribeAspectEvent(s.events, s.types...)
}

func (s *singleAspectSystem) Stop() {
	s.engine.UnsubscribeEvent(s.events)
	s.engine.UnsubscribeAspectEvent(s.events, s.types...)
	s.entities = []*Entity{} // TODO: gc?
}

type updateSystem struct {
	engine *Engine
	events chan Event
}

// Creates a simple update loop System
func UpdateSystem(e *Engine, prio int, update func(time.Duration)) *updateSystem {
	c := make(chan Event)
	s := &updateSystem{
		engine: e,
		events: c,
	}

	go func() {
		s.Restart(prio)

		for event := range c {
			switch e := event.(type) {
			case UpdateEvent:
				update(e.Delta)
			}
		}
	}()

	return s
}

func (s *updateSystem) Restart(prio int) {
	s.engine.SubscribeEvent(s.events, prio)
}

func (s *updateSystem) Stop() {
	s.engine.UnsubscribeEvent(s.events)
}
