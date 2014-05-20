package game

import (
	"log"
	"time"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"
)

type MovementSystem struct {
	engine *ecs.Engine
	prio   ecs.SystemPriority

	messages chan ecs.Message

	moveable []ecs.Entity
}

func NewMovementSystem(engine *ecs.Engine) *MovementSystem {
	s := &MovementSystem{
		engine: engine,
		prio:   PriorityBeforeRender,

		messages: make(chan ecs.Message),
	}

	go func() {
		s.Restart()

		for event := range s.messages {
			switch e := event.(type) {
			case ecs.MessageEntityAdd:
				s.moveable = append(s.moveable, e.Added)
			case ecs.MessageEntityRemove:
				for i, f := range s.moveable {
					if f == e.Removed {
						s.moveable = append(s.moveable[:i], s.moveable[i+1:]...)
						break
					}
				}

			case ecs.MessageUpdate:
				if err := s.Update(e.Delta); err != nil {
					log.Fatal("could not update game state:", err)
				}
			}
		}
	}()

	return s
}

func (s *MovementSystem) Restart() {
	s.engine.Subscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, s.prio, s.messages)

	s.engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{TransformationType, VelocityType},
	}, s.prio, s.messages)
}

func (s *MovementSystem) Stop() {
	s.engine.Unsubscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, s.messages)

	s.engine.Unsubscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{TransformationType, VelocityType},
	}, s.messages)

	s.moveable = []ecs.Entity{}
}

func (s *MovementSystem) Update(delta time.Duration) error {
	for _, en := range s.moveable {
		ec, err := s.engine.Get(en, TransformationType)
		if err != nil {
			return err
		}
		transform := ec.(Transformation)

		ec, err = s.engine.Get(en, VelocityType)
		if err != nil {
			return err
		}
		velocity := ec.(Velocity)

		var update bool
		if v := velocity.Velocity; !v.Equals(math.Vector{}, 6) {
			update = true
			transform.Position = transform.Position.Add(v.MulScalar(delta.Seconds()))
		}

		if a := velocity.Angular; !a.Equals(math.Vector{}, 6) {
			update = true

			// http://www.euclideanspace.com/physics/kinematics/angularvelocity/#quaternion
			// dq/dt = 1/2 w(t) q(t)

			q := transform.Rotation.Normalize()
			a = a.MulScalar(delta.Seconds())
			w := math.Quaternion{a[0], a[1], a[2], 0}
			transform.Rotation = transform.Rotation.Add(w.Mul(q).MulScalar(0.5)).Normalize()
		}

		if update {
			transform.updatedMatrix = false // TODO: remove

			if err := s.engine.Set(en, transform); err != nil {
				return err
			}
		}
	}
	return nil
}
