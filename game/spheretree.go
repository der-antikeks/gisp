package game

import (
	"fmt" // TODO: for debugging
	m "math"
	"strings"

	"github.com/der-antikeks/gisp/math"
)

type SphereTree struct {
	root                   *Node
	restraint              float64
	pool, sinsert, srecalc []*Node
}

func NewSphereTree(restraint float64) *SphereTree {
	return &SphereTree{
		restraint: restraint,
	}
}

// TODO: for debugging
func (t *SphereTree) String() string {
	s := fmt.Sprintf("SphereTree(%v), %v, %v, %v\n", t.restraint, len(t.pool), len(t.sinsert), len(t.srecalc))
	if t.root != nil {
		s += t.root.StringWalk(0)
	} else {
		s += "nil"
	}
	return s
}

// add empty node to pool
func (t *SphereTree) put(n *Node) {
	n.parent, n.children = nil, nil
	n.center, n.radius = math.Vector{}, 0.0
	t.pool = append(t.pool, n)
}

// get free node from pool or creates new if empty
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
		t.root.typ = LeafNode // BranchNode
		t.root.center = p
		t.root.radius = r + t.restraint
		return t.root
	}

	n := t.get()
	n.typ = LeafNode
	n.center = p
	n.radius = r

	// TODO: prevent Update()->parent before scheduleInsert
	//t.scheduleInsert(n)
	t.insert(n)
	return n
}

func (t *SphereTree) scheduleInsert(n *Node) {
	t.sinsert = append(t.sinsert, n)
}

func (t *SphereTree) scheduleRecalc(n *Node) {
	t.srecalc = append(t.srecalc, n)
}

func (t *SphereTree) Update() {
	// recalc
	var n *Node
	for len(t.srecalc) > 0 {
		n, t.srecalc = t.srecalc[0], t.srecalc[1:]
		t.recalc(n)
	}

	// insert
	for len(t.sinsert) > 0 {
		n, t.sinsert = t.sinsert[0], t.sinsert[1:]
		t.insert(n)
	}
}

// add node to root/traverse children
func (t *SphereTree) insert(n *Node) {
	var sibling *Node
	mindist := m.Inf(1)
	t.root.walk(func(p *Node) bool {
		dist := p.center.Sub(n.center).Length()
		if p.radius+n.radius < dist {
			// Disjoint
			return false
		}
		if p.radius-n.radius < dist {
			// Intersects
			if p.typ == BranchNode {
				return false
			}
		} else {
			// Contains
		}

		// TODO: still have to think about whether that is optimal
		//if dist+n.radius <= mindist {
		if dist <= mindist {
			mindist = dist
			sibling = p
		}

		return true
	})

	if sibling == nil {
		// child node is bigger than root
		// create a new, bigger root
		branch := t.mergeNodes(t.root, n)
		branch.radius += t.restraint

		branch.addChild(t.root)
		t.root.parent = branch

		t.root = branch
		sibling = branch
	}
	if sibling.typ == LeafNode {
		// nearest sibling is leaf, create branch wich encloses both

		branch := t.mergeNodes(sibling, n)
		branch.radius += t.restraint
		branch.parent = sibling.parent

		if sibling.parent == nil {
			// sibling was root
			t.root = branch
		} else {
			sibling.parent.removeChild(sibling)
			sibling.parent.addChild(branch)
		}

		branch.addChild(sibling)
		sibling.parent = branch

		sibling = branch
	}

	sibling.addChild(n)
	n.parent = sibling
}

// delete branch nodes without children
// recalculate the bounding sphere of the sum of the children
func (t *SphereTree) recalc(n *Node) {
	if n.typ == LeafNode {
		// leaf node, do nothing
		return
	}
	if len(n.children) == 0 {
		// empty branch, delete
		if n.parent != nil {
			n.parent.removeChild(n)
			//t.scheduleRecalc(n.parent)
			t.recalc(n.parent)
		}
		t.put(n)
		return
	}
	if len(n.children) == 1 {
		// single child, move branch up
		if n.parent != nil {
			n.children[0].parent = n.parent
			n.parent.removeChild(n)
			n.parent.addChild(n.children[0])
		} else {
			n.children[0].parent = nil
			t.root = n.children[0]
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

// TODO: for debugging
func (n *Node) String() string {
	t := "Leaf"
	if n.typ == BranchNode {
		t = "Branch"
	}
	return fmt.Sprintf("{%s: %v: %5.2f}", t, n.center, n.radius)
}

// TODO: for debugging
func (n *Node) StringWalk(level int) string {
	s := strings.Repeat("  ", level) + n.String() + "\n"
	for _, c := range n.children {
		s += c.StringWalk(level + 1)
	}
	return s
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

func (n *Node) walk(f func(n *Node) bool) {
	if !f(n) {
		return
	}
	for _, c := range n.children {
		c.walk(f)
	}
}

func (n *Node) Update(p math.Vector, r float64) error {
	if n.typ != LeafNode {
		return fmt.Errorf("updating node that is not a leaf: %v", n)
	}

	if n.parent == nil {
		// node is root and no branch!
		n.center = p
		n.radius = r
		return nil
	}

	dist := n.parent.center.Sub(p).Length()
	if dist+r > n.parent.radius {
		// fits no longer into parent
		// remove from parent
		n.parent.removeChild(n)
		//n.tree.scheduleRecalc(n.parent)
		n.tree.recalc(n.parent)
		n.parent = nil

		// and reinsert
		n.center = p
		n.radius = r
		//n.tree.scheduleInsert(n)
		n.tree.insert(n)
	}
	return nil
}

func (n *Node) Delete() error {
	if n.typ != LeafNode {
		return fmt.Errorf("deleting node that is not a leaf: %v", n)
	}

	if n.parent != nil {
		n.parent.removeChild(n)
		//n.tree.scheduleRecalc(n.parent)
		n.tree.recalc(n.parent)
	} else {
		n.tree.root = nil
	}

	n.tree.put(n)
	return nil
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
