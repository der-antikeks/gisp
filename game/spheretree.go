package game

import (
	"log"
	m "math"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"
)

type SphereTree struct {
	root      *Node
	restraint float64
	lookup    map[ecs.Entity]*Node
	// add/remove/recompute nodes only on rendersysmte.update?
}

func NewSphereTree() *SphereTree {
	return &SphereTree{
		restraint: 1.0,
		lookup:    map[ecs.Entity]*Node{},
	}
}

// update/insert entity with bounding sphere
func (t *SphereTree) Add(e ecs.Entity, p math.Vector, r float64) {
	if t.root == nil {
		t.root = &Node{
			center: p,
			radius: r,
		}
	}
	n := &Node{
		entity: e,
		center: p,
		radius: r,
	}
	t.lookup[e] = n
	t.insert(t.root, n)
}
func (t *SphereTree) Update(e ecs.Entity, p math.Vector, r float64) {
	n, ok := t.lookup[e]
	if !ok {
		t.Add(e, p, r)
	}

	dist := n.center.Sub(p).Length()
	if dist+r > n.radius {
		n.center = p
		n.radius = r

		if n.parent != nil {
			n.parent.removeChild(n)
			n.parent = nil
			t.recalc(n.parent)
		}

		t.insert(t.root, n)
	}
}
func (t *SphereTree) Remove(e ecs.Entity) {
	n, ok := t.lookup[e]
	if !ok {
		return
	}
	delete(t.lookup, e)

	n.entity = -1
	t.recalc(n)
}

// add node to root/traverse children
func (t *SphereTree) insert(p, c *Node) {
	if d := p.center.Sub(c.center).Length(); d+p.radius < c.radius {
		// parent is inside of child
		if p.parent == nil {
			// parent is root
			t.root = p.Merge(c)
			t.root.addChild(p)
			p.parent = t.root

			p = t.root
		} else {
			// this should not happen
			log.Fatalln("parent is inside of new node but is not root!")
		}
	}

	var sibling *Node
	var dist float64
	mindist := m.Inf(1)
	for _, s := range p.children {
		dist = s.center.Sub(c.center).Length()
		if dist < mindist {
			if c.radius+dist < s.radius {
				// close enough sibling
				sibling = s
				mindist = dist
			}
		}
	}

	if sibling != nil {
		// create new branch with sibling
		branch := sibling.Merge(c)

		p.addChild(branch)
		branch.parent = p

		p.removeChild(sibling)
		branch.addChild(sibling)
		sibling.parent = branch

		branch.addChild(c)
		c.parent = branch
	} else {
		// add to parent
		p.addChild(c)
		c.parent = p
	}

	t.recalc(p)
}

// check if has childrens, delete
// recalculate bounding sphere sum of childrens
func (t *SphereTree) recalc(n *Node) {
	if n.entity == -1 && len(n.children) == 0 {
		// no leaf and no branch, delete
		n.parent.removeChild(n)
		n.parent = nil
	}
	if n.entity != 0 {
		// leaf node, do nothing
		return
	}

	// recalculate center and radius over all children
	n.center = n.children[0].center
	n.radius = n.children[0].radius
	for i := 1; i < len(n.children); i++ {
		diff := n.children[i].center.Sub(n.center)
		dist := diff.Length()
		v := diff.MulScalar(1.0 / dist)
		min := m.Min(-n.radius, dist-n.children[i].radius)
		max := (m.Max(n.radius, dist+n.children[i].radius) - min) * 0.5

		n.center = n.center.Add(v.MulScalar(max + min))
		n.radius = max
	}
}

type Node struct {
	center math.Vector
	radius float64

	parent   *Node
	children []*Node
	entity   ecs.Entity
}

func (n *Node) addChild(c *Node) {
	n.children = append(n.children, c)
}
func (n *Node) removeChild(c *Node) {
	for i, f := range n.children {
		if f == c {
			copy(n.children[i:], n.children[i+1:])
			n.children[len(n.children)-1] = nil
			n.children = n.children[:len(n.children)-1]
			return
		}
	}
}

func (n *Node) Merge(b *Node) *Node {
	diff := b.center.Sub(n.center)
	dist := diff.Length()

	if n.radius+b.radius >= dist {
		// intersects
		if n.radius-b.radius >= dist {
			// b inside a
			return &Node{
				center: n.center,
				radius: n.radius,
			}
		}
		if b.radius-n.radius >= dist {
			// a inside b
			return &Node{
				center: b.center,
				radius: b.radius,
			}
		}
	}

	v := diff.MulScalar(1.0 / dist)
	min := m.Min(-n.radius, dist-b.radius)
	max := (m.Max(n.radius, dist+b.radius) - min) * 0.5

	return &Node{
		center: n.center.Add(v.MulScalar(max + min)),
		radius: max,
	}
}

type Intersection int

const (
	Disjoint Intersection = iota
	Intersects
	Contains
)

func (n *Node) Intersects(b *Node) Intersection {
	dist := n.center.Sub(b.center).Length()

	if n.radius+b.radius < dist {
		return Disjoint
	}
	if n.radius-b.radius < dist {
		return Intersects
	}
	return Contains
}
