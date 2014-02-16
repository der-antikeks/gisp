package benchmarks

import (
	"fmt"
	"reflect"
	"testing"
)

type Component interface{}

type TestComponentA struct {
	Component
	Value float64
}

type TestComponentB struct {
	Component
	Value string
}

type TestComponentC struct {
	Component
	Value []float64
}

type TestComponentD struct {
	Component
	Value map[string]float64
}

type TCE struct{ Component }
type TCF struct{ Component }
type TCG struct{ Component }
type TCH struct{ Component }

type Collection interface {
	Add(Component)
	Remove(Component)
	Get(Component) Component
}

func testAll(t *testing.T, c Collection) {
	testAdd(t, c)
	if !t.Failed() {
		testGet(t, c)
	}
	if !t.Failed() {
		testRemove(t, c)
	}
	if !t.Failed() {
		testPointer(t, c)
	}
}

func testAdd(t *testing.T, c Collection) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("add failed: %v", err)
		}
	}()

	for i := 0; i <= 5; i++ {
		c.Add(&TestComponentA{Value: float64(i) * 1.23})
	}

	for i := 0; i <= 5; i++ {
		c.Add(&TestComponentB{Value: fmt.Sprintf("%#v", i)})
	}

	for i := 0; i <= 5; i++ {
		v := make([]float64, 5)
		for j := range v {
			v[j] = float64(i*j+j) * 1.23
		}
		c.Add(&TestComponentC{Value: v})
	}

	for i := 0; i <= 5; i++ {
		v := make(map[string]float64)
		for j := 0; j <= 5; j++ {
			v[fmt.Sprintf("%#v", j)] = float64(i*j+j) * 1.23
		}
		c.Add(&TestComponentD{Value: v})
	}
}

func testGet(t *testing.T, c Collection) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("get failed: %v", err)
		}
	}()

	{ // get pointer
		in := &TestComponentA{}
		out := TestComponentA{Value: float64(5) * 1.23}
		o := c.Get(in).(*TestComponentA)
		if *o != out {
			t.Errorf("Get(%v) => %v, want %v", in, o, out)
		}
	}

	{ // get value
		in := TestComponentA{}
		out := TestComponentA{Value: float64(5) * 1.23}
		o := c.Get(in).(*TestComponentA)
		if *o != out {
			t.Errorf("Get(%v) => %v, want %v", in, o, out)
		}
	}

	{ // get pointer
		in := &TestComponentB{}
		out := TestComponentB{Value: fmt.Sprintf("%#v", 5)}
		o := c.Get(in).(*TestComponentB)
		if *o != out {
			t.Errorf("Get(%v) => %v, want %v", in, o, out)
		}
	}

	{ // get value
		in := TestComponentB{}
		out := TestComponentB{Value: fmt.Sprintf("%#v", 5)}
		o := c.Get(in).(*TestComponentB)
		if *o != out {
			t.Errorf("Get(%v) => %v, want %v", in, o, out)
		}
	}
}

func testRemove(t *testing.T, c Collection) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("remove failed: %v", err)
		}
	}()

	{ // remove pointer
		in := &TestComponentA{}
		c.Remove(in)
		o := c.Get(in)
		if o != nil {
			t.Errorf("Remove(%v) => %v, want %v", in, o, nil)
		}
	}

	{ // remove value
		in := TestComponentB{}
		c.Remove(in)
		o := c.Get(in)
		if o != nil {
			t.Errorf("Remove(%v) => %v, want %v", in, o, nil)
		}
	}
}

func testPointer(t *testing.T, c Collection) {
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("remove failed: %v", err)
		}
	}()

	in := &TestComponentA{Value: 4.89}
	c.Add(in)

	k := &TestComponentA{Value: 1.23}
	o := c.Get(k)

	if in != o {
		t.Errorf("Get(%v) => %p, want %p", k, o, in)
	}
}

func TestPointerComparison(t *testing.T) {
	a := &TestComponentA{Value: 1.23}
	b := &TestComponentA{Value: 1.23}
	c := a
	m := make(map[string]*TestComponentA)
	m["a"] = a
	d := m["a"]

	if a == b {
		t.Errorf("%v == %v", a, b)
	}

	if *a != *b {
		t.Errorf("%v != %v", *a, *b)
	}

	if a != c {
		t.Errorf("%v != %v", a, c)
	}

	if a != d {
		t.Errorf("%v != %v", a, d)
	}

	if a != m["a"] {
		t.Errorf("%v != %v", a, m["a"])
	}
}

func benchAll(b *testing.B, c Collection) {
	benchAdd(b, c)
	benchGet(b, c)
	benchRemove(b, c)
}

func benchAdd(b *testing.B, c Collection) {
	// create components
	ta := &TestComponentA{Value: 1.23}
	tb := &TestComponentB{Value: "b"}

	sv := make([]float64, 5)
	for j := range sv {
		sv[j] = float64(j) * 1.23
	}
	tc := &TestComponentC{Value: sv}

	mv := make(map[string]float64)
	for j := 0; j <= 5; j++ {
		mv[fmt.Sprintf("%#v", j)] = float64(j) * 1.23
	}
	td := &TestComponentD{Value: mv}

	// benchmark add
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Add(ta)
		c.Add(tb)
		c.Add(tc)
		c.Add(td)
	}
}

func benchGet(b *testing.B, c Collection) {
	// add components and create keys
	c.Add(&TestComponentA{Value: 1.23})
	ka := TestComponentA{}

	c.Add(&TestComponentB{Value: "b"})
	kb := TestComponentB{}

	sv := make([]float64, 5)
	for j := range sv {
		sv[j] = float64(j) * 1.23
	}
	c.Add(&TestComponentC{Value: sv})
	kc := TestComponentC{}

	mv := make(map[string]float64)
	for j := 0; j <= 5; j++ {
		mv[fmt.Sprintf("%#v", j)] = float64(j) * 1.23
	}
	c.Add(&TestComponentD{Value: mv})
	kd := TestComponentD{}

	// add placeholders
	c.Add(&TCE{})
	c.Add(&TCF{})
	c.Add(&TCG{})
	c.Add(&TCH{})

	// benchmark get
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(kd)
		c.Get(kc)
		c.Get(kb)
		c.Get(ka)
	}
}

func benchRemove(b *testing.B, c Collection) {
	// create components and keys
	ta := &TestComponentA{Value: 1.23}
	ka := TestComponentA{}

	tb := &TestComponentB{Value: "b"}
	kb := TestComponentB{}

	sv := make([]float64, 5)
	for j := range sv {
		sv[j] = float64(j) * 1.23
	}
	tc := &TestComponentC{Value: sv}
	kc := TestComponentC{}

	mv := make(map[string]float64)
	for j := 0; j <= 5; j++ {
		mv[fmt.Sprintf("%#v", j)] = float64(j) * 1.23
	}
	td := &TestComponentD{Value: mv}
	kd := TestComponentD{}

	// add placeholders
	c.Add(&TCE{})
	c.Add(&TCF{})
	c.Add(&TCG{})
	c.Add(&TCH{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// add comüonents
		b.StopTimer()
		c.Add(ta)
		c.Add(tb)
		c.Add(tc)
		c.Add(td)

		// benchmark remove
		b.StartTimer()
		c.Remove(kd)
		c.Remove(kc)
		c.Remove(kb)
		c.Remove(ka)
	}
}

// slice collection

type SliceCollection struct {
	components []Component
}

func NewSliceCollection() *SliceCollection {
	return &SliceCollection{
		components: []Component{},
	}
}

func (c *SliceCollection) key(component Component) reflect.Type {
	k := reflect.TypeOf(component)
	if k.Kind() == reflect.Ptr {
		k = k.Elem()
	}

	return k
}

func (c *SliceCollection) Add(component Component) {
	k := c.key(component)
	found := false

	for i, o := range c.components {
		if c.key(o) == k {
			found = true
			c.components[i] = component
		}
	}

	if !found {
		c.components = append(c.components, component)
	}
}

func (c *SliceCollection) Remove(component Component) {
	k := c.key(component)
	for i, o := range c.components {
		if c.key(o) == k {
			copy(c.components[i:], c.components[i+1:])
			c.components[len(c.components)-1] = nil
			c.components = c.components[:len(c.components)-1]

			return
		}
	}
}

func (c *SliceCollection) Get(component Component) Component {
	k := c.key(component)
	for _, o := range c.components {
		if c.key(o) == k {
			return o
		}
	}

	return nil
}

func TestSliceCollection(t *testing.T) {
	testAll(t, NewSliceCollection())
}

func BenchmarkSliceCollection_Add(b *testing.B) {
	benchAdd(b, NewSliceCollection())
}

func BenchmarkSliceCollection_Get(b *testing.B) {
	benchGet(b, NewSliceCollection())
}

func BenchmarkSliceCollection_Remove(b *testing.B) {
	benchRemove(b, NewSliceCollection())
}

// reflect map collection

type ReflectMapCollection struct {
	components map[reflect.Type]Component
}

func NewReflectMapCollection() *ReflectMapCollection {
	return &ReflectMapCollection{
		components: map[reflect.Type]Component{},
	}
}

func (c *ReflectMapCollection) key(component Component) reflect.Type {
	k := reflect.TypeOf(component)
	if k.Kind() == reflect.Ptr {
		k = k.Elem()
	}

	return k
}

func (c *ReflectMapCollection) Add(component Component) {
	k := c.key(component)
	c.components[k] = component
}

func (c *ReflectMapCollection) Remove(component Component) {
	k := c.key(component)
	delete(c.components, k)
}

func (c *ReflectMapCollection) Get(component Component) Component {
	k := c.key(component)
	if r, ok := c.components[k]; ok {
		return r
	}
	return nil
}
func TestReflectMapCollection(t *testing.T) {
	testAll(t, NewReflectMapCollection())
}

func BenchmarkReflectMapCollection_Add(b *testing.B) {
	benchAdd(b, NewReflectMapCollection())
}

func BenchmarkReflectMapCollection_Get(b *testing.B) {
	benchGet(b, NewReflectMapCollection())
}

func BenchmarkReflectMapCollection_Remove(b *testing.B) {
	benchRemove(b, NewReflectMapCollection())
}

// string map collection

type StringMapCollection struct {
	components map[string]Component
}

func NewStringMapCollection() *StringMapCollection {
	return &StringMapCollection{
		components: map[string]Component{},
	}
}

func (c *StringMapCollection) key(component Component) string {
	k := fmt.Sprintf("%T", component)
	if k[0] == '*' {
		k = k[1:]
	}
	return k
}

func (c *StringMapCollection) Add(component Component) {
	k := c.key(component)
	c.components[k] = component
}

func (c *StringMapCollection) Remove(component Component) {
	k := c.key(component)
	delete(c.components, k)
}

func (c *StringMapCollection) Get(component Component) Component {
	k := c.key(component)
	if r, ok := c.components[k]; ok {
		return r
	}
	return nil
}

func TestStringMapCollection(t *testing.T) {
	testAll(t, NewStringMapCollection())
}

func BenchmarkStringMapCollection_Add(b *testing.B) {
	benchAdd(b, NewStringMapCollection())
}

func BenchmarkStringMapCollection_Get(b *testing.B) {
	benchGet(b, NewStringMapCollection())
}

func BenchmarkStringMapCollection_Remove(b *testing.B) {
	benchRemove(b, NewStringMapCollection())
}
