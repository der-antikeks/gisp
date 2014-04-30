package ecs

import (
	"time"
)

type SystemPriority int

type System interface {
	AddedToEngine(*Engine) error
	RemovedFromEngine(*Engine) error
	Update(time.Duration) error
}

type collectionSystem struct {
	types  []ComponentType
	update func(time.Duration, *Entity) error

	engine     *Engine
	collection *collection
}

// Creates a System with a single Collection of Components
func CollectionSystem(update func(time.Duration, *Entity) error, types []ComponentType) System {
	return &collectionSystem{
		types:  types,
		update: update,
	}
}

func (s *collectionSystem) AddedToEngine(e *Engine) error {
	s.engine = e
	s.collection = e.Collection(s.types...)
	return nil
}

func (s *collectionSystem) RemovedFromEngine(*Engine) error {
	s.engine = nil
	s.collection = nil
	return nil
}

func (s *collectionSystem) Update(delta time.Duration) error {
	if s.collection == nil {
		return nil
	}

	for _, e := range s.collection.Entities() {
		if err := s.update(delta, e); err != nil {
			return err
		}
	}

	return nil
}

type updateSystem struct {
	update func(time.Duration) error
}

// Creates a simple update loop System without a collection
func UpdateSystem(update func(time.Duration) error) System {
	return &updateSystem{
		update: update,
	}
}

func (s *updateSystem) AddedToEngine(*Engine) error     { return nil }
func (s *updateSystem) RemovedFromEngine(*Engine) error { return nil }

func (s *updateSystem) Update(delta time.Duration) error {
	return s.update(delta)
}
