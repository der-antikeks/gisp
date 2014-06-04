package game

import (
	"testing"
	"time"
)

type TestMessage int

func TestObserver_Block(t *testing.T) {
	o := NewObserver()

	c := make(chan interface{})
	o.Subscribe(c, 1)

	o.Publish(TestMessage(1))
	o.Publish(TestMessage(2))
	o.Publish(TestMessage(3))
}

func TestObserver_Mux(t *testing.T) {
	o := NewObserver()

	c := make(chan interface{})
	o.Subscribe(c, 1)

	o.Publish(TestMessage(1))
	o.Publish(TestMessage(2))
	o.Publish(TestMessage(3))

	for i := 1; i <= 3; i++ {
		select {
		case v := <-c:
			if r := int(v.(TestMessage)); r != i {
				t.Errorf("received %v instead of %v", r, i)
			}
		case <-time.After(1 * time.Second):
			t.Error("timeout")
		}
	}
}

func TestObserver_Broadcast(t *testing.T) {
	o := NewObserver()

	c1 := make(chan interface{})
	o.Subscribe(c1, 1)
	c2 := make(chan interface{})
	o.Subscribe(c2, 1)

	o.Publish(TestMessage(1))

	select {
	case v := <-c1:
		if r := int(v.(TestMessage)); r != 1 {
			t.Errorf("received %v instead of %v", r, 1)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout")
	}

	select {
	case v := <-c2:
		if r := int(v.(TestMessage)); r != 1 {
			t.Errorf("received %v instead of %v", r, 1)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout")
	}
}

func TestObserver_Unsubscribe(t *testing.T) {
	o := NewObserver()

	c := make(chan interface{})
	o.Subscribe(c, 1)

	o.Publish(TestMessage(1))
	o.Unsubscribe(c)
	o.Publish(TestMessage(2))

	select {
	case v := <-c:
		if r := int(v.(TestMessage)); r != 1 {
			t.Errorf("received %v instead of %v", r, 1)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout")
	}

	select {
	case v := <-c:
		t.Errorf("received %v instead of timeout", v)
	case <-time.After(1 * time.Second):
	}
}

func TestObserver_Sort(t *testing.T) {
	o := NewObserver()

	c1 := make(chan interface{})
	o.Subscribe(c1, 10)
	c2 := make(chan interface{})
	o.Subscribe(c2, 1)

	o.Publish(TestMessage(1))

	select {
	case <-c1:
		t.Error("prio 10 received before prio 1")
	case <-c2:
	case <-time.After(1 * time.Second):
		t.Error("timeout")
	}
}
