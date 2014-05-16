package game

import (
	"log"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"
)

type MotionControlSystem struct {
	engine *ecs.Engine
	im     *InputManager
	prio   ecs.SystemPriority

	messages    chan ecs.Message
	controlable []ecs.Entity
}

func NewControlSystem(engine *ecs.Engine, im *InputManager) *MotionControlSystem {
	s := &MotionControlSystem{
		engine: engine,
		im:     im,
		prio:   PriorityBeforeRender,

		messages:    make(chan ecs.Message),
		controlable: []ecs.Entity{},
	}

	go func() {
		s.Restart()

		for event := range s.messages {
			switch e := event.(type) {
			case ecs.MessageEntityAdd:
				s.controlable = append(s.controlable, e.Added)
			case ecs.MessageEntityRemove:
				for i, f := range s.controlable {
					if f == e.Removed {
						s.controlable = append(s.controlable[:i], s.controlable[i+1:]...)
						break
					}
				}

			case ecs.MessageUpdate:
				if err := s.Update(); err != nil {
					log.Fatal("could not update game state:", err)
				}
			}
		}
	}()

	return s
}

func (s *MotionControlSystem) Restart() {
	s.engine.Subscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, s.prio, s.messages)

	s.engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{TransformationType, MotionControlType},
	}, s.prio, s.messages)
}

func (s *MotionControlSystem) Stop() {
	s.engine.Unsubscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, s.messages)

	s.engine.Unsubscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{TransformationType, MotionControlType},
	}, s.messages)

	s.controlable = s.controlable[:0]
}

func (s *MotionControlSystem) Update() error {
	for _, en := range s.controlable {

		ec, err := s.engine.Get(en, TransformationType)
		if err != nil {
			return err
		}
		transform := ec.(Transformation)

		ec, err = s.engine.Get(en, MotionControlType)
		if err != nil {
			return err
		}
		control := ec.(MotionControl)

		var updateTransform bool

		if s.im.IsKeyDown(control.ForwardKey) {
			p := (math.Vector{0, 0, 1}).MulScalar(control.MovementSpeed)
			transform.Position = transform.Position.Add(p)
			updateTransform = true
		}

		if s.im.IsKeyDown(control.BackwardKey) {
			p := (math.Vector{0, 0, -1}).MulScalar(control.MovementSpeed)
			transform.Position = transform.Position.Add(p)
			updateTransform = true
		}

		if s.im.IsKeyDown(control.LeftKey) {
			p := (math.Vector{-1, 0, 0}).MulScalar(control.MovementSpeed)
			transform.Position = transform.Position.Add(p)
			updateTransform = true
		}

		if s.im.IsKeyDown(control.RightKey) {
			p := (math.Vector{1, 0, 0}).MulScalar(control.MovementSpeed)
			transform.Position = transform.Position.Add(p)
			updateTransform = true
		}

		if s.im.IsKeyDown(control.UpKey) {
			p := (math.Vector{0, 1, 0}).MulScalar(control.MovementSpeed)
			transform.Position = transform.Position.Add(p)
			updateTransform = true
		}

		if s.im.IsKeyDown(control.DownKey) {
			p := (math.Vector{0, -1, 0}).MulScalar(control.MovementSpeed)
			transform.Position = transform.Position.Add(p)
			updateTransform = true
		}

		if s.im.IsKeyDown(control.RotLeft) {
			r := math.QuaternionFromAxisAngle(math.Vector{0, 1, 0}, control.RotationSpeed)
			transform.Rotation = transform.Rotation.Mul(r)
			updateTransform = true
		}

		if s.im.IsKeyDown(control.RotRight) {
			r := math.QuaternionFromAxisAngle(math.Vector{0, -1, 0}, control.RotationSpeed)
			transform.Rotation = transform.Rotation.Mul(r)
			updateTransform = true
		}

		if updateTransform {
			if err := s.engine.Set(en, transform); err != nil {
				return err
			}
		}
	}
	return nil
}
