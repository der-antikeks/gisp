package game

import (
	"log"
	"sync"

	"github.com/der-antikeks/mathgl/mgl32"
)

type uiSystem struct {
	messages chan interface{}
	camera   Entity

	doConc chan func()
	doDone chan struct{}
}

var (
	uiInstance *uiSystem
	uiOnce     sync.Once
)

func UiSystem() *uiSystem {
	uiOnce.Do(func() {
		uiInstance = &uiSystem{
			messages: make(chan interface{}),

			doConc: make(chan func()),
			doDone: make(chan struct{}),
		}

		go func() {
			uiInstance.Restart()

			var width, height float64 // := GlContextSystem(nil).Size()
			var mx, my float64

			for {
				select {
				case event := <-uiInstance.messages:
					switch e := event.(type) {
					case MessageMouseButton:
						if GlContextSystem(nil).IsMouseDown(MouseLeft) {
							// get window coordinates
							mx, my = GlContextSystem(nil).MousePos()
							win := mgl32.Vec4{
								float32((mx/width)*2 - 1),
								float32(-(my/height)*2 + 1),
								1,
								1,
							}

							// get camera components
							if uiInstance.camera == NoEntity {
								log.Println("no camera for ui-system")
								continue
							}

							ec, err := EntitySystem().Get(uiInstance.camera, ProjectionType)
							if err != nil {
								log.Fatal(err)

							}
							p := ec.(Projection)

							ec, err = EntitySystem().Get(uiInstance.camera, TransformationType)
							if err != nil {
								log.Fatal(err)
							}
							t := ec.(Transformation)

							ec, err = EntitySystem().Get(uiInstance.camera, SceneType)
							if err != nil {
								log.Fatal(err)
							}
							sc := ec.(Scene).Name

							// unproject to object coordinates
							obj := unprojectVector(win, t.MatrixWorld(), p.Matrix)
							origin := t.MatrixWorld().Mul4x1(mgl32.Vec4{0, 0, 0, 1})
							direction := obj.Sub(origin).Normalize()

							// get intersections
							intersections := SpatialSystem().IntersectsRay(sc, origin, direction)
							for _, e := range intersections {
								ec, err := EntitySystem().Get(e, MaterialType)
								if err != nil {
									log.Println("could not get material of entity:", e, err)
									// TODO
									continue
								}
								mat := ec.(Material)
								mat.Set("diffuse", mgl32.Vec3{1, 0, 0})

								if err := EntitySystem().Set(e, mat); err != nil {
									log.Fatal("could not set color of entity:", e, err)
								}
							}
						}

					case MessageResize:
						width = float64(e.Width)
						height = float64(e.Height)

					}

				case f := <-uiInstance.doConc:
					f()
					uiInstance.doDone <- struct{}{}
				}
			}
		}()
	})

	return uiInstance
}

func (s *uiSystem) Restart() {
	GlContextSystem(nil).OnMouseButton().Subscribe(s.messages, PriorityBeforeRender)
	GlContextSystem(nil).OnResize().Subscribe(s.messages, PriorityBeforeRender)
}

func (s *uiSystem) Stop() {
	GlContextSystem(nil).OnMouseButton().Unsubscribe(s.messages)
	GlContextSystem(nil).OnResize().Unsubscribe(s.messages)

	s.camera = NoEntity
}

// run function on main thread
func (s *uiSystem) do(f func()) {
	s.doConc <- f
	<-s.doDone
}

func (s *uiSystem) SetCamera(c Entity) {
	s.do(func() {
		s.camera = c
	})
}

// www.opengl.org/wiki/GluProject_and_gluUnProject_code
func unprojectVector(win mgl32.Vec4, modelview, projection mgl32.Mat4) (obj mgl32.Vec4) {
	m := modelview.Mul4(projection.Inv())

	// perspective divide
	d := 1 / (m[3]*win[0] + m[7]*win[1] + m[11]*win[2] + m[15])

	return mgl32.Vec4{
		(m[0]*win[0] + m[4]*win[1] + m[8]*win[2] + m[12]) * d,
		(m[1]*win[0] + m[5]*win[1] + m[9]*win[2] + m[13]) * d,
		(m[2]*win[0] + m[6]*win[1] + m[10]*win[2] + m[14]) * d,
		1,
	}
}
