package benchmarks

import (
	"sync"
	"testing"
)

func BenchmarkLocking_Mutex(b *testing.B) {
	var l sync.Mutex
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Lock()
		l.Unlock()
	}
}

func BenchmarkLocking_RWMutex(b *testing.B) {
	var l sync.RWMutex
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Lock()
		l.Unlock()
	}
}

func BenchmarkLocking_ChanInt(b *testing.B) {
	l := make(chan int, 1)
	l <- 1
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-l
		l <- 1
	}
}

func BenchmarkLocking_ChanStruct(b *testing.B) {
	l := make(chan struct{}, 1)
	l <- struct{}{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-l
		l <- struct{}{}
	}
}
