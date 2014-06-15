package game

import (
	"sort"
)

type Priority int

const (
	PriorityFirst Priority = iota
	PriorityBeforeRender
	PriorityRender
	PriorityAfterRender
	PriorityLast
)

type subchan struct {
	c chan<- interface{}
	p Priority
}

type msgchan struct {
	c chan<- interface{}
	m interface{}
}

type Observer struct {
	sub   chan subchan
	unsub chan chan<- interface{}
	in    chan interface{}

	subs   []chan<- interface{}
	prio   map[chan<- interface{}]Priority
	update bool

	send    chan msgchan
	pending []msgchan
}

func NewObserver(newsub func() <-chan interface{}) *Observer {
	o := &Observer{
		sub:   make(chan subchan),
		unsub: make(chan chan<- interface{}),
		in:    make(chan interface{}),

		subs: make([]chan<- interface{}, 0),
		prio: make(map[chan<- interface{}]Priority),

		send:    make(chan msgchan),
		pending: make([]msgchan, 0),
	}

	// subscription
	go func() {
		var (
			sc subchan
			c  chan<- interface{}
			ok bool
			m  interface{}
		)
		for {
			select {
			case sc, ok = <-o.sub:
				if !ok {
					return
				}
				o.subs = append(o.subs, sc.c)
				o.prio[sc.c] = sc.p
				o.update = true

				if newsub != nil {
					for m := range newsub() {
						o.send <- msgchan{sc.c, m}
					}
				}
			case c = <-o.unsub:
				for i, f := range o.subs {
					if f == c {
						l := len(o.subs)
						copy(o.subs[i:], o.subs[i+1:])
						o.subs[l-1] = nil
						o.subs = o.subs[:l-1]
						break
					}
				}
			case m = <-o.in:
				if o.update {
					sort.Sort(byPriority{o.subs, o.prio})
					o.update = false
				}
				for _, c = range o.subs {
					o.send <- msgchan{c, m}
				}
			}
		}
	}()

	// non blocking send
	go func() {
		var (
			mc msgchan
			ok bool
		)
		for {
			if len(o.pending) == 0 {
				mc, ok = <-o.send
				if !ok {
					return
				}
				o.pending = append(o.pending, mc)
			}
			select {
			case mc, ok = <-o.send:
				if !ok {
					return
				}
				o.pending = append(o.pending, mc)
			case o.pending[0].c <- o.pending[0].m:
				o.pending = o.pending[1:]
			}
		}
	}()

	return o
}

func (o *Observer) Subscribe(c chan<- interface{}, p Priority) {
	o.sub <- subchan{c, p}
}

func (o *Observer) Unsubscribe(c chan<- interface{}) {
	o.unsub <- c
}

func (o *Observer) Publish(msg interface{}) {
	o.in <- msg
}

func (o *Observer) PublishAndWait(msg interface{}) {
	o.Publish(msg)
	// TODO: wait for completion of all subscribers
}

func (o *Observer) Close() {
	close(o.sub)
	close(o.send)
}

// byPriority attaches the methods of sort.Interface to []subs, sorting in increasing order of map[]prio
type byPriority struct {
	subs []chan<- interface{}
	prio map[chan<- interface{}]Priority
}

func (s byPriority) Len() int {
	return len(s.subs)
}
func (s byPriority) Swap(i, j int) {
	s.subs[i], s.subs[j] = s.subs[j], s.subs[i]
}
func (s byPriority) Less(i, j int) bool {
	return s.prio[s.subs[i]] < s.prio[s.subs[j]]
}
