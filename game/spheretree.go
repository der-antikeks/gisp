package game

import (
	m "math"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"
)

type SphereTree struct {
	root   *Node
	lookup map[ecs.Entity]*Node
	// add/remove/recompute nodes only on rendersysmte.update?
}

// update/insert entity with bounding sphere
func (s *SphereTree) AddEntity(en ecs.Entity, pos math.Vector, rad float64)    {}
func (s *SphereTree) UpdateEntity(en ecs.Entity, pos math.Vector, rad float64) {}
func (s *SphereTree) RemoveEntity(en ecs.Entity)                               {}

// add node to root/traverse children
func (s *SphereTree) AddNode(n *Node) {}

// delete node from root/children
func (s *SphereTree) RemoveNode(n *Node) {}

type Node struct {
	center math.Vector
	radius float64

	parent   *Node
	children []*Node
}

// check if has childrens, delete
// recalculate bounding sphere sum of childrens
func (n *Node) Recompute(offset float64) {}
