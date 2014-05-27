package benchmarks

import (
	"sync"
	"testing"
	"time"
)

// iterator for uniqe items
// concurrent read/write possible
type Iterator struct {
	sync.RWMutex
	items []int
}

func NewIterator() *Iterator {
	return &Iterator{items: []int{}}
}

func (i *Iterator) Add(item int) (done chan struct{}) {
	done = make(chan struct{})
	go func() {
		i.Lock()
		defer close(done)
		defer i.Unlock()

		for _, e := range i.items {
			if e == item {
				return
			}
		}
		i.items = append(i.items, item)
	}()
	return done
}

func (i *Iterator) Remove(item int) (done chan struct{}) {
	done = make(chan struct{})
	go func() {
		i.Lock()
		defer close(done)
		defer i.Unlock()

		for p, e := range i.items {
			if e == item {
				copy(i.items[p:], i.items[p+1:])
				i.items[len(i.items)-1] = 0
				i.items = i.items[:len(i.items)-1]

				return
			}
		}
	}()
	return done
}

func (i *Iterator) Iterate(done <-chan struct{}) <-chan int {
	out := make(chan int)
	go func() {
		i.RLock()
		defer close(out)
		defer i.RUnlock()

		for _, e := range i.items {
			select {
			case out <- e:
			case <-done:
				return
			}
		}
	}()
	return out
}

func TestConcurrentIterator_Add(t *testing.T) {
	it := NewIterator()
	<-it.Add(1)
	<-it.Add(2)
	<-it.Add(3)

	done := make(chan struct{})
	c := it.Iterate(done)

	if r := <-c; r != 1 {
		t.Errorf("received %v instead of 1", r)
	}

	if r := <-c; r != 2 {
		t.Errorf("received %v instead of 2", r)
	}

	close(done)
	time.Sleep(1 * time.Millisecond)
	if r, ok := <-c; r != 0 || ok {
		t.Errorf("received %v, %v instead of 0, false", r, ok)
	}
}

func TestConcurrentIterator_Remove(t *testing.T) {
	it := NewIterator()
	<-it.Add(1)
	<-it.Add(2)
	<-it.Add(3)

	<-it.Remove(2)

	done := make(chan struct{})
	c := it.Iterate(done)

	if r := <-c; r != 1 {
		t.Errorf("received %v instead of 1", r)
	}

	if r := <-c; r != 3 {
		t.Errorf("received %v instead of 3", r)
	}

	if r, ok := <-c; r != 0 || ok {
		t.Errorf("received %v, %v instead of 0, false", r, ok)
	}

	close(done)

	if r, ok := <-c; r != 0 || ok {
		t.Errorf("received %v, %v instead of 0, false", r, ok)
	}
}

func TestConcurrentIterator_Iterate(t *testing.T) {
	it := NewIterator()
	n := 10
	sum := 0

	for i := 0; i < n; i++ {
		<-it.Add(i)
		sum += i
	}

	done := make(chan struct{})
	for c := range it.Iterate(done) {
		sum -= c
	}

	if sum != 0 {
		t.Errorf("sum of %v instead of 0", sum)
	}
}

func TestConcurrentIterator_Lock(t *testing.T) {
	it := NewIterator()
	<-it.Add(1)
	<-it.Add(2)

	done := make(chan struct{})
	c := it.Iterate(done)

	if r := <-c; r != 1 {
		t.Errorf("received %v instead of 1", r)
	}

	it.Add(3)
	it.Remove(2)

	if r := <-c; r != 2 {
		t.Errorf("received %v instead of 2", r)
	}

	if r, ok := <-c; r != 0 || ok {
		t.Errorf("received %v, %v instead of 0, false", r, ok)
	}

	close(done)
	time.Sleep(1 * time.Millisecond)

	done = make(chan struct{})
	c = it.Iterate(done)

	if r := <-c; r != 1 {
		t.Errorf("received %v instead of 1", r)
	}

	if r := <-c; r != 3 {
		t.Errorf("received %v instead of 3", r)
	}

	if r, ok := <-c; r != 0 || ok {
		t.Errorf("received %v, %v instead of 0, false", r, ok)
	}
}

func BenchmarkConcurrentIterator(b *testing.B) {
	it := NewIterator()
	sum := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := (i % 100) + 1
		switch i % 3 {
		case 0: // add
			it.Add(n)
		case 1: // remove
			it.Remove(n)
		case 2: // iterate
			for e := range it.Iterate(nil) {
				sum += e
			}
		}
	}

	b.Log(b.N, "->", sum)
}

func BenchmarkSliceIterator(b *testing.B) {
	items := []int{}
	sum := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := (i % 100) + 1
		switch i % 3 {
		case 0: // add
			for _, e := range items {
				if e == n {
					continue
				}
			}
			items = append(items, n)

		case 1: // remove
			for p, e := range items {
				if e == n {
					copy(items[p:], items[p+1:])
					items[len(items)-1] = 0
					items = items[:len(items)-1]

					continue
				}
			}

		case 2: // iterate
			for _, e := range items {
				sum += e
			}
		}
	}

	b.Log(b.N, "->", sum)
}
