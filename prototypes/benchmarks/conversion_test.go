package benchmarks

import (
	"testing"
)

type IdInt int
type IdStruct struct{ Id int }

func BenchmarkIdIntSet(b *testing.B) {
	var v IdInt
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v = IdInt(i)
	}
	_ = v
}

func BenchmarkIdStructSet(b *testing.B) {
	var v IdStruct
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v = IdStruct{i}
	}
	_ = v
}

func BenchmarkIdIntGet(b *testing.B) {
	var v IdInt
	var r int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v = IdInt(i)
		r = int(v)
	}
	_ = r
}

func BenchmarkIdStructGet(b *testing.B) {
	var v IdStruct
	var r int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v = IdStruct{i}
		r = v.Id
	}
	_ = r
}

type StructSlice struct{ data []int }

func (a StructSlice) Equals(b []int) bool {
	var found bool
	for _, t := range a.data {
		found = false
		for _, t2 := range b {
			if t == t2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestStructSlice(t *testing.T) {
	tests := []struct {
		Slice    StructSlice
		Equal    []int
		Expected bool
	}{
		{
			StructSlice{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9}},
			[]int{9, 8, 7, 6, 5, 4, 3, 2, 1},
			true,
		},
		{
			StructSlice{[]int{1, 2, 3}},
			[]int{9, 8, 7},
			false,
		},
	}

	for _, c := range tests {
		if r := c.Slice.Equals(c.Equal); r != c.Expected {
			t.Errorf("(%v).Equals(%v) != %v (got %v)", c.Slice, c.Equal, c.Expected, r)
		}
	}
}

func BenchmarkStructSlice(b *testing.B) {
	s := StructSlice{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9}}
	e := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}
	var r bool
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = s.Equals(e)
	}
	_ = r
}

type PointerSlice struct{ data []int }

func (a *PointerSlice) Equals(b []int) bool {
	var found bool
	for _, t := range a.data {
		found = false
		for _, t2 := range b {
			if t == t2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestPointerSlice(t *testing.T) {
	tests := []struct {
		Slice    *PointerSlice
		Equal    []int
		Expected bool
	}{
		{
			&PointerSlice{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9}},
			[]int{9, 8, 7, 6, 5, 4, 3, 2, 1},
			true,
		},
		{
			&PointerSlice{[]int{1, 2, 3}},
			[]int{9, 8, 7},
			false,
		},
	}

	for _, c := range tests {
		if r := c.Slice.Equals(c.Equal); r != c.Expected {
			t.Errorf("(%v).Equals(%v) != %v (got %v)", c.Slice, c.Equal, c.Expected, r)
		}
	}
}

func BenchmarkPointerSlice(b *testing.B) {
	s := &PointerSlice{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9}}
	e := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}
	var r bool
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = s.Equals(e)
	}
	_ = r
}

type NamedSlice []int

func (a NamedSlice) Equals(b []int) bool {
	var found bool
	for _, t := range a {
		found = false
		for _, t2 := range b {
			if t == t2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestNamedSlice(t *testing.T) {
	tests := []struct {
		Slice    NamedSlice
		Equal    []int
		Expected bool
	}{
		{
			NamedSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9}),
			[]int{9, 8, 7, 6, 5, 4, 3, 2, 1},
			true,
		},
		{
			NamedSlice([]int{1, 2, 3}),
			[]int{9, 8, 7},
			false,
		},
	}

	for _, c := range tests {
		if r := c.Slice.Equals(c.Equal); r != c.Expected {
			t.Errorf("(%v).Equals(%v) != %v (got %v)", c.Slice, c.Equal, c.Expected, r)
		}
	}
}

func BenchmarkNamedSlice(b *testing.B) {
	s := NamedSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})
	e := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}
	var r bool
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = s.Equals(e)
	}
	_ = r
}

type PNamedSlice []int

func (a *PNamedSlice) Equals(b []int) bool {
	var found bool
	for _, t := range *a {
		found = false
		for _, t2 := range b {
			if t == t2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestPNamedSlice(t *testing.T) {
	pn := func(d []int) *PNamedSlice {
		v := PNamedSlice(d)
		return &v
	}
	tests := []struct {
		Slice    *PNamedSlice
		Equal    []int
		Expected bool
	}{
		{
			pn([]int{1, 2, 3, 4, 5, 6, 7, 8, 9}),
			[]int{9, 8, 7, 6, 5, 4, 3, 2, 1},
			true,
		},
		{
			pn([]int{1, 2, 3}),
			[]int{9, 8, 7},
			false,
		},
	}

	for _, c := range tests {
		if r := c.Slice.Equals(c.Equal); r != c.Expected {
			t.Errorf("(%v).Equals(%v) != %v (got %v)", c.Slice, c.Equal, c.Expected, r)
		}
	}
}

func BenchmarkPNamedSlice(b *testing.B) {
	s := PNamedSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})
	p := &s
	e := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}
	var r bool
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = p.Equals(e)
	}
	_ = r
}
