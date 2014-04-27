package benchmarks

// non blocking and thread safe iterators

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().Unix())
}

type Key interface{}
type Value interface{}

type List interface {
	Add(Value)
	Get() Value
	Remove(Value)
	Iterate(ListIterator)
}

type ListIterator func(Value) bool // return true if iteration should be stopped

type Hash interface {
	Add(Key, Value)
	Get(Key) Value
	Remove(Key)
	Iterate(HashIterator)
}

type HashIterator func(Key, Value) bool

func testList(t *testing.T, l List) {
	// test add and get
	v := 123
	l.Add(v)
	r := l.Get()
	if r != v {
		t.Fatalf("%v and %v do not match", v, r)
	}

	// test order of add
	v2 := 3
	l.Add(v2)
	r = l.Get()
	if r != v {
		t.Fatalf("%v and %v do not match", v, r)
	}

	// test remove
	l.Remove(v)
	r = l.Get()
	if r != v2 {
		t.Fatalf("%v and %v do not match", v2, r)
	}

	// test full iteration
	l.Add(4)
	l.Add(5)

	sum, cnt := 0, 0
	l.Iterate(func(i Value) bool {
		cnt++
		sum += i.(int)
		return false
	})

	if cnt != 3 {
		t.Fatalf("wrong length: %v", cnt)
	}

	if sum != 12 {
		t.Fatalf("wrong sum: %v", sum)
	}

	// test stopped iteration
	sum, cnt = 0, 0
	l.Iterate(func(i Value) bool {
		cnt++
		sum += i.(int)
		return cnt == 2
	})

	if cnt != 2 || sum != 7 {
		t.Fatalf("iteration did not stop: %v", cnt)
	}
}

func testConcurrentList(t *testing.T, l List) {
	var wg sync.WaitGroup
	n := 100

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			l.Add(n)
			sum, cnt := 0, 0
			l.Iterate(func(i Value) bool {
				cnt++
				sum += i.(int)
				return false
			})
		}(i)
	}
	wg.Wait()

	sum, cnt := 0, 0
	l.Iterate(func(i Value) bool {
		cnt++
		sum += i.(int)
		return false
	})

	if cnt != n {
		t.Fatalf("wrong length: %v", cnt)
	}

	if sum != (n-1)*n/2 {
		t.Fatalf("wrong sum: %v", sum)
	}
}

func benchmarkListAdd(b *testing.B, l List) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Add(rand.Int())
	}
}

func benchmarkListRemove(b *testing.B, l List) {
	for i := 0; i < b.N; i++ {
		l.Add(rand.Int())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Remove(rand.Int())
	}
}

func benchmarkListIterate(b *testing.B, l List) {
	for i := 0; i < b.N; i++ {
		l.Add(rand.Int())
	}

	cnt := 0
	f := func(i Value) bool {
		cnt++
		return false
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Iterate(f)
	}
}

func benchmarkListMixed(b *testing.B, l List) {
	f := func(i Value) bool {
		return false
	}
	var c bool
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c = rand.Intn(2) == 0
		if c {
			wg.Add(1)
		}

		switch p := rand.Intn(3); p {
		case 0:
			if c {
				go func() {
					defer wg.Done()
					l.Add(rand.Int())
				}()
			} else {
				l.Add(rand.Int())
			}
		case 1:
			if c {
				go func() {
					defer wg.Done()
					l.Remove(rand.Int())
				}()
			} else {
				l.Remove(rand.Int())
			}
		case 2:
			if c {
				go func() {
					defer wg.Done()
					l.Iterate(f)
				}()
			} else {
				l.Iterate(f)
			}
		}
	}

	wg.Wait()
}

type SliceList struct {
	data []Value
	sync.RWMutex
}

func NewSliceList() *SliceList {
	return &SliceList{data: []Value{}}
}
func (l *SliceList) Add(v Value) {
	l.Lock()
	defer l.Unlock()
	l.data = append(l.data, v)
}
func (l *SliceList) Get() Value {
	l.RLock()
	defer l.RUnlock()
	return l.data[0]
}
func (l *SliceList) Remove(v Value) {
	l.Lock()
	defer l.Unlock()
	for i, d := range l.data {
		if d == v {
			copy(l.data[i:], l.data[i+1:])
			l.data[len(l.data)-1] = nil
			l.data = l.data[:len(l.data)-1]
		}
	}
}
func (l *SliceList) Iterate(f ListIterator) {
	l.RLock()
	defer l.RUnlock()
	for _, d := range l.data {
		if f(d) {
			return
		}
	}
}

func TestSliceList(t *testing.T) {
	testList(t, NewSliceList())
	testConcurrentList(t, NewSliceList())
}

func BenchmarkSliceListAdd(b *testing.B) {
	benchmarkListAdd(b, NewSliceList())
}

func BenchmarkSliceListRemove(b *testing.B) {
	benchmarkListRemove(b, NewSliceList())
}

func BenchmarkSliceListIterate(b *testing.B) {
	benchmarkListIterate(b, NewSliceList())
}

func BenchmarkSliceListMixed(b *testing.B) {
	benchmarkListMixed(b, NewSliceList())
}

type CopySliceList struct {
	data []Value
	sync.RWMutex
}

func NewCopySliceList() *CopySliceList {
	return &CopySliceList{data: []Value{}}
}
func (l *CopySliceList) Add(v Value) {
	l.Lock()
	defer l.Unlock()
	l.data = append(l.data, v)
}
func (l *CopySliceList) Get() Value {
	l.RLock()
	defer l.RUnlock()
	return l.data[0]
}
func (l *CopySliceList) Remove(v Value) {
	l.Lock()
	defer l.Unlock()
	for i, d := range l.data {
		if d == v {
			copy(l.data[i:], l.data[i+1:])
			l.data[len(l.data)-1] = nil
			l.data = l.data[:len(l.data)-1]
		}
	}
}
func (l *CopySliceList) Iterate(f ListIterator) {
	l.RLock()
	it := make([]Value, len(l.data))
	copy(it, l.data)
	l.RUnlock()

	for _, d := range it {
		if f(d) {
			return
		}
	}
}

func TestCopySliceList(t *testing.T) {
	testList(t, NewCopySliceList())
	testConcurrentList(t, NewCopySliceList())
}

func BenchmarkCopySliceListAdd(b *testing.B) {
	benchmarkListAdd(b, NewCopySliceList())
}

func BenchmarkCopySliceListRemove(b *testing.B) {
	benchmarkListRemove(b, NewCopySliceList())
}

func BenchmarkCopySliceListIterate(b *testing.B) {
	benchmarkListIterate(b, NewCopySliceList())
}

func BenchmarkCopySliceListMixed(b *testing.B) {
	benchmarkListMixed(b, NewCopySliceList())
}
