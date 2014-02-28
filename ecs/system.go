package ecs

import (
	//"fmt"
	//"reflect"
	"time"
)

type SystemPriority int

type System interface {
	Priority() SystemPriority
	AddedToEngine(*Engine) error
	RemovedFromEngine(*Engine) error
	Update(time.Duration) error
}

/*
TODO:

type magicSystem struct {
	priority SystemPriority
	engine   *Engine

	types      []reflect.Type
	targets    map[*Entity]map[reflect.Type]reflect.Value
	updatefunc reflect.Value

	initfunc       func() error
	cleanupfunc    func() error
	preupdatefunc  func()
	postupdatefunc func()
}

// Creates a new System that takes Entities with the set of components determined by the supplied update function
func newMagicSystem(update interface{}) *magicSystem {
	t := reflect.TypeOf(update)

	s := &magicSystem{
		types:      make([]reflect.Type, t.NumIn()),
		targets:    map[*Entity]map[reflect.Type]reflect.Value{},
		updatefunc: reflect.ValueOf(update),
	}

	for i := 0; i < t.NumIn(); i++ {
		s.types[i] = t.In(i)
	}

	return s
}

// Set Engine
func (s *magicSystem) setEngine(engine *Engine) { s.engine = engine }

// Set init function that is called before the system is added to an engine
func (s *magicSystem) SetInitFunc(f func() error) { s.initfunc = f }

// Set clean up function that is called before the system is removed from an engine
func (s *magicSystem) SetCleanupFunc(f func() error) { s.cleanupfunc = f }

// Set function which is executed every time before the Entities are updated
func (s *magicSystem) SetPreUpdateFunc(f func()) { s.preupdatefunc = f }

// Set function which is executed every time after the Entities are updated
func (s *magicSystem) SetPostUpdateFunc(f func()) { s.postupdatefunc = f }

// Initialize System and call user supplied init function
func (s *magicSystem) init() error {
	if s.initfunc != nil {
		return s.initfunc()
	}
	return nil
}

// Call user supplied clean up function and clean up System, release mapped pointers
func (s *magicSystem) cleanup() error {
	if s.cleanupfunc != nil {
		if err := s.cleanupfunc(); err != nil {
			return err
		}
	}

	// TODO: reflect.Value(pointer) possible memory leak?
	for i := range s.targets {
		//s.targets[i].Reset()
		s.targets[i] = nil
	}
	s.targets = map[*Entity]map[reflect.Type]reflect.Value{}

	return nil
}

// Adds the Entity to the System if it contains the specific set of components and returns true.
// Returns false if components do not match.
func (s *magicSystem) add(entity *Entity) bool {
	//fmt.Printf("adding Entity %s to System %s\n", entity.Name, s.Name)

	//fmt.Println("using set", s.types)
	//fmt.Println("entity set", entity.Components)

	subset := map[reflect.Type]reflect.Value{}

	for _, t := range s.types {
		switch t.String() { // TODO: there must be a better way than string comparison!
		case "*ecs.Engine", "ecs.System", "time.Duration", "*ecs.Entity":
			//fmt.Println("ignoring", t.String())
		default:
			r := entity.Get(t)
			if r == nil {
				//fmt.Println("set is not matching")
				return false
			}
			subset[t] = reflect.ValueOf(r)
		}
	}

	//fmt.Println("resulting set", subset)

	s.targets[entity] = subset
	return true
}

// Remove Entity from System
func (s *magicSystem) remove(entity *Entity) {
	delete(s.targets, entity)
}

// Call update function on all Entities
func (s *magicSystem) update(td time.Duration) error {
	if s.preupdatefunc != nil {
		s.preupdatefunc()
	}

	if s.postupdatefunc != nil {
		defer s.postupdatefunc()
	}

	args := make([]reflect.Value, len(s.types))

	for i, t := range s.types {
		switch t.String() {
		case "*ecs.Engine":
			args[i] = reflect.ValueOf(s.engine)
			//fmt.Println("set engine")
		case "ecs.System":
			args[i] = reflect.ValueOf(s)
			//fmt.Println("set system")
		case "time.Duration":
			args[i] = reflect.ValueOf(td)
			//fmt.Println("set time")
		default:
		}
	}

	// TODO: possible opportunity for parallelization?
	for e, t := range s.targets {
		//fmt.Printf("updating entity %s\n", e.Name)

		for i, tt := range s.types {
			switch tt.String() {
			case "*ecs.Engine", "*ecs.System", "time.Duration":
			case "*ecs.Entity":
				args[i] = reflect.ValueOf(e)
			default:
				args[i] = t[tt]
			}

			if !args[i].IsValid() {
				return fmt.Errorf("Value not found for type %s", tt)
			}
		}

		//fmt.Println("updating:", t)
		//fmt.Println(s.update)
		//fmt.Println(args)

		// TODO: break if error returned?
		s.updatefunc.Call(args)
	}

	return nil
}
*/

type collectionSystem struct {
	priority SystemPriority
	types    []ComponentType
	update   func(time.Duration, *Entity)

	engine     *Engine
	collection *Collection
}

// Creates a System with a single Collection of Components
func CollectionSystem(update func(time.Duration, *Entity), p SystemPriority, types []ComponentType) System {
	return &collectionSystem{
		priority: p,
		types:    types,
		update:   update,
	}
}

func (s *collectionSystem) Priority() SystemPriority {
	return s.priority
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
	for _, e := range s.collection.Entities() {
		s.update(delta, e)
	}

	return nil
}

type updateSystem struct {
	priority SystemPriority
	update   func(time.Duration)
}

// Creates a simple update loop System without a Collections
func UpdateSystem(update func(time.Duration), p SystemPriority) System {
	return &updateSystem{
		priority: p,
		update:   update,
	}
}

func (s *updateSystem) Priority() SystemPriority        { return s.priority }
func (s *updateSystem) AddedToEngine(*Engine) error     { return nil }
func (s *updateSystem) RemovedFromEngine(*Engine) error { return nil }

func (s *updateSystem) Update(delta time.Duration) error {
	s.update(delta)
	return nil
}
