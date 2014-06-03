package game

import (
	"fmt" // TODO: for debugging
	"log"
	"math"
	"strings"

	"github.com/der-antikeks/mathgl/mgl32"
)

/*
	collisions, visibility of spatially aware entities

	map[string]*Node	//	root node of scenegraph
	VisibleEntities(scene, frustum) []Entity
	Collisions, FollowNearby
*/
type SpatialSystem struct {
	ents  *EntitySystem
	state *GameStateSystem

	messages chan interface{}
	trees    map[string]*SphereTree
}

func NewSpatialSystem(ents *EntitySystem, state *GameStateSystem) *SpatialSystem {
	s := &SpatialSystem{
		ents:  ents,
		state: state,

		messages: make(chan interface{}),
		trees:    map[string]*SphereTree{},
	}

	go func() {
		s.Restart()

		for event := range s.messages {
			switch e := event.(type) {
			case MessageEntityAdd:
				if err := s.AddEntity(e.Added); err != nil {
					log.Fatal("could not add entity to scene:", err)
				}
			case MessageEntityUpdate:
				if err := s.UpdateEntity(e.Updated); err != nil {
					log.Fatal("could not update entity:", err)
				}
			case MessageEntityRemove:
				if err := s.RemoveEntity(e.Removed); err != nil {
					log.Fatal("could not remove entity from scene:", err)
				}
			case MessageUpdate:
				if err := s.UpdateTrees(); err != nil {
					log.Fatal("could not update scene tree:", err)
				}
			}
		}
	}()

	return s
}

func (s *SpatialSystem) Restart() {
	s.state.OnUpdate().Subscribe(s.messages, PriorityBeforeRender)

	s.ents.OnAdd(TransformationType, SceneTreeType).Subscribe(s.messages, PriorityBeforeRender)
	s.ents.OnUpdate(TransformationType, SceneTreeType).Subscribe(s.messages, PriorityBeforeRender)
	s.ents.OnRemove(TransformationType, SceneTreeType).Subscribe(s.messages, PriorityBeforeRender)
}

func (s *SpatialSystem) Stop() {
	s.state.OnUpdate().Unsubscribe(s.messages)

	s.ents.OnAdd(TransformationType, SceneTreeType).Unsubscribe(s.messages)
	s.ents.OnUpdate(TransformationType, SceneTreeType).Unsubscribe(s.messages)
	s.ents.OnRemove(TransformationType, SceneTreeType).Unsubscribe(s.messages)

	// TODO: empty trees?
}

func (s *SpatialSystem) getData(en Entity) (stc SceneTree, pos mgl32.Vec3, radius float32, err error) {
	ec, err := s.ents.Get(en, TransformationType)
	if err != nil {
		return
	}
	transform := ec.(Transformation)

	ec, err = s.ents.Get(en, SceneTreeType)
	if err != nil {
		return
	}
	stc = ec.(SceneTree)

	ec, err = s.ents.Get(en, GeometryType)
	if err != nil {
		pos = transform.Position
		// TODO: consider parent transformation
		// transform.Parent.MatrixWorld().Transform(pos)

		err = nil
		return
	}

	pos, radius = ec.(Geometry).Bounding.Sphere()
	pos4 := transform.MatrixWorld().Mul4x1(mgl32.Vec4{pos[0], pos[1], pos[2], 0})
	pos = mgl32.Vec3{pos4[0], pos4[1], pos4[2]}
	radius *= transform.MatrixWorld().MaxScale()
	return
}

func (s *SpatialSystem) AddEntity(en Entity) error {
	stc, pos, radius, err := s.getData(en)
	if err != nil {
		return err
	}
	if stc.leaf != nil {
		return fmt.Errorf("added entity with existing leaf node")
	}

	tree, ok := s.trees[stc.Name]
	if !ok {
		// new scene
		tree = NewSphereTree(0.0)
		s.trees[stc.Name] = tree
	}

	stc.leaf = tree.Add(pos, radius)
	if err := s.ents.Set(en, stc); err != nil {
		return err
	}
	return nil
}

func (s *SpatialSystem) UpdateEntity(en Entity) error {
	stc, pos, radius, err := s.getData(en)
	if err != nil {
		return err
	}
	if stc.leaf == nil {
		return fmt.Errorf("updating entity without leaf node")
	}
	if err := stc.leaf.Update(pos, radius); err != nil {
		return err
	}
	return nil
}

func (s *SpatialSystem) RemoveEntity(en Entity) error {
	ec, err := s.ents.Get(en, SceneTreeType)
	if err != nil {
		return err
	}

	stc := ec.(SceneTree)
	if stc.leaf == nil {
		return fmt.Errorf("removing entity without leaf node")
	}

	if err := stc.leaf.Delete(); err != nil {
		return err
	}

	stc.leaf = nil
	if err := s.ents.Set(en, stc); err != nil {
		return err
	}
	return nil
}

func (s *SpatialSystem) UpdateTrees() error {
	for _, tree := range s.trees {
		tree.Update()
	}
	return nil
}

type SphereTree struct {
	root                   *Node
	restraint              float32
	pool, sinsert, srecalc []*Node
}

func NewSphereTree(restraint float32) *SphereTree {
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
	n.center, n.radius = mgl32.Vec3{}, 0.0
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
func (t *SphereTree) Add(p mgl32.Vec3, r float32) *Node {
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
	mindist := float32(math.Inf(1))
	t.root.walk(func(p *Node) bool {
		dist := p.center.Sub(n.center).Len()
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
		dist := diff.Len()
		v := diff.Mul(1.0 / dist)
		min := float32(math.Min(float64(-n.radius), float64(dist-n.children[i].radius)))
		max := (float32(math.Max(float64(n.radius), float64(dist+n.children[i].radius))) - min) * 0.5

		n.center = n.center.Add(v.Mul(max + min))
		n.radius = max + t.restraint
	}
}

func (t *SphereTree) mergeNodes(a, b *Node) *Node {
	diff := b.center.Sub(a.center)
	dist := diff.Len()

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

	v := diff.Mul(1.0 / dist)
	min := float32(math.Min(float64(-a.radius), float64(dist-b.radius)))
	max := (float32(math.Max(float64(a.radius), float64(dist+b.radius))) - min) * 0.5

	n := t.get()
	n.typ = BranchNode
	n.center = a.center.Add(v.Mul(max + min))
	n.radius = max
	return n
}

type NodeType int

const (
	LeafNode NodeType = iota
	BranchNode
)

type Node struct {
	center mgl32.Vec3
	radius float32

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

func (n *Node) Update(p mgl32.Vec3, r float32) error {
	if n.typ != LeafNode {
		return fmt.Errorf("updating node that is not a leaf: %v", n)
	}

	if n.parent == nil {
		// node is root and no branch!
		n.center = p
		n.radius = r
		return nil
	}

	dist := n.parent.center.Sub(p).Len()
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
	dist := n.center.Sub(b.center).Len()

	if n.radius+b.radius < dist {
		return Disjoint
	}
	if n.radius-b.radius < dist {
		return Intersects
	}
	return Contains
}

type Boundary struct {
	Min, Max mgl32.Vec3
}

func NewBoundary() Boundary {
	p, m := float32(math.Inf(1)), float32(math.Inf(-1))
	return Boundary{
		Min: mgl32.Vec3{p, p, p},
		Max: mgl32.Vec3{m, m, m},
	}
}

func BoundaryFromPoints(pts ...mgl32.Vec3) Boundary {
	b := NewBoundary()
	for _, p := range pts {
		b.AddPoint(p)
	}

	return b
}

func (b Boundary) ApproxEqual(e Boundary) bool {
	return b.Min.ApproxEqual(e.Min) && b.Max.ApproxEqual(e.Max)
}

func min(a, b float32) float32 { return float32(math.Min(float64(a), float64(b))) }
func max(a, b float32) float32 { return float32(math.Max(float64(a), float64(b))) }

func (b *Boundary) AddPoint(p mgl32.Vec3) {
	b.Min[0], b.Max[0] = min(b.Min[0], p[0]), max(b.Max[0], p[0])
	b.Min[1], b.Max[1] = min(b.Min[1], p[1]), max(b.Max[1], p[1])
	b.Min[2], b.Max[2] = min(b.Min[2], p[2]), max(b.Max[2], p[2])
}

func (b *Boundary) AddBoundary(a Boundary) {
	if b.ApproxEqual(a) {
		return
	}

	b.AddPoint(a.Max)
	b.AddPoint(a.Min)
}

func (b Boundary) Center() mgl32.Vec3 {
	return b.Min.Add(b.Max).Mul(0.5)
}

func (b Boundary) Size() mgl32.Vec3 {
	return b.Max.Sub(b.Min)
}

func (b Boundary) Sphere() (center mgl32.Vec3, radius float32) {
	return b.Center(), b.Size().Len() * 0.5
}

type Plane struct {
	normal   mgl32.Vec4
	distance float32
}

func (p Plane) Normalize() Plane {
	magnitude := p.normal.Len()

	return Plane{
		normal:   p.normal.Mul(1.0 / magnitude),
		distance: p.distance / magnitude,
	}
}

type Frustum [6]Plane

func Mat4ToFrustum(m mgl32.Mat4) Frustum {
	f := Frustum{
		Plane{mgl32.Vec4{m[3] - m[0], m[7] - m[4], m[11] - m[8]}, m[15] - m[12]},
		Plane{mgl32.Vec4{m[3] + m[0], m[7] + m[4], m[11] + m[8]}, m[15] + m[12]},
		Plane{mgl32.Vec4{m[3] + m[1], m[7] + m[5], m[11] + m[9]}, m[15] + m[13]},
		Plane{mgl32.Vec4{m[3] - m[1], m[7] - m[5], m[11] - m[9]}, m[15] - m[13]},
		Plane{mgl32.Vec4{m[3] - m[2], m[7] - m[6], m[11] - m[10]}, m[15] - m[14]},
		Plane{mgl32.Vec4{m[3] + m[2], m[7] + m[6], m[11] + m[10]}, m[15] + m[14]},
	}

	for i, p := range f {
		f[i] = p.Normalize()
	}

	return f
}

func (f Frustum) ContainsPoint(point mgl32.Vec3) bool {
	p4 := mgl32.Vec4{point[0], point[1], point[2], 0}
	for _, p := range f {
		if p4.Dot(p.normal)+p.distance <= 0 {
			return false
		}
	}

	return true
}

func (f Frustum) IntersectsSphere(center mgl32.Vec3, radius float32) bool {
	c4 := mgl32.Vec4{center[0], center[1], center[2], 0}
	for _, p := range f {
		if c4.Dot(p.normal)+p.distance <= -radius {
			return false
		}
	}

	return true
}
