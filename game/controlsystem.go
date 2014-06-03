package game

import (
	"log"

	"github.com/der-antikeks/mathgl/mgl32"
)

// change entities based on controller input
type OrbitControlSystem struct {
	context *GlContextSystem
	ents    *EntitySystem
	state   *GameStateSystem

	messages    chan interface{}
	controlable []Entity
}

func NewControlSystem(context *GlContextSystem, ents *EntitySystem, state *GameStateSystem) *OrbitControlSystem {
	s := &OrbitControlSystem{
		context: context,
		ents:    ents,
		state:   state,

		messages:    make(chan interface{}),
		controlable: []Entity{},
	}

	go func() {
		s.Restart()

		var dragging bool
		var oldx, oldy float64
		var deltax, deltay, deltaz float64
		/*
			TODO: initial value
			var width, height float64

			s.context.width
			s.context.height
		*/

		for event := range s.messages {
			switch e := event.(type) {
			case MessageEntityAdd:
				s.controlable = append(s.controlable, e.Added)
			case MessageEntityRemove:
				for i, f := range s.controlable {
					if f == e.Removed {
						s.controlable = append(s.controlable[:i], s.controlable[i+1:]...)
						break
					}
				}

			case MessageMouseButton:
				if !dragging && s.context.IsMouseDown(MouseRight) {
					dragging = true
					oldx, oldy = s.context.MousePos()
				} else if dragging && !s.context.IsMouseDown(MouseRight) {
					dragging = false
					deltax, deltay = 0, 0
				}

			case MessageMouseScroll:
				deltaz -= float64(e)

			case MessageResize:
				/*
					width = float64(e.Width)
					height = float64(e.Height)
				*/

			case MessageUpdate:
				if dragging {
					x, y := s.context.MousePos()
					deltax, deltay = x-oldx, y-oldy // /width, /height
					oldx, oldy = x, y
				}

				if deltax != 0 || deltay != 0 || deltaz != 0 {
					if err := s.Update(deltax, deltay, deltaz); err != nil {
						log.Fatal("could not update game state:", err)
					}
					deltaz = 0
				}
			}
		}
	}()

	return s
}

func (s *OrbitControlSystem) Restart() {
	s.state.OnUpdate().Subscribe(s.messages, PriorityBeforeRender)

	s.context.OnMouseButton().Subscribe(s.messages, PriorityBeforeRender)
	s.context.OnMouseScroll().Subscribe(s.messages, PriorityBeforeRender)
	s.context.OnResize().Subscribe(s.messages, PriorityBeforeRender)

	s.ents.OnAdd(TransformationType, OrbitControlType).Subscribe(s.messages, PriorityBeforeRender)
	s.ents.OnRemove(TransformationType, OrbitControlType).Subscribe(s.messages, PriorityBeforeRender)
}

func (s *OrbitControlSystem) Stop() {
	s.state.OnUpdate().Unsubscribe(s.messages)

	s.context.OnMouseButton().Unsubscribe(s.messages)
	s.context.OnMouseScroll().Unsubscribe(s.messages)
	s.context.OnResize().Unsubscribe(s.messages)

	s.ents.OnAdd(TransformationType, OrbitControlType).Unsubscribe(s.messages)
	s.ents.OnRemove(TransformationType, OrbitControlType).Unsubscribe(s.messages)

	s.controlable = s.controlable[:0]
}

func (s *OrbitControlSystem) Update(deltax, deltay, deltaz float64) error {
	for _, en := range s.controlable {

		ec, err := s.ents.Get(en, TransformationType)
		if err != nil {
			return err
		}
		transform := ec.(Transformation)

		ec, err = s.ents.Get(en, OrbitControlType)
		if err != nil {
			return err
		}
		control := ec.(OrbitControl)

		var target mgl32.Vec3
		// TODO: if no target is set return error
		if control.Target != 0 {
			ec, err = s.ents.Get(control.Target, TransformationType)
			if err != nil {
				return err
			}
			target = ec.(Transformation).Position
		}

		// TODO: exponential zoom?
		distance := mgl32.Clampf(
			transform.Position.Sub(target).Len()+float32(deltaz*control.ZoomSpeed),
			float32(control.Min), float32(control.Max))

		delta := mgl32.AnglesToQuat(
			float32(deltay*control.RotationSpeed),
			float32(deltax*control.RotationSpeed),
			0,
			mgl32.XYZ).Inverse()

		transform.Rotation = transform.Rotation.Mul(delta)
		transform.Position = transform.Rotation.Rotate(mgl32.Vec3{
			0,
			0,
			distance,
		}).Add(target)

		if err := s.ents.Set(en, transform); err != nil {
			return err
		}
	}
	return nil
}
