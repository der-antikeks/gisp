package engine

import (
	"reflect"
	"testing"

	"github.com/der-antikeks/gisp/math"
)

func TestGroup(t *testing.T) {
	testObject_Position(NewGroup(), t)
	testObject_Up(NewGroup(), t)
	testObject_Rotation(NewGroup(), t)
	testObject_Scale(NewGroup(), t)

	testObject_Relationship(NewGroup(), t)
	if t.Failed() {
		t.Skip("Skip matrix tests until relationship tests succeed")
	}

	testObject_Matrix(NewGroup(), t)
}

func TestPerspectiveCamera(t *testing.T) {
	testObject_Position(NewPerspectiveCamera(45.0, 4.0/3.0, 0.1, 100.0), t)
	testObject_Up(NewPerspectiveCamera(45.0, 4.0/3.0, 0.1, 100.0), t)
	testObject_Rotation(NewPerspectiveCamera(45.0, 4.0/3.0, 0.1, 100.0), t)
	testObject_Scale(NewPerspectiveCamera(45.0, 4.0/3.0, 0.1, 100.0), t)

	testObject_Relationship(NewPerspectiveCamera(45.0, 4.0/3.0, 0.1, 100.0), t)
	if t.Failed() {
		t.Skip("Skip matrix tests until relationship tests succeed")
	}

	testObject_Matrix(NewPerspectiveCamera(45.0, 4.0/3.0, 0.1, 100.0), t)
}

func TestOrthographicCamera(t *testing.T) {
	testObject_Position(NewOrthographicCamera(-1, 1, 1, -1, 0, 1), t)
	testObject_Up(NewOrthographicCamera(-1, 1, 1, -1, 0, 1), t)
	testObject_Rotation(NewOrthographicCamera(-1, 1, 1, -1, 0, 1), t)
	testObject_Scale(NewOrthographicCamera(-1, 1, 1, -1, 0, 1), t)

	testObject_Relationship(NewOrthographicCamera(-1, 1, 1, -1, 0, 1), t)
	if t.Failed() {
		t.Skip("Skip matrix tests until relationship tests succeed")
	}

	testObject_Matrix(NewOrthographicCamera(-1, 1, 1, -1, 0, 1), t)
}

func TestMesh(t *testing.T) {
	testObject_Position(NewMesh(nil, nil), t)
	testObject_Up(NewMesh(nil, nil), t)
	testObject_Rotation(NewMesh(nil, nil), t)
	testObject_Scale(NewMesh(nil, nil), t)

	testObject_Relationship(NewMesh(nil, nil), t)
	if t.Failed() {
		t.Skip("Skip matrix tests until relationship tests succeed")
	}

	testObject_Matrix(NewMesh(nil, nil), t)
}

func TestScene(t *testing.T) {
	testObject_Position(NewScene(), t)
	testObject_Up(NewScene(), t)
	testObject_Rotation(NewScene(), t)
	testObject_Scale(NewScene(), t)

	testObject_Relationship(NewScene(), t)
	if t.Failed() {
		t.Skip("Skip matrix tests until relationship tests succeed")
	}

	testObject_Matrix(NewScene(), t)
}

func testObject_Position(o Object, t *testing.T) {
	if r := o.Position(); !r.Equals(math.Vector{}, 6) {
		t.Errorf("Initial position should equal %v (got %v)", math.Vector{}, r)
	}

	tests := []math.Vector{
		math.Vector{1, 2, 3, 4},
		math.Vector{},
		math.Vector{-5, 2.3456, 0},
	}

	for _, c := range tests {
		o.SetPosition(c)

		if r := o.Position(); !r.Equals(c, 6) {
			t.Errorf("Position should equal %v (got %v)", c, r)
		}
	}
}

func testObject_Up(o Object, t *testing.T) {
	if r := o.Up(); !r.Equals(math.Vector{0, 1, 0}, 6) {
		t.Errorf("Initial up should equal %v (got %v)", math.Vector{0, 1, 0}, r)
	}

	tests := []math.Vector{
		math.Vector{2, 0, 0},
		math.Vector{},
		math.Vector{1, 0, 1},
	}

	for _, c := range tests {
		o.SetUp(c)

		if r := o.Up(); !r.Equals(c.Normalize(), 6) {
			t.Errorf("Up should equal %v (got %v)", c.Normalize(), r)
		}
	}
}

func testObject_Rotation(o Object, t *testing.T) {
	if r := o.Rotation(); !r.Equals(math.Quaternion{}, 6) {
		t.Errorf("Initial rotation should equal %v (got %v)", math.Quaternion{}, r)
	}

	q1 := math.QuaternionFromAxisAngle(math.Vector{1, 0, 0}, 12.46)
	q3 := math.QuaternionFromAxisAngle(math.Vector{0, 0, 1}, 45.8)

	tests := []math.Quaternion{
		q1,
		math.Quaternion{},
		q3,
	}

	for _, c := range tests {
		o.SetRotation(c)

		if r := o.Rotation(); !r.Equals(c, 6) {
			t.Errorf("Rotation should equal %v (got %v)", c, r)
		}
	}
}

func testObject_Scale(o Object, t *testing.T) {
	if r := o.Scale(); !r.Equals(math.Vector{1, 1, 1}, 6) {
		t.Errorf("Initial scale should equal %v (got %v)", math.Vector{1, 1, 1}, r)
	}

	tests := []math.Vector{
		math.Vector{1, 2, 1},
		math.Vector{},
		math.Vector{2, 1, 2},
	}

	for _, c := range tests {
		o.SetScale(c)

		if r := o.Scale(); !r.Equals(c, 6) {
			t.Errorf("Scale should equal %v (got %v)", c, r)
		}
	}
}

func testObject_Relationship(o Object, t *testing.T) {
	if r := o.Parent(); r != nil {
		t.Errorf("Initial parent should be nil (got %p)", r)
	}

	if r := o.Children(); r != nil {
		t.Errorf("Initial children should be empty (got %v)", r)
	}

	vt := reflect.ValueOf(o).Elem().Type()
	a := reflect.New(vt).Interface().(Object)
	b := reflect.New(vt).Interface().(Object)
	c := reflect.New(vt).Interface().(Object)

	// o[a[]]
	o.AddChild(a)

	if r := o.Children(); r == nil || len(r) != 1 || r[0] != a {
		t.Errorf("First and only child should be %p (got %v)", a, r)
	}

	if r := a.Parent(); r != o {
		t.Errorf("Parent should be %p (got %p)", o, r)
	}

	// o[a[], b[], c[]]
	o.AddChild(b, c)

	if r := o.Children(); r == nil || len(r) != 3 || r[0] != a || r[1] != b || r[2] != c {
		t.Errorf("Children should be %p, %p and %p (got %v)", b, c, r)
	}

	if r := c.Parent(); r != o {
		t.Errorf("Parent should be %p (got %p)", o, r)
	}

	// o[a[], b[]]
	o.RemoveChild(c)

	if r := o.Children(); r == nil || len(r) != 2 || r[0] != a || r[1] != b {
		t.Errorf("Children should be %p and %p (got %v)", a, b, r)
	}

	if r := c.Parent(); r != nil {
		t.Errorf("Parent should be nil (got %p)", r)
	}

	// o[a[], b[c]]
	b.AddChild(c)

	if r := o.Children(); r == nil || len(r) != 2 || r[0] != a || r[1] != b {
		t.Errorf("Children should be %p and %p (got %v)", a, b, r)
	} else {
		if r := r[1].Children(); len(r) != 1 || r[0] != c {
			t.Errorf("Child should be %p (got %v)", c, r)
		}
	}

	if r := c.Parent(); r != b {
		t.Errorf("Parent should be %p (got %p)", b, r)
	}
}

func testObject_Matrix(o Object, t *testing.T) {
	// setup test relationships
	// o[a[], b[c]]
	vt := reflect.ValueOf(o).Elem().Type()
	a := reflect.New(vt).Interface().(Object)
	b := reflect.New(vt).Interface().(Object)
	c := reflect.New(vt).Interface().(Object)

	o.AddChild(a, b)
	b.AddChild(c)

	if r := b.MatrixWorld(); !r.Equals(math.Matrix{}, 6) {
		t.Errorf("Initial world matrix should be \n%v (got \n%v)", math.Matrix{}, r)
	}

	a.SetScale(math.Vector{1, 1, 1})
	b.SetScale(math.Vector{1, 1, 1})
	c.SetScale(math.Vector{1, 1, 1})

	if r := b.Matrix(); !r.Equals(math.Identity(), 6) {
		t.Errorf("Initial matrix should be \n%v (got \n%v)", math.Identity(), r)
	}

	o.SetPosition(math.Vector{0, -10, 0})
	a.SetPosition(math.Vector{0, 0, 10})
	b.SetPosition(math.Vector{10, 0, 0})

	o.UpdateMatrixWorld(false)

	/*
		o y-10
			a z+10
			b x+10
				c x+10, y-10, z+0
	*/

	expected := math.Identity()
	expected.SetPosition(math.Vector{10, -10, 0})

	if r := c.MatrixWorld(); !r.Equals(expected, 6) {
		t.Errorf("Translated world matrix should be \n%v (got \n%v)", expected, r)
	}

	/*
		o y90°
			a
			b x45°
				c x45° y90° z0°
	*/

	q1 := math.QuaternionFromAxisAngle(math.Vector{0, 1, 0}, 90*math.DEG2RAD)
	o.SetRotation(q1)
	q2 := math.QuaternionFromAxisAngle(math.Vector{1, 0, 0}, 45*math.DEG2RAD)
	b.SetRotation(q2)
	o.UpdateMatrixWorld(false)

	expected = q1.RotationMatrix()
	expected.SetPosition(math.Vector{0, -10, 0})
	m := q2.RotationMatrix()
	m.SetPosition(math.Vector{10, 0, 0})
	expected = expected.Mul(m)

	if r := c.MatrixWorld(); !r.Equals(expected, 6) {
		t.Errorf("Rotated world matrix should be \n%v (got \n%v)", expected, r)
	}
}
