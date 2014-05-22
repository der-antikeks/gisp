package game

import (
	"fmt"
	"log"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"
)

type SceneSystem struct {
	engine *ecs.Engine
	prio   ecs.SystemPriority

	messages chan ecs.Message
	trees    map[string]*SphereTree
}

func NewSceneSystem(engine *ecs.Engine) *SceneSystem {
	s := &SceneSystem{
		engine: engine,
		prio:   PriorityBeforeRender,

		messages: make(chan ecs.Message),
		trees:    map[string]*SphereTree{},
	}

	go func() {
		s.Restart()

		for event := range s.messages {
			switch e := event.(type) {
			case ecs.MessageEntityAdd:
				if err := s.AddEntity(e.Added); err != nil {
					log.Fatal("could not add entity to scene:", err)
				}
			case ecs.MessageEntityUpdate:
				if err := s.UpdateEntity(e.Updated); err != nil {
					log.Fatal("could not update entity:", err)
				}
			case ecs.MessageEntityRemove:
				if err := s.RemoveEntity(e.Removed); err != nil {
					log.Fatal("could not remove entity from scene:", err)
				}
			case ecs.MessageUpdate:
				if err := s.UpdateTrees(); err != nil {
					log.Fatal("could not update scene tree:", err)
				}
			}
		}
	}()

	return s
}

func (s *SceneSystem) Restart() {
	s.engine.Subscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, s.prio, s.messages)

	s.engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType, ecs.EntityUpdateMessageType},
		Aspect: []ecs.ComponentType{TransformationType, SceneTreeType},
	}, s.prio, s.messages)
}

func (s *SceneSystem) Stop() {
	s.engine.Unsubscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, s.messages)

	s.engine.Unsubscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType, ecs.EntityUpdateMessageType},
		Aspect: []ecs.ComponentType{TransformationType, SceneTreeType},
	}, s.messages)

	// TODO: empty trees?
}

func (s *SceneSystem) getData(en ecs.Entity) (stc SceneTree, pos math.Vector, radius float64, err error) {
	ec, err := s.engine.Get(en, TransformationType)
	if err != nil {
		return
	}
	transform := ec.(Transformation)

	ec, err = s.engine.Get(en, SceneTreeType)
	if err != nil {
		return
	}
	stc = ec.(SceneTree)

	ec, err = s.engine.Get(en, GeometryType)
	if err != nil {
		pos = transform.Position
		// TODO: consider parent transformation
		// transform.Parent.MatrixWorld().Transform(pos)

		err = nil
		return
	}

	pos, radius = ec.(Geometry).Bounding.Sphere()
	pos = transform.MatrixWorld().Transform(pos)
	radius *= transform.MatrixWorld().MaxScaleOnAxis()
	return
}

func (s *SceneSystem) AddEntity(en ecs.Entity) error {
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
	if err := s.engine.Set(en, stc); err != nil {
		return err
	}
	return nil
}

func (s *SceneSystem) UpdateEntity(en ecs.Entity) error {
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

func (s *SceneSystem) RemoveEntity(en ecs.Entity) error {
	ec, err := s.engine.Get(en, SceneTreeType)
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
	if err := s.engine.Set(en, stc); err != nil {
		return err
	}
	return nil
}

func (s *SceneSystem) UpdateTrees() error {
	for _, tree := range s.trees {
		tree.Update()
	}
	return nil
}
