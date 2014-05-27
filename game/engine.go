package game

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrNoSuchEntity    = errors.New("ecs: no such entity")
	ErrNoSuchComponent = errors.New("ecs: no such component")
)

type msgchan struct {
	c chan<- Message
	m Message
}

// Engine collects and connects Systems with matching Entities
type Engine struct {
	lock sync.RWMutex

	observers  map[*aspect]map[MessageType][]chan<- Message
	priorities map[chan<- Message]SystemPriority
	pending    []msgchan
	send       chan msgchan

	nextEntity       int
	deletedEntities  []int
	entityComponents map[int]map[ComponentType]Component
	entityAspects    map[int][]*aspect
}

// Creates a new Engine
func NewEngine() *Engine {
	e := &Engine{
		observers:  map[*aspect]map[MessageType][]chan<- Message{nil: map[MessageType][]chan<- Message{}},
		priorities: map[chan<- Message]SystemPriority{},
		pending:    []msgchan{},
		send:       make(chan msgchan),

		nextEntity:       1,
		deletedEntities:  []int{},
		entityComponents: map[int]map[ComponentType]Component{},
		entityAspects:    map[int][]*aspect{},
	}

	go func() {
		var (
			mc msgchan
			ok bool
		)
		for {
			if len(e.pending) == 0 {
				mc, ok = <-e.send
				if !ok {
					return
				}
				e.pending = append(e.pending, mc)
			}
			select {
			case mc, ok = <-e.send:
				if !ok {
					return
				}
				e.pending = append(e.pending, mc)
			case e.pending[0].c <- e.pending[0].m:
				e.pending = e.pending[1:]
			}
		}
	}()

	return e
}

// Create a new Entity
func (e *Engine) Entity() Entity {
	e.lock.Lock()
	defer e.lock.Unlock()

	var id int
	if l := len(e.deletedEntities); l > 0 {
		id, e.deletedEntities = e.deletedEntities[l-1], e.deletedEntities[:l-1]
	} else {
		id = e.nextEntity
		e.nextEntity++
	}

	e.entityComponents[id] = map[ComponentType]Component{}
	e.entityAspects[id] = []*aspect{}

	return Entity(id)
}

// Delete Entity from Engine and send RemoveEvents to all registered observers
func (e *Engine) Delete(en Entity) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	id := int(en)
	if _, ok := e.entityComponents[id]; !ok {
		return ErrNoSuchEntity
	}

	for _, a := range e.entityAspects[id] {
		e.publish(MessageEntityRemove{Removed: en}, a)
		e.publish(MessageEntityRemove{Removed: en}, nil)
	}

	delete(e.entityComponents, id)
	delete(e.entityAspects, id)
	e.deletedEntities = append(e.deletedEntities, id)
	return nil
}

func (e *Engine) Query(types ...ComponentType) []Entity {
	e.lock.RLock()
	defer e.lock.RUnlock()

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
	// unexported func, lock is/mustbe handled by caller
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
	e.lock.Lock()
	defer e.lock.Unlock()

	id := int(en)
	if _, ok := e.entityComponents[id]; !ok {
		return ErrNoSuchEntity
	}

	// add/update component of entity
	var updated, added bool
	for _, c := range components {
		if !updated || !added {
			if _, ok := e.entityComponents[id][c.Type()]; ok {
				updated = true
			} else {
				added = true
			}
		}

		e.entityComponents[id][c.Type()] = c
	}

	// send update to old aspect observers
	if updated {
		for _, a := range e.entityAspects[id] {
			e.publish(MessageEntityUpdate{Updated: en}, a)
			e.publish(MessageEntityUpdate{Updated: en}, nil)
		}
	}

	// add new aspect observers
	if added {
		var already bool
		for a := range e.observers {
			if a == nil {
				continue
			}

			already = false
			for _, h := range e.entityAspects[id] {
				if a == h {
					already = true
					break
				}
			}

			// add aspect to entity
			if !already && a.accepts(e.componentTypes(en)) {
				e.publish(MessageEntityAdd{Added: en}, a)
				e.publish(MessageEntityAdd{Added: en}, nil)

				e.entityAspects[id] = append(e.entityAspects[id], a)
			}
		}
	}

	return nil
}

func (e *Engine) Remove(en Entity, types ...ComponentType) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	id := int(en)
	if _, ok := e.entityComponents[id]; !ok {
		return ErrNoSuchEntity
	}

	for _, t := range types {
		delete(e.entityComponents[id], t)
	}

	for i, a := range e.entityAspects[id] {
		if !a.accepts(e.componentTypes(en)) {
			// aspect does not accept entity anymore
			// remove aspect from entities slice
			copy(e.entityAspects[id][i:], e.entityAspects[id][i+1:])
			e.entityAspects[id][len(e.entityAspects[id])-1] = nil
			e.entityAspects[id] = e.entityAspects[id][:len(e.entityAspects[id])-1]

			e.publish(MessageEntityRemove{Removed: en}, a)
			e.publish(MessageEntityRemove{Removed: en}, nil)
		}
	}
	return nil
}

func (e *Engine) Get(en Entity, t ComponentType) (Component, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

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

func (e *Engine) Subscribe(f Filter, prio SystemPriority, c chan<- Message) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.priorities[c] = prio
	defer e.sortObservers()

	// no aspect message observer or entity message for all aspects
	if len(f.Aspect) == 0 {
		for _, t := range f.Types {
			e.observers[nil][t] = append(e.observers[nil][t], c)
		}
		return
	}

	var hasAddMessage bool
	for _, t := range f.Types {
		if t == EntityAddMessageType {
			hasAddMessage = true
			break
		}
	}

	// old aspect
	for a := range e.observers {
		if a != nil && a.equals(f.Aspect) {
			// new subscribers of existing aspects must get all previously added entities
			if hasAddMessage {
				for id, as := range e.entityAspects {
					en := Entity(id)
					for _, f := range as {
						if f == a {
							e.send <- msgchan{c, MessageEntityAdd{Added: en}}
							break
						}
					}
				}
			}

			for _, t := range f.Types {
				e.observers[a][t] = append(e.observers[a][t], c)
			}
			return
		}
	}

	// new aspect
	a := &aspect{f.Aspect}

	// add observer
	e.observers[a] = map[MessageType][]chan<- Message{}
	for _, t := range f.Types {
		e.observers[a][t] = append(e.observers[a][t], c)
	}

	// add aspect to entity
	for id := range e.entityAspects {
		en := Entity(id)
		if a.accepts(e.componentTypes(en)) {
			e.entityAspects[id] = append(e.entityAspects[id], a)

			// send entity to observer
			if hasAddMessage {
				e.send <- msgchan{c, MessageEntityAdd{Added: en}}
			}
		}
	}
}

func (e *Engine) Unsubscribe(f Filter, c chan<- Message) {
	e.lock.Lock()
	defer e.lock.Unlock()

	// no aspect message observer or entity message for all aspects
	if len(f.Aspect) == 0 {
		for _, t := range f.Types {
			for i, o := range e.observers[nil][t] {
				if o == c {
					copy(e.observers[nil][t][i:], e.observers[nil][t][i+1:])
					e.observers[nil][t][len(e.observers[nil][t])-1] = nil
					e.observers[nil][t] = e.observers[nil][t][:len(e.observers[nil][t])-1]

					break
				}
			}
		}
		return
	}

	for a := range e.observers {
		if a != nil && a.equals(f.Aspect) {
			for _, t := range f.Types {
				for i, o := range e.observers[a][t] {
					if o == c {
						copy(e.observers[a][t][i:], e.observers[a][t][i+1:])
						e.observers[a][t][len(e.observers[a][t])-1] = nil
						e.observers[a][t] = e.observers[a][t][:len(e.observers[a][t])-1]

						break
					}
				}
			}
			return
		}
	}
}

func (e *Engine) Publish(msg Message) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	// aspect-less observers or message-types
	e.publish(msg, nil)

	// external entity messages
	if emsg, ok := msg.(EntityMessage); ok {
		id := int(emsg.Entity())
		for _, a := range e.entityAspects[id] {
			e.publish(msg, a)
		}
	}
}

// aspect observers, mostly entity-messages
func (e *Engine) publish(msg Message, a *aspect) {
	for _, o := range e.observers[a][msg.Type()] {
		e.send <- msgchan{o, msg}
	}
}

// byPriority attaches the methods of sort.Interface to []eventObservers, sorting in increasing order of priority
type byPriority struct {
	observers  []chan<- Message
	priorities map[chan<- Message]SystemPriority
}

func (a byPriority) Len() int {
	return len(a.observers)
}
func (a byPriority) Swap(i, j int) {
	a.observers[i], a.observers[j] = a.observers[j], a.observers[i]
}
func (a byPriority) Less(i, j int) bool {
	return a.priorities[a.observers[i]] < a.priorities[a.observers[j]]
}

func (e *Engine) sortObservers() {
	for a, as := range e.observers {
		for t := range as {
			sort.Sort(byPriority{
				observers:  e.observers[a][t],
				priorities: e.priorities,
			})
		}
	}
}
