package benchmarks

import (
	"math/rand"
	"testing"
)

func BenchmarkBasicMap_Append(b *testing.B) {
	//b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := b.N
		if l > 1000000000 {
			l = 1000000000
		}
		m := map[int]bool{}

		for n := 0; n < l; n++ {
			m[n] = true
		}
	}
}

func BenchmarkBasicSlice_Append(b *testing.B) {
	//b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := b.N
		if l > 1000000000 {
			l = 1000000000
		}
		s := []int{}

		for n := 0; n < l; n++ {
			s = append(s, n)
		}
	}
}

func BenchmarkBasicMap_Search(b *testing.B) {
	l := 100000
	m := map[int]bool{}
	for n := 0; n < l; n++ {
		m[n] = true
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := rand.Intn(l - 1)
		_ = m[f]
	}
}

func BenchmarkBasicSlice_Search(b *testing.B) {
	l := 100000
	s := []int{}
	for n := 0; n < l; n++ {
		s = append(s, n)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := rand.Intn(l - 1)
		for _, v := range s {
			if v == f {
				break
			}
		}
	}
}
