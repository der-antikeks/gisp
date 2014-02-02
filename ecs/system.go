package ecs

import (
	"fmt"
	"reflect"
	"time"
)

type System struct {
	Name     string
	Priority int
	engine   *Engine

	types   []reflect.Type
	targets map[*Entity]map[reflect.Type]reflect.Value
	update  reflect.Value

	init       func() error
	cleanup    func() error
	preupdate  func()
	postupdate func()
}

// Creates a new System that takes Entities with the set of components determined by the supplied update function
func NewSystem(name string, update interface{}) *System {
	t := reflect.TypeOf(update)

	s := &System{
		Name: name,

		types:   make([]reflect.Type, t.NumIn()),
		targets: map[*Entity]map[reflect.Type]reflect.Value{},
		update:  reflect.ValueOf(update),
	}

	for i := 0; i < t.NumIn(); i++ {
		s.types[i] = t.In(i)
	}

	return s
}

// Set Engine
func (s *System) SetEngine(engine *Engine) { s.engine = engine }

// Set init function
func (s *System) SetInitFunc(f func() error) { s.init = f }

// Set clean up function
func (s *System) SetCleanupFunc(f func() error) { s.cleanup = f }

// Set function which is executed before the update function
func (s *System) SetPreUpdateFunc(f func()) { s.preupdate = f }

// Set function which is executed after the update function
func (s *System) SetPostUpdateFunc(f func()) { s.postupdate = f }

// Initialize System and call user supplied init function
func (s *System) Init() error {
	if s.init != nil {
		return s.init()
	}
	return nil
}

// Call user supplied clean up function and clean up System, release mapped pointers
func (s *System) Cleanup() error {
	if s.cleanup != nil {
		if err := s.cleanup(); err != nil {
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
func (s *System) Add(entity *Entity) bool {
	//fmt.Printf("adding Entity %s to System %s\n", entity.Name, s.Name)

	//fmt.Println("using set", s.types)
	//fmt.Println("entity set", entity.Components)

	subset := map[reflect.Type]reflect.Value{}

	for _, t := range s.types {
		switch t.String() { // TODO: there must be a better way than string comparison!
		case "*ecs.Engine", "*ecs.System", "time.Duration", "*ecs.Entity":
			//fmt.Println("ignoring", t.String())
		default:
			r := entity.Components.Get(t)
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
func (s *System) Remove(entity *Entity) {
	delete(s.targets, entity)
}

// Call update function on all Entities
func (s *System) Update(td time.Duration) error {
	if s.preupdate != nil {
		s.preupdate()
	}

	if s.postupdate != nil {
		defer s.postupdate()
	}

	args := make([]reflect.Value, len(s.types))

	for i, t := range s.types {
		switch t.String() {
		case "*ecs.Engine":
			args[i] = reflect.ValueOf(s.engine)
			//fmt.Println("set engine")
		case "*ecs.System":
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
		/*ret :=*/ s.update.Call(args)
		/*
			if ret != nil {
				fmt.Println("returned:", ret)
			}
		*/
	}

	return nil
}

//type Component interface{}
/*
func ResetComponent(c interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("could not reset component %T: %s", c, e)
		}
	}()

	value := reflect.ValueOf(c).Elem()
	zero := reflect.Zero(value.Type())
	value.Set(zero)

	return nil
}
*/
