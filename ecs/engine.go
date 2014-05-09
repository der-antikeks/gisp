package ecs

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrNoSuchEntity    = errors.New("ecs: no such entity")
	ErrNoSuchComponent = errors.New("ecs: no such component")
)

// Engine collects and connects Systems with matching Entities
type Engine struct {
	eventLock     sync.RWMutex // TODO: replace mutex(es) with single goroutine? single monolithic event-scheduler...
	eventObserver []chan<- Message
	priorities    []SystemPriority

	entityLock       sync.RWMutex
	nextEntity       int
	entityComponents map[int]map[ComponentType]Component
	entityAspects    map[int][]*aspect
	entityObservers  map[*aspect][]chan<- Message // TODO: move channel slice inside of aspect singleton?
}

// Creates a new Engine
func NewEngine() *Engine {
	return &Engine{
		eventObserver: []chan<- Message{},
		priorities:    []SystemPriority{},

		nextEntity:       1,
		entityComponents: map[int]map[ComponentType]Component{},
		entityAspects:    map[int][]*aspect{},
		entityObservers:  map[*aspect][]chan<- Message{},
	}
}

// Create a new Entity
func (e *Engine) Entity() Entity {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	id := e.nextEntity
	e.nextEntity++

	e.entityComponents[id] = map[ComponentType]Component{}
	e.entityAspects[id] = []*aspect{}

	return Entity(id)
}

// Delete Entity from Engine and send RemoveEvents to all registered observers
func (e *Engine) Delete(en Entity) error {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	id := int(en)
	if _, ok := e.entityComponents[id]; !ok {
		return ErrNoSuchEntity
	}

	for _, a := range e.entityAspects[id] {
		for _, o := range e.entityObservers[a] {
			go func(c chan<- Message, en Entity) {
				c <- MessageEntityRemove{Removed: en}
			}(o, en)
		}
	}

	delete(e.entityComponents, id)
	delete(e.entityAspects, id)
	return nil
}

func (e *Engine) Query(types ...ComponentType) []Entity {
	e.entityLock.RLock()
	defer e.entityLock.RUnlock()

	ret := []Entity{}
	for id, ecs := range e.entityComponents {
		found := true
		for _, t := range types {
			if _, ok := ecs[t]; !ok {
				found = false
				break
			}
		}
		if found {
			ret = append(ret, Entity(id))
		}
	}
	return ret
}

func (e *Engine) componentTypes(en Entity) []ComponentType {
	//e.entityLock.RLock()
	//defer e.entityLock.RUnlock()

	id := int(en)
	if _, ok := e.entityComponents[id]; !ok {
		return nil
	}

	ret := make([]ComponentType, 0, len(e.entityComponents[id]))
	for t := range e.entityComponents[id] {
		ret = append(ret, t)
	}
	return ret
}

func (e *Engine) Set(en Entity, components ...Component) error {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	id := int(en)
	if _, ok := e.entityComponents[id]; !ok {
		return ErrNoSuchEntity
	}

	var updated bool
	for _, c := range components {
		if !updated {
			if _, ok := e.entityComponents[id][c.Type()]; ok {
				updated = true
			}
		}

		e.entityComponents[id][c.Type()] = c
	}

	// update old aspect observers
	if updated {
		for _, a := range e.entityAspects[id] {
			for _, o := range e.entityObservers[a] {
				go func(c chan<- Message, en Entity) {
					c <- MessageEntityUpdate{Updated: en}
				}(o, en)
			}
		}
	}

	// add new aspect observers
	var already bool
	for a := range e.entityObservers {
		already = false
		for _, h := range e.entityAspects[id] {
			if a == h {
				already = true
				break
			}
		}

		// add aspect to entity
		if !already && a.accepts(e.componentTypes(en)) {
			for _, o := range e.entityObservers[a] {
				go func(c chan<- Message, en Entity) {
					c <- MessageEntityAdd{Added: en}
				}(o, en)
			}

			e.entityAspects[id] = append(e.entityAspects[id], a)
		}
	}

	return nil
}

func (e *Engine) Remove(en Entity, types ...ComponentType) error {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	id := int(en)
	if _, ok := e.entityComponents[id]; !ok {
		return ErrNoSuchEntity
	}

	for _, t := range types {
		/*
			if _, ok := e.entityComponents[id]; !ok {
				return ErrNoSuchComponent
			}
		*/
		delete(e.entityComponents[id], t)
	}

	for i, a := range e.entityAspects[id] {
		if !a.accepts(e.componentTypes(en)) {
			// aspect does not accept entity anymore
			// remove aspect from entities slice
			copy(e.entityAspects[id][i:], e.entityAspects[id][i+1:])
			e.entityAspects[id][len(e.entityAspects[id])-1] = nil
			e.entityAspects[id] = e.entityAspects[id][:len(e.entityAspects[id])-1]

			for _, o := range e.entityObservers[a] {
				go func(c chan<- Message, en Entity) {
					c <- MessageEntityRemove{Removed: en}
				}(o, en)
			}
		}
	}
	return nil
}

func (e *Engine) Get(en Entity, t ComponentType) (Component, error) {
	e.entityLock.RLock()
	defer e.entityLock.RUnlock()

	id := int(en)
	if _, ok := e.entityComponents[id]; !ok {
		return nil, ErrNoSuchEntity
	}

	c, ok := e.entityComponents[id][t]
	if !ok {
		return nil, ErrNoSuchComponent
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
			for id, as := range e.entityAspects {
				for _, f := range as {
					if f == a {
						go func(c chan<- Message, en Entity) {
							c <- MessageEntityAdd{Added: en}
						}(c, Entity(id))
						break
					}
				}
				// TODO: test post break
			}

			e.entityObservers[a] = append(e.entityObservers[a], c)
			return
		}
	}

	// new aspect
	a := &aspect{
		types: types,
	}

	// add observer
	e.entityObservers[a] = append(e.entityObservers[a], c)

	// add aspect to entity
	for id := range e.entityAspects {
		en := Entity(id)
		if a.accepts(e.componentTypes(en)) {
			e.entityAspects[id] = append(e.entityAspects[id], a)

			// send entity to observer
			go func(c chan<- Message, en Entity) {
				c <- MessageEntityAdd{Added: en}
			}(c, en)
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
