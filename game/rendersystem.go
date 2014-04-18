package game

import (
	"time"

	"github.com/der-antikeks/gisp/ecs"

	"github.com/go-gl/gl"
)

type RenderSystem struct {
	wm       *WindowManager
	drawable *ecs.Collection
}

func NewRenderSystem(wm *WindowManager) ecs.System {
	return &RenderSystem{
		wm: wm,
	}
}

func (s *RenderSystem) AddedToEngine(e *ecs.Engine) error {
	s.drawable = e.Collection(PositionType, MeshType, ColorType)
	return nil
}
func (s *RenderSystem) RemovedFromEngine(e *ecs.Engine) error {
	return nil
}
func (s *RenderSystem) Update(delta time.Duration) error {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.LoadIdentity()

	for _, e := range s.drawable.Entities() {
		p := e.Get(PositionType).(*PositionComponent)
		m := e.Get(MeshType).(*MeshComponent)
		c := e.Get(ColorType).(*ColorComponent)

		//fmt.Println("rendering", e.Name, "at", p)

		gl.LoadIdentity()
		gl.Translated(p.Position.X, p.Position.Y, 0)
		gl.Rotated(p.Rotation-90, 0, 0, 1)
		gl.Color3d(c.R, c.G, c.B)

		gl.Begin(gl.LINE_LOOP)
		for _, point := range m.Points {
			gl.Vertex3d(point.X, point.Y, 0)
		}
		gl.End()
	}

	s.wm.Update()
	return nil
}
