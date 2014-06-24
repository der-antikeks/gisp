package game

import (
	"log"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
)

// change entities based on controller input
type orbitControlSystem struct {
	messages    chan interface{}
	controlable []Entity
}

var (
	controlInstance *orbitControlSystem
	controlOnce     sync.Once
)

func ControlSystem() *orbitControlSystem {
	controlOnce.Do(func() {
		controlInstance = &orbitControlSystem{
			messages:    make(chan interface{}),
			controlable: []Entity{},
		}

		go func() {
			controlInstance.Restart()

			var dragging bool
			var oldx, oldy float64
			var deltax, deltay, deltaz float64
			/*
				TODO: initial value
				var width, height float64

				GlContextSystem(nil).width
				GlContextSystem(nil).height
			*/

			for event := range controlInstance.messages {
				switch e := event.(type) {
				case MessageEntityAdd:
					controlInstance.controlable = append(controlInstance.controlable, e.Added)
				case MessageEntityRemove:
					for i, f := range controlInstance.controlable {
						if f == e.Removed {
							controlInstance.controlable = append(controlInstance.controlable[:i], controlInstance.controlable[i+1:]...)
							break
						}
					}

				case MessageMouseButton:
					if !dragging && GlContextSystem(nil).IsMouseDown(MouseRight) {
						dragging = true
						oldx, oldy = GlContextSystem(nil).MousePos()
					} else if dragging && !GlContextSystem(nil).IsMouseDown(MouseRight) {
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
						x, y := GlContextSystem(nil).MousePos()
						deltax, deltay = x-oldx, y-oldy // /width, /height
						oldx, oldy = x, y
					}

					if deltax != 0 || deltay != 0 || deltaz != 0 {
						if err := controlInstance.Update(deltax, deltay, deltaz); err != nil {
							log.Fatal("could not update game state:", err)
						}
						deltaz = 0
					}
				}
			}
		}()
	})

	return controlInstance
}

func (s *orbitControlSystem) Restart() {
	GameStateSystem().OnUpdate().Subscribe(s.messages, PriorityBeforeRender)

	GlContextSystem(nil).OnMouseButton().Subscribe(s.messages, PriorityBeforeRender)
	GlContextSystem(nil).OnMouseScroll().Subscribe(s.messages, PriorityBeforeRender)
	GlContextSystem(nil).OnResize().Subscribe(s.messages, PriorityBeforeRender)

	EntitySystem().OnAdd(TransformationType, OrbitControlType).Subscribe(s.messages, PriorityBeforeRender)
	EntitySystem().OnRemove(TransformationType, OrbitControlType).Subscribe(s.messages, PriorityBeforeRender)
}

func (s *orbitControlSystem) Stop() {
	GameStateSystem().OnUpdate().Unsubscribe(s.messages)

	GlContextSystem(nil).OnMouseButton().Unsubscribe(s.messages)
	GlContextSystem(nil).OnMouseScroll().Unsubscribe(s.messages)
	GlContextSystem(nil).OnResize().Unsubscribe(s.messages)

	EntitySystem().OnAdd(TransformationType, OrbitControlType).Unsubscribe(s.messages)
	EntitySystem().OnRemove(TransformationType, OrbitControlType).Unsubscribe(s.messages)

	s.controlable = s.controlable[:0]
}

func (s *orbitControlSystem) Update(deltax, deltay, deltaz float64) error {
	for _, en := range s.controlable {

		ec, err := EntitySystem().Get(en, TransformationType)
		if err != nil {
			return err
		}
		transform := ec.(Transformation)

		ec, err = EntitySystem().Get(en, OrbitControlType)
		if err != nil {
			return err
		}
		control := ec.(OrbitControl)

		var target mgl32.Vec3
		// TODO: if no target is set return error
		if control.Target != 0 {
			ec, err = EntitySystem().Get(control.Target, TransformationType)
			if err != nil {
				return err
			}
			target = ec.(Transformation).Position
		}

		// TODO: exponential zoom?
		distance := mgl32.Clamp(
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

		if err := EntitySystem().Set(en, transform); err != nil {
			return err
		}
	}
	return nil
}
