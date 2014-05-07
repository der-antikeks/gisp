package ecs

import (
	"sort"
	"sync"
)

// Engine collects and connects Systems with matching Entities
type Engine struct {
	eventLock     sync.RWMutex // TODO: replace mutex(es) with single goroutine? single monolithic event-scheduler...
	eventObserver []chan<- Message
	priorities    []SystemPriority

	entityLock      sync.RWMutex
	entityObservers map[*aspect][]chan<- Message // TODO: move channel slice inside of aspect singleton?
	entities        map[*Entity][]*aspect        // TODO: replace Entity-pointers with int-ids and move component set/get-logic to engine
}

// Creates a new Engine
func NewEngine() *Engine {
	return &Engine{
		eventObserver: []chan<- Message{},
		priorities:    []SystemPriority{},

		entityObservers: map[*aspect][]chan<- Message{},
		entities:        map[*Entity][]*aspect{},
	}
}

// Create an Entity
func (e *Engine) CreateEntity(name string) *Entity {
	en := &Entity{
		Name:       name,
		engine:     e,
		components: map[ComponentType]Component{},
	}

	// add entity to entities map
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	e.entities[en] = []*aspect{}
	return en
}

// Delete Entity from Engine and send RemoveEvents to all registered observers
func (e *Engine) DeleteEntity(en *Entity) {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	if _, found := e.entities[en]; !found {
		return
	}
	en.engine = nil
	for _, a := range e.entities[en] {
		for _, o := range e.entityObservers[a] {
			go func(c chan<- Message, en *Entity) {
				c <- MessageEntityRemove{Removed: en}
			}(o, en)
			// TODO: waitgroup?
		}
	}
	delete(e.entities, en)
}

// Called by the Entity whose components are removed
func (e *Engine) entityRemovedComponent(en *Entity) {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

	for i, a := range e.entities[en] {
		if !a.accepts(en) {
			// aspect does not accept entity anymore
			// remove aspect from entities slice
			copy(e.entities[en][i:], e.entities[en][i+1:])
			e.entities[en][len(e.entities[en])-1] = nil
			e.entities[en] = e.entities[en][:len(e.entities[en])-1]

			for _, o := range e.entityObservers[a] {
				go func(c chan<- Message, en *Entity) {
					c <- MessageEntityRemove{Removed: en}
				}(o, en)
				// TODO: waitgroup?
			}
		}
	}
}

// Called by the Entity, if components are updated
func (e *Engine) entityUpdatedComponent(en *Entity) {
	e.entityLock.RLock()
	defer e.entityLock.RUnlock()

	for _, a := range e.entities[en] {
		for _, o := range e.entityObservers[a] {
			go func(c chan<- Message, en *Entity) {
				c <- MessageEntityUpdate{Updated: en}
			}(o, en)
			// TODO: waitgroup?
		}
	}
}

// Called by the Entity, if components are added
func (e *Engine) entityAddedComponent(en *Entity) {
	e.entityLock.Lock()
	defer e.entityLock.Unlock()

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
				go func(c chan<- Message, en *Entity) {
					c <- MessageEntityAdd{Added: en}
				}(o, en)
				// TODO: waitgroup?
			}

			e.entities[en] = append(e.entities[en], a)
		}
	}
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
			for en, as := range e.entities {
				for _, f := range as {
					if f == a {
						go func(c chan<- Message, en *Entity) {
							c <- MessageEntityAdd{Added: en}
						}(c, en)
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
	for en := range e.entities {
		if a.accepts(en) {
			e.entities[en] = append(e.entities[en], a)

			go func(c chan<- Message, en *Entity) {
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
