package benchmarks

import (
	"fmt"
	"sync"
	"testing"
)

type FuncObserver []func(msg string)

func (obs *FuncObserver) Subscribe(f func(msg string)) {
	*obs = append(*obs, f)
}

func (obs *FuncObserver) Publish(msg string) {
	for _, o := range *obs {
		o(msg)
	}
}

func TestFuncObserver(t *testing.T) {
	obs := new(FuncObserver)
	var wg sync.WaitGroup

	obs.Subscribe(func(msg string) {
		wg.Done()
	})

	obs.Subscribe(func(msg string) {
		wg.Done()
	})

	for i := 0; i < 10; i++ {
		wg.Add(2)
		obs.Publish(fmt.Sprintf("Entity %v", i))
	}

	wg.Wait()
}

func BenchmarkFuncObserver(b *testing.B) {
	obs := new(FuncObserver)
	var wg sync.WaitGroup

	obs.Subscribe(func(msg string) {
		wg.Done()
	})

	obs.Subscribe(func(msg string) {
		wg.Done()
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(2)
		obs.Publish(fmt.Sprintf("Entity %v", i))
	}

	wg.Wait()
}

type ChanObserver []chan string

func (obs *ChanObserver) Subscribe(c chan string) {
	*obs = append(*obs, c)
}

func (obs *ChanObserver) Publish(msg string) {
	for _, c := range *obs {
		c <- msg
	}
}

func TestChanObserver(t *testing.T) {
	obs := new(ChanObserver)
	var wg sync.WaitGroup

	c1 := make(chan string)
	obs.Subscribe(c1)
	go func() {
		for _ = range c1 {
			wg.Done()
		}
	}()

	c2 := make(chan string)
	obs.Subscribe(c2)
	go func() {
		for _ = range c2 {
			wg.Done()
		}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			wg.Add(2)
			obs.Publish(fmt.Sprintf("Entity %v", i))
		}
	}()

	wg.Wait()
}

func BenchmarkChanObserver(b *testing.B) {
	obs := new(ChanObserver)
	var wg sync.WaitGroup

	c1 := make(chan string)
	obs.Subscribe(c1)
	go func() {
		for _ = range c1 {
			wg.Done()
		}
	}()

	c2 := make(chan string)
	obs.Subscribe(c2)
	go func() {
		for _ = range c2 {
			wg.Done()
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(2)
		obs.Publish(fmt.Sprintf("Entity %v", i))
	}

	wg.Wait()
}
