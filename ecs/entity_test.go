package ecs

import (
	"testing"
)

func TestEntityState(t *testing.T) {
	tca := &TestAComponent{Value: 1} // TestAType
	tcb := &TestBComponent{Value: 2} // TestBType
	tcc := &TestCComponent{Value: 3} // TestCType

	en := NewEntity("Test Entity", tca)
	en.State("testing b").Add(tcb)
	en.State("testing c").Add(tcc)

	if r := en.Get(TestAType); r != tca {
		t.Errorf("get default state component returned %p instead of %p", r, tca)
	}

	if r := en.Get(TestBType); r != nil {
		t.Errorf("get stateful component returned %p instead of nil at default state", r)
	}

	en.ChangeState("testing b")

	if r := en.Get(TestAType); r != tca {
		t.Errorf("get default state component at specified state returned %p instead of %p", r, tca)
	}

	if r := en.Get(TestBType); r != tcb {
		t.Errorf("get stateful component returned %p instead of %p", r, tcb)
	}

	if r := en.Get(TestCType); r != nil {
		t.Errorf("get stateful component at wrong state returned %p instead of nil", r)
	}

	en.ChangeState("testing c")

	if r := en.Get(TestAType); r != tca {
		t.Errorf("get default state component at specified state returned %p instead of %p", r, tca)
	}

	if r := en.Get(TestCType); r != tcc {
		t.Errorf("get stateful component returned %p instead of %p", r, tcc)
	}

	if r := en.Get(TestBType); r != nil {
		t.Errorf("get stateful component at wrong state returned %p instead of nil", r)
	}

	en.ChangeState("unknown state")

	if r := en.Get(TestAType); r != tca {
		t.Errorf("get default state component at unknown state returned %p instead of %p", r, tca)
	}

	if r := en.Get(TestCType); r != nil {
		t.Errorf("get stateful component at unknown state returned %p instead of nil", r)
	}

	en.ChangeState("even more unknown state")

	if r := en.Get(TestAType); r != tca {
		t.Errorf("get default state component at unknown state returned %p instead of %p", r, tca)
	}

}
