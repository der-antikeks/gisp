package game

import (
	"log"
	"sync"
	"time"

	"github.com/der-antikeks/mathgl/mgl32"
)

// move entities with velocity
type movementSystem struct {
	messages chan interface{}
	moveable []Entity
}

var (
	moveInstance *movementSystem
	moveOnce     sync.Once
)

func MovementSystem() *movementSystem {
	moveOnce.Do(func() {
		moveInstance = &movementSystem{
			messages: make(chan interface{}),
		}

		go func() {
			moveInstance.Restart()

			for event := range moveInstance.messages {
				switch e := event.(type) {
				case MessageEntityAdd:
					moveInstance.moveable = append(moveInstance.moveable, e.Added)
				case MessageEntityRemove:
					for i, f := range moveInstance.moveable {
						if f == e.Removed {
							moveInstance.moveable = append(moveInstance.moveable[:i], moveInstance.moveable[i+1:]...)
							break
						}
					}

				case MessageUpdate:
					if err := moveInstance.Update(e.Delta); err != nil {
						log.Fatal("could not update movement:", err)
					}
				}
			}
		}()
	})

	return moveInstance
}

func (s *movementSystem) Restart() {
	GameStateSystem().OnUpdate().Subscribe(s.messages, PriorityBeforeRender)

	EntitySystem().OnAdd(TransformationType, VelocityType).Subscribe(s.messages, PriorityBeforeRender)
	EntitySystem().OnRemove(TransformationType, VelocityType).Subscribe(s.messages, PriorityBeforeRender)
}

func (s *movementSystem) Stop() {
	GameStateSystem().OnUpdate().Unsubscribe(s.messages)

	EntitySystem().OnAdd(TransformationType, VelocityType).Unsubscribe(s.messages)
	EntitySystem().OnRemove(TransformationType, VelocityType).Unsubscribe(s.messages)

	s.moveable = []Entity{}
}

func (s *movementSystem) Update(delta time.Duration) error {
	for _, en := range s.moveable {
		ec, err := EntitySystem().Get(en, TransformationType)
		if err != nil {
			return err
		}
		transform := ec.(Transformation)

		ec, err = EntitySystem().Get(en, VelocityType)
		if err != nil {
			return err
		}
		velocity := ec.(Velocity)

		var update bool
		if v := velocity.Velocity; !v.ApproxEqual(mgl32.Vec3{}) {
			update = true
			transform.Position = transform.Position.Add(v.Mul(float32(delta.Seconds())))
		}

		if a := velocity.Angular; !a.ApproxEqual(mgl32.Vec3{}) {
			update = true

			// http://www.euclideanspace.com/physics/kinematics/angularvelocity/#quaternion
			// dq/dt = 1/2 w(t) q(t)

			q := transform.Rotation.Normalize()
			a = a.Mul(float32(delta.Seconds()))
			w := mgl32.Quat{0, mgl32.Vec3{a[0], a[1], a[2]}}
			transform.Rotation = transform.Rotation.Add(w.Mul(q).Scale(0.5)).Normalize()
		}

		if update {
			transform.matrix = Compose(transform.Position, transform.Rotation, transform.Scale)
			transform.updatedMatrix = true

			if err := EntitySystem().Set(en, transform); err != nil {
				return err
			}
		}
	}
	return nil
}
