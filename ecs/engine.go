package ecs

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrNoSuchEntity         = errors.New("no such entity")
	ErrInvalidComponentType = errors.New("invalid component type")
)

// Engine collects and connects Systems with matching Entities
type Engine struct {
	eventLock     sync.RWMutex // TODO: replace mutex(es) with single goroutine? single monolithic event-scheduler...
	eventObserver []chan<- Message
	priorities    []SystemPriority

	entityLock      sync.RWMutex
	entityObservers map[*aspect][]chan<- Message // TODO: move channel slice inside of aspect singleton?
	entities        map[*entity][]*aspect        // TODO: replace Entity-pointers with int-ids and move component set/get-logic to engine

	entityToId []*entity
}

// Creates a new Engine
func NewEngine() *Engine {
	return &Engine{
		eventObserver: []chan<- Message{},
		priorities:    []SystemPriority{},

		entityObservers: map[*aspect][]chan<- Message{},
		entities:        map[*entity][]*aspect{},

		entityToId: []*entity{},
	}
}

// Create an Entity
func (e *Engine) CreateEntity(name string) Entity {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	en := &entity{
		Name:       name,
		engine:     e,
		components: map[ComponentType]Component{},
	}

	// add entity to entities map
	e.entities[en] = []*aspect{}
	e.entityToId = append(e.entityToId, en)

	return Entity(len(e.entityToId) - 1)
}

// Delete Entity from Engine and send RemoveEvents to all registered observers
func (e *Engine) DeleteEntity(id Entity) error {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	if len(e.entityToId) <= int(id) {
		return ErrNoSuchEntity
	}
	en := e.entityToId[id]
	if _, found := e.entities[en]; !found {
		return ErrNoSuchEntity
	}
	en.engine = nil
	for _, a := range e.entities[en] {
		for _, o := range e.entityObservers[a] {
			go func(c chan<- Message, id Entity) {
				c <- MessageEntityRemove{Removed: id}
			}(o, id)
			// TODO: waitgroup?
		}
	}
	delete(e.entities, en)
	e.entityToId = append(e.entityToId[:id], e.entityToId[id+1:]...)
	return nil
}

func (e *Engine) SetComponents(id Entity, components ...Component) error {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	if len(e.entityToId) <= int(id) {
		return ErrNoSuchEntity
	}
	en := e.entityToId[id]
	if en.Set(components...) {
		for _, a := range e.entities[en] {
			for _, o := range e.entityObservers[a] {
				go func(c chan<- Message, id Entity) {
					c <- MessageEntityUpdate{Updated: id}
				}(o, id)
				// TODO: waitgroup?
			}
		}
	} else {
		var already bool
		for a := range e.entityObservers {
			already = false
			for _, h := range e.entities[en] {
				if a == h {
					already = true
					break
				}
			}

			// add entity to matching collections slice
			if !already && a.accepts(en) {
				for _, o := range e.entityObservers[a] {
					go func(c chan<- Message, id Entity) {
						c <- MessageEntityAdd{Added: id}
					}(o, id)
					// TODO: waitgroup?
				}

				e.entities[en] = append(e.entities[en], a)
			}
		}
	}
	return nil
}

func (e *Engine) RemoveComponents(id Entity, types ...ComponentType) error {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	if len(e.entityToId) <= int(id) {
		return ErrNoSuchEntity
	}
	en := e.entityToId[id]
	en.Remove(types...)

	for i, a := range e.entities[en] {
		if !a.accepts(en) {
			// aspect does not accept entity anymore
			// remove aspect from entities slice
			copy(e.entities[en][i:], e.entities[en][i+1:])
			e.entities[en][len(e.entities[en])-1] = nil
			e.entities[en] = e.entities[en][:len(e.entities[en])-1]

			for _, o := range e.entityObservers[a] {
				go func(c chan<- Message, id Entity) {
					c <- MessageEntityRemove{Removed: id}
				}(o, id)
				// TODO: waitgroup?
			}
		}
	}
	return nil
}

/*
var c Component
var id Entity
if err := engine.GetComponent(id, &c);err != nil {
	log.Fatal(err)
}

func (e *Engine) GetComponent(id Entity, c Component) error {
	*c = en.Get(c.Type())
}
*/

func (e *Engine) GetComponent(id Entity, t ComponentType) (Component, error) {
	e.entityLock.RLock()
	defer e.entityLock.RUnlock()

	if len(e.entityToId) <= int(id) {
		return nil, ErrNoSuchEntity
	}
	en := e.entityToId[id]
	c := en.Get(t)
	if c == nil {
		return nil, ErrInvalidComponentType
	}
	return c, nil
}

func (e *Engine) Subscribe(f Filter, p SystemPriority, c chan<- Message) {
	if len(f.Aspect) > 0 {
		e.subscribeAspectEvent(c, f.Aspect...)
		return
	}
	e.subscribeEvent(c, p)
}

func (e *Engine) Unsubscribe(f Filter, c chan<- Message) {
	if len(f.Aspect) > 0 {
		e.unsubscribeAspectEvent(c, f.Aspect...)
		return
	}
	e.unsubscribeEvent(c)
}

func (e *Engine) subscribeEvent(c chan<- Message, prio SystemPriority) {
	e.eventLock.Lock()
	defer e.eventLock.Unlock()

	e.eventObserver = append(e.eventObserver, c)
	e.priorities = append(e.priorities, prio)
	e.sortObservers()
}

func (e *Engine) unsubscribeEvent(c chan<- Message) {
	e.eventLock.Lock()
	defer e.eventLock.Unlock()

	for i, o := range e.eventObserver {
		if o == c {
			copy(e.eventObserver[i:], e.eventObserver[i+1:])
			e.eventObserver[len(e.eventObserver)-1] = nil
			e.eventObserver = e.eventObserver[:len(e.eventObserver)-1]

			return
		}
	}
}

func (e *Engine) Publish(ev Message) {
	e.eventLock.RLock()
	defer e.eventLock.RUnlock()

	var wg sync.WaitGroup
	var cur SystemPriority = -1

	for i, o := range e.eventObserver {
		if p := e.priorities[i]; p != cur {
			// wait until previous observers are finished
			wg.Wait()
			cur = p
		}

		wg.Add(1)
		go func(c chan<- Message) {
			defer wg.Done()
			c <- ev
		}(o)
	}

	wg.Wait()
}

func (e *Engine) subscribeAspectEvent(c chan<- Message, types ...ComponentType) {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	// old aspect
	for a := range e.entityObservers {
		if a.equals(types) {

			// new subscribers of existing aspects must get all previosly added entities
			for id, en := range e.entityToId {
				for _, f := range e.entities[en] {
					if f == a {
						go func(c chan<- Message, id Entity) {
							c <- MessageEntityAdd{Added: id}
						}(c, Entity(id))
						// TODO: waitgroup?
						break
					}
				}
				// TODO: test post break
			}

			return
		}
	}

	// new aspect
	a := &aspect{
		types: types,
	}

	// add observer
	e.entityObservers[a] = append(e.entityObservers[a], c)

	// broadcast entities
	for id, en := range e.entityToId {
		if a.accepts(en) {
			e.entities[en] = append(e.entities[en], a)

			go func(c chan<- Message, id Entity) {
				c <- MessageEntityAdd{Added: id}
			}(c, Entity(id))
		}
	}
}

func (e *Engine) unsubscribeAspectEvent(c chan<- Message, types ...ComponentType) {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	for a, os := range e.entityObservers {
		if a.equals(types) {

			for i, o := range os {
				if o == c {
					copy(e.entityObservers[a][i:], e.entityObservers[a][i+1:])
					e.entityObservers[a][len(e.entityObservers[a])-1] = nil
					e.entityObservers[a] = e.entityObservers[a][:len(e.entityObservers[a])-1]

					return
				}
			}

			return
		}
	}
}

// byPriority attaches the methods of sort.Interface to []eventObservers, sorting in increasing order of priority
type byPriority struct {
	observer   []chan<- Message
	priorities []SystemPriority
}

func (a byPriority) Len() int { return len(a.observer) }
func (a byPriority) Swap(i, j int) {
	a.observer[i], a.observer[j] = a.observer[j], a.observer[i]
	a.priorities[i], a.priorities[j] = a.priorities[j], a.priorities[i]
}
func (a byPriority) Less(i, j int) bool {
	return a.priorities[i] < a.priorities[j]
}

func (e *Engine) sortObservers() {
	sort.Sort(byPriority{
		observer:   e.eventObserver,
		priorities: e.priorities,
	})
}
