package game

import (
	"github.com/willf/bitset"
)

// General purpose object that identifies a set of components.
type Entity uint

func (e Entity) Uint() uint {
	return uint(e)
}

const (
	NoEntity Entity = 0
)

type EntityCollection interface {
	Set(Entity)
	Get(Entity) bool
	Remove(Entity)
	Length() int
	Iterate(func(Entity) bool)
}

// not safe for concurrent use
// slow set/get/remove, fast iterate
type EntitySlice struct {
	data []Entity
}

// faster set but does not test for duplicate entries
func (c *EntitySlice) SetUnsafe(e Entity) {
	c.data = append(c.data, e)
}

func (c *EntitySlice) Set(e Entity) {
	if c.Get(e) {
		return
	}
	c.SetUnsafe(e)
}

func (c *EntitySlice) Get(e Entity) bool {
	for _, f := range c.data {
		if f == e {
			return true
		}
	}
	return false
}

func (c *EntitySlice) Remove(e Entity) {
	for i, f := range c.data {
		if f == e {
			c.data = append(c.data[:i], c.data[i+1:]...)
			return
		}
	}
}

func (c *EntitySlice) Length() int {
	return len(c.data)
}

func (c *EntitySlice) Iterate(f func(Entity) bool) {
	for _, e := range c.data {
		if !f(e) {
			return
		}
	}
}

// not safe for concurrent use
// fast set/get/remove, slow iterate
type EntityMap struct {
	data map[Entity]struct{}
}

func (c *EntityMap) Set(e Entity) {
	if c.data == nil {
		c.data = map[Entity]struct{}{}
	}
	c.data[e] = struct{}{}
}

func (c *EntityMap) Get(e Entity) bool {
	_, ok := c.data[e]
	return ok
}

func (c *EntityMap) Remove(e Entity) {
	delete(c.data, e)
}

func (c *EntityMap) Length() int {
	return len(c.data)
}

func (c *EntityMap) Iterate(f func(Entity) bool) {
	for e := range c.data {
		if !f(e) {
			return
		}
	}
}

// not safe for concurrent use
// fast set/get/remove, slow iterate
type EntityBitset struct {
	set bitset.BitSet
}

func (c *EntityBitset) Set(e Entity) {
	c.set.Set(e.Uint())
}

func (c *EntityBitset) Get(e Entity) bool {
	return c.set.Test(e.Uint())
}

func (c *EntityBitset) Remove(e Entity) {
	c.set.Clear(e.Uint())
}

func (c *EntityBitset) Length() int {
	return int(c.set.Count())
}

func (c *EntityBitset) Iterate(f func(Entity) bool) {
	for i, ok := c.set.NextSet(0); ok; i, ok = c.set.NextSet(i + 1) {
		if !f(Entity(i)) {
			return
		}
	}
}
