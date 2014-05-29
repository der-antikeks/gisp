package game

import (
	"log"
	"time"

	"github.com/der-antikeks/gisp/math"
)

type MovementSystem struct {
	ents  *EntitySystem
	state *GameStateSystem

	messages chan Message

	moveable []Entity
}

func NewMovementsSystem(ents *EntitySystem, state *GameStateSystem) *MovementSystem {
	s := &MovementSystem{
		ents:  ents,
		state: state,

		messages: make(chan Message),
	}

	go func() {
		s.Restart()

		for event := range s.messages {
			switch e := event.(type) {
			case MessageEntityAdd:
				s.moveable = append(s.moveable, e.Added)
			case MessageEntityRemove:
				for i, f := range s.moveable {
					if f == e.Removed {
						s.moveable = append(s.moveable[:i], s.moveable[i+1:]...)
						break
					}
				}

			case MessageUpdate:
				if err := s.Update(e.Delta); err != nil {
					log.Fatal("could not update movement:", err)
				}
			}
		}
	}()

	return s
}

func (s *MovementSystem) Restart() {
	s.state.OnUpdate().Subscribe(s.messages, PriorityBeforeRender)

	s.ents.OnAdd(TransformationType, VelocityType).Subscribe(s.messages, PriorityBeforeRender)
	s.ents.OnRemove(TransformationType, VelocityType).Subscribe(s.messages, PriorityBeforeRender)
}

func (s *MovementSystem) Stop() {
	s.state.OnUpdate().Unsubscribe(s.messages)

	s.ents.OnAdd(TransformationType, VelocityType).Unsubscribe(s.messages)
	s.ents.OnRemove(TransformationType, VelocityType).Unsubscribe(s.messages)

	s.moveable = []Entity{}
}

func (s *MovementSystem) Update(delta time.Duration) error {
	for _, en := range s.moveable {
		ec, err := s.ents.Get(en, TransformationType)
		if err != nil {
			return err
		}
		transform := ec.(Transformation)

		ec, err = s.ents.Get(en, VelocityType)
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

			if err := s.ents.Set(en, transform); err != nil {
				return err
			}
		}
	}
	return nil
}
