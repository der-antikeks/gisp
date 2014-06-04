package benchmarks

import (
	"math/rand"
	"testing"
)

const NumItems int = 1e6 // 1000000

var int_data []int = make([]int, NumItems)

func init() {
	for i := 0; i < NumItems; i++ {
		int_data[i] = rand.Int()
	}
}

func BenchmarkIntSliceIterator(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sum int = 0
		for _, val := range int_data {
			sum += val
		}
	}
}

func IntCallbackIterator(cb func(int)) {
	for _, val := range int_data {
		cb(val)
	}
}

func BenchmarkIntCallbackIterator(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sum int = 0
		b.StopTimer()
		cb := func(val int) {
			sum += val
		}
		b.StartTimer()
		IntCallbackIterator(cb)
	}
}

type StatefulIterator interface {
	Value() int
	Next() bool
}

type intStatefulIterator struct {
	current int
	data    []int
}

func (it *intStatefulIterator) Value() int {
	return it.data[it.current]
}

func (it *intStatefulIterator) Next() bool {
	it.current++
	if it.current >= len(it.data) {
		return false
	}
	return true
}

func NewIntStatefulIterator(data []int) *intStatefulIterator {
	return &intStatefulIterator{data: data, current: -1}
}

func BenchmarkIntStatefulIterator(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sum int = 0
		b.StopTimer()
		sit := NewIntStatefulIterator(int_data)
		b.StartTimer()
		for it := sit; it.Next(); {
			sum += it.Value()
		}
	}
}

func IntChannelIterator() <-chan int {
	ch := make(chan int)
	go func() {
		for _, val := range int_data {
			ch <- val
		}
		close(ch)
	}()
	return ch
}

func BenchmarkIntsChannelIterator(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sum int = 0
		b.StopTimer()
		ichan := IntChannelIterator()
		b.StartTimer()
		for val := range ichan {
			sum += val
		}
	}
}

func IntBufChannelIterator() <-chan int {
	ch := make(chan int, 50)
	go func() {
		for _, val := range int_data {
			ch <- val
		}
		close(ch)
	}()
	return ch
}

func BenchmarkIntsBufChannelIterator(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sum int = 0
		b.StopTimer()
		ichan := IntBufChannelIterator()
		b.StartTimer()
		for val := range ichan {
			sum += val
		}
	}
}
