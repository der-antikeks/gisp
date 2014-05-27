package game

import (
	"sort"
)

type subchan struct {
	c chan<- Message
	p int
}

type msgchan struct {
	c chan<- Message
	m Message
}

type Observer struct {
	sub   chan subchan
	unsub chan chan<- Message
	in    chan Message

	subs   []chan<- Message
	prio   map[chan<- Message]int
	update bool

	send    chan msgchan
	pending []msgchan
}

func NewObserver() *Observer {
	o := &Observer{
		sub:   make(chan subchan),
		unsub: make(chan chan<- Message),
		in:    make(chan Message),

		subs: make([]chan<- Message, 0),
		prio: make(map[chan<- Message]int),

		send:    make(chan msgchan),
		pending: make([]msgchan, 0),
	}

	// subscription
	go func() {
		var (
			sc subchan
			c  chan<- Message
			ok bool
			m  Message
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
					sort.Sort(o)
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

func (o *Observer) Subscribe(c chan<- Message, prio int) {
	o.sub <- subchan{c, prio}
}

func (o *Observer) Unsubscribe(c chan<- Message) {
	o.unsub <- c
}

func (o *Observer) Publish(msg Message) {
	o.in <- msg
}

func (o *Observer) Close() {
	close(o.sub)
	close(o.send)
}

func (o *Observer) Len() int {
	return len(o.subs)
}
func (o *Observer) Swap(i, j int) {
	o.subs[i], o.subs[j] = o.subs[j], o.subs[i]
}
func (o *Observer) Less(i, j int) bool {
	return o.prio[o.subs[i]] < o.prio[o.subs[j]]
}
