package game

import (
	"log"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"
)

type OrbitControlSystem struct {
	engine *ecs.Engine
	im     *InputManager
	prio   ecs.SystemPriority

	messages    chan ecs.Message
	controlable []ecs.Entity
}

func NewOrbitControlSystem(engine *ecs.Engine, im *InputManager) *OrbitControlSystem {
	s := &OrbitControlSystem{
		engine: engine,
		im:     im,
		prio:   PriorityBeforeRender,

		messages:    make(chan ecs.Message),
		controlable: []ecs.Entity{},
	}

	go func() {
		s.Restart()

		var dragging bool
		var oldx, oldy float64

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

			case MessageMouseButton:
				if !dragging && s.im.IsMouseDown(MouseRight) {
					dragging = true
					oldx, oldy = s.im.MousePos()
				} else if dragging && !s.im.IsMouseDown(MouseRight) {
					dragging = false
				}

			case MessageMouseMove:

			case ecs.MessageUpdate:
				if !dragging {
					continue
				}

				x, y := s.im.MousePos()
				deltax, deltay := x-oldx, y-oldy
				oldx, oldy = x, y

				if deltax != 0 || deltay != 0 {
					if err := s.Update(deltax, deltay, 0); err != nil {
						log.Fatal("could not update game state:", err)
					}
				}
			}
		}
	}()

	return s
}

func (s *OrbitControlSystem) Restart() {
	s.engine.Subscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType, MouseButtonMessageType, MouseMoveMessageType},
	}, s.prio, s.messages)

	s.engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{TransformationType, OrbitControlType},
	}, s.prio, s.messages)
}

func (s *OrbitControlSystem) Stop() {
	s.engine.Unsubscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType, MouseButtonMessageType, MouseMoveMessageType},
	}, s.messages)

	s.engine.Unsubscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{TransformationType, OrbitControlType},
	}, s.messages)

	s.controlable = s.controlable[:0]
}

func (s *OrbitControlSystem) Update(deltax, deltay, deltaz float64) error {
	for _, en := range s.controlable {

		ec, err := s.engine.Get(en, TransformationType)
		if err != nil {
			return err
		}
		transform := ec.(Transformation)

		ec, err = s.engine.Get(en, OrbitControlType)
		if err != nil {
			return err
		}
		control := ec.(OrbitControl)

		var target math.Vector
		if control.Target != 0 {
			ec, err = s.engine.Get(control.Target, TransformationType)
			if err != nil {
				return err
			}
			target = ec.(Transformation).Position
		}

		distance := transform.Position.Sub(target)
		delta := math.QuaternionFromEuler(math.Vector{
			deltax * control.RotationSpeed,
			deltay * control.RotationSpeed,
			0,
		}, math.RotateXYZ).Inverse()

		transform.Position = delta.Rotate(distance).Add(target)
		//transform.Rotation = transform.Rotation.Mul(delta)
		transform.Rotation = math.QuaternionFromRotationMatrix(math.LookAt(transform.Position, target, transform.Up))

		if err := s.engine.Set(en, transform); err != nil {
			return err
		}
	}
	return nil
}
