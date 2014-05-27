package benchmarks

import (
	"reflect"
	"testing"
)

type ReflectComponent interface{}

type TestReflectComponentA struct {
	ReflectComponent
	Value float64
}

type TestReflectComponentB struct {
	ReflectComponent
	Value string
}

type ReflectEntity struct {
	components map[reflect.Type]ReflectComponent
}

func NewReflectEntity() *ReflectEntity {
	return &ReflectEntity{
		components: map[reflect.Type]ReflectComponent{},
	}
}

func (e *ReflectEntity) Add(c ReflectComponent) {
	e.components[reflect.TypeOf(c)] = c
}

func (e *ReflectEntity) Get(c ReflectComponent) ReflectComponent {
	if r, ok := e.components[reflect.TypeOf(c)]; ok {
		return r
	}
	return nil
}

func (e *ReflectEntity) Scan(c ReflectComponent) {
	cvp := reflect.ValueOf(c)
	cv := reflect.Indirect(cvp)

	if r, ok := e.components[cv.Type()]; ok {
		cv.Set(reflect.ValueOf(r))
	}
	return
}

func BenchmarkReflectComponent(b *testing.B) {
	e := NewReflectEntity()
	ca := &TestReflectComponentA{Value: 1.23}
	cb := &TestReflectComponentB{Value: "a"}
	var va *TestReflectComponentA
	var vb *TestReflectComponentB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Add(ca)
		e.Add(cb)

		va = e.Get(va).(*TestReflectComponentA)
		vb = e.Get(vb).(*TestReflectComponentB)
	}

	b.Log(va, vb)
}

func BenchmarkReflectScanComponent(b *testing.B) {
	e := NewReflectEntity()
	ca := &TestReflectComponentA{Value: 1.23}
	cb := &TestReflectComponentB{Value: "a"}
	var va *TestReflectComponentA
	var vb *TestReflectComponentB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Add(ca)
		e.Add(cb)

		e.Scan(&va)
		e.Scan(&vb)
	}

	b.Log(va, vb)
}

type ConstComponentType int

const (
	ConstComponentTypeA ConstComponentType = iota
	ConstComponentTypeB
)

type ConstComponent interface {
	Type() ConstComponentType
}

type TestConstComponentA struct {
	ConstComponent
	Value float64
}

func (c TestConstComponentA) Type() ConstComponentType {
	return ConstComponentTypeA
}

type TestConstComponentB struct {
	ConstComponent
	Value string
}

func (c TestConstComponentB) Type() ConstComponentType {
	return ConstComponentTypeB
}

type ConstEntity struct {
	components map[ConstComponentType]ConstComponent
}

func NewConstEntity() *ConstEntity {
	return &ConstEntity{
		components: map[ConstComponentType]ConstComponent{},
	}
}

func (e *ConstEntity) Add(c ConstComponent) {
	e.components[c.Type()] = c
}

func (e *ConstEntity) Get(t ConstComponentType) ConstComponent {
	if r, ok := e.components[t]; ok {
		return r
	}
	return nil
}

func BenchmarkConstComponent(b *testing.B) {
	e := NewConstEntity()
	ca := &TestConstComponentA{Value: 1.23}
	cb := &TestConstComponentB{Value: "a"}
	var va *TestConstComponentA
	var vb *TestConstComponentB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Add(ca)
		e.Add(cb)

		va = e.Get(ConstComponentTypeA).(*TestConstComponentA)
		vb = e.Get(ConstComponentTypeB).(*TestConstComponentB)
	}

	b.Log(va, vb)
}
