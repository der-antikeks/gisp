package game

import (
	"log"
	m "math"

	"github.com/der-antikeks/gisp/math"
)

type SphereTree struct {
	root      *Node
	restraint float64
	pool      []*Node
}

func NewSphereTree(restraint float64) *SphereTree {
	return &SphereTree{
		restraint: restraint,
	}
}

func (t *SphereTree) put(n *Node) {
	n.parent, n.children = nil, nil
	n.center, n.radius = math.Vector{}, 0.0
	t.pool = append(t.pool, n)
}

func (t *SphereTree) get() *Node {
	if l := len(t.pool) - 1; l >= 0 {
		n := t.pool[l]
		t.pool = t.pool[:l]
		return n
	}
	return &Node{tree: t}
}

// creates a new node in the size of the passed bounding sphere
func (t *SphereTree) Add(p math.Vector, r float64) *Node {
	if t.root == nil {
		t.root = t.get()
		t.root.typ = BranchNode
		t.root.center = p
		t.root.radius = r + t.restraint
	}

	n := t.get()
	n.typ = LeafNode
	n.center = p
	n.radius = r

	t.insert(t.root, n)
	return n
}

// add node to root/traverse children
func (t *SphereTree) insert(p, c *Node) {
	if d := p.center.Sub(c.center).Length(); d+p.radius < c.radius {
		// parent is inside of child
		if p.parent == nil {
			// parent is root
			t.root = t.mergeNodes(p, c)
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
		branch := t.mergeNodes(sibling, c)

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

// delete branch nodes without children
// recalculate the bounding sphere of the sum of the children
func (t *SphereTree) recalc(n *Node) {
	if n.typ == LeafNode {
		// leaf node, do nothing
		return
	}
	if n.typ == BranchNode && len(n.children) == 0 {
		// empty branch, delete
		if n.parent != nil {
			n.parent.removeChild(n)
		}
		t.put(n)
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
		n.radius = max + t.restraint
	}
}

func (t *SphereTree) mergeNodes(a, b *Node) *Node {
	diff := b.center.Sub(a.center)
	dist := diff.Length()

	if a.radius+b.radius >= dist {
		// intersects
		if a.radius-b.radius >= dist {
			// b inside a
			n := t.get()
			n.typ = BranchNode
			n.center = a.center
			n.radius = a.radius
			return n
		}
		if b.radius-a.radius >= dist {
			// a inside b
			n := t.get()
			n.typ = BranchNode
			n.center = b.center
			n.radius = b.radius
			return n
		}
	}

	v := diff.MulScalar(1.0 / dist)
	min := m.Min(-a.radius, dist-b.radius)
	max := (m.Max(a.radius, dist+b.radius) - min) * 0.5

	n := t.get()
	n.typ = BranchNode
	n.center = a.center.Add(v.MulScalar(max + min))
	n.radius = max
	return n
}

type NodeType int

const (
	LeafNode NodeType = iota
	BranchNode
)

type Node struct {
	center math.Vector
	radius float64

	typ      NodeType
	tree     *SphereTree
	parent   *Node
	children []*Node
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

func (n *Node) Update(p math.Vector, r float64) {
	if n.typ != LeafNode {
		log.Fatalln("updating node that is not a leaf: ", n)
	}

	dist := n.center.Sub(p).Length()
	if dist+r > n.radius {
		n.center = p
		n.radius = r

		if n.parent != nil {
			n.parent.removeChild(n)
			n.parent = nil
			n.tree.recalc(n.parent)
		}

		n.tree.insert(n.tree.root, n)
	}
}

func (n *Node) Delete() {
	if n.typ != LeafNode {
		log.Fatalln("deleting node that is not a leaf: ", n)
	}

	if n.parent != nil {
		n.parent.removeChild(n)
		n.tree.recalc(n.parent)
	}

	n.tree.put(n)
}

type IntersectionType int

const (
	Disjoint IntersectionType = iota
	Intersects
	Contains
)

func (n *Node) Intersects(b *Node) IntersectionType {
	dist := n.center.Sub(b.center).Length()

	if n.radius+b.radius < dist {
		return Disjoint
	}
	if n.radius-b.radius < dist {
		return Intersects
	}
	return Contains
}
