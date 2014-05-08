package ecs

import (
	"time"
)

type SystemPriority int

type singleAspectSystem struct {
	engine   *Engine
	messages chan Message

	types    []ComponentType
	entities []Entity
}

// Creates a System with a single Components-Aspect
// the supplied update function is invoked for all entites of this aspect
// after receiving an UpdateEvent from the Engine
func SingleAspectSystem(e *Engine, prio SystemPriority, update func(time.Duration, Entity), types []ComponentType) *singleAspectSystem {
	c := make(chan Message)
	s := &singleAspectSystem{
		engine:   e,
		messages: c,
		types:    types,
	}

	go func() {
		s.Restart(prio)

		for event := range c {
			switch e := event.(type) {
			case MessageEntityAdd:
				s.entities = append(s.entities, e.Added)
			case MessageEntityRemove:
				for i, f := range s.entities {
					if f == e.Removed {
						s.entities = append(s.entities[:i], s.entities[i+1:]...)
						// TODO: break?
					}
				}

			case MessageUpdate:
				for _, en := range s.entities {
					update(e.Delta, en)
				}
			}
		}
	}()

	return s
}

func (s *singleAspectSystem) Restart(prio SystemPriority) {
	s.engine.Subscribe(Filter{Types: []MessageType{UpdateMessageType}}, prio, s.messages)
	s.engine.Subscribe(Filter{Aspect: s.types}, prio, s.messages)
}

func (s *singleAspectSystem) Stop() {
	s.engine.Unsubscribe(Filter{Types: []MessageType{UpdateMessageType}}, s.messages)
	s.engine.Unsubscribe(Filter{Aspect: s.types}, s.messages)

	s.entities = []Entity{} // TODO: gc?
}

type updateSystem struct {
	engine   *Engine
	messages chan Message
}

// Creates a simple update loop System
func UpdateSystem(e *Engine, prio SystemPriority, update func(time.Duration)) *updateSystem {
	c := make(chan Message)
	s := &updateSystem{
		engine:   e,
		messages: c,
	}

	go func() {
		s.Restart(prio)

		for event := range c {
			switch e := event.(type) {
			case MessageUpdate:
				update(e.Delta)
			}
		}
	}()

	return s
}

func (s *updateSystem) Restart(prio SystemPriority) {
	s.engine.Subscribe(Filter{Types: []MessageType{UpdateMessageType}}, prio, s.messages)
}

func (s *updateSystem) Stop() {
	s.engine.Unsubscribe(Filter{Types: []MessageType{UpdateMessageType}}, s.messages)
}
