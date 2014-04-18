package game

import (
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

type RenderSystem struct {
	wm *WindowManager
}

func NewRenderSystem(wm *WindowManager) ecs.System {
	return &RenderSystem{
		wm: wm,
	}
}

func (s *RenderSystem) AddedToEngine(e *ecs.Engine) error {
	return nil
}
func (s *RenderSystem) RemovedFromEngine(e *ecs.Engine) error {
	return nil
}
func (s *RenderSystem) Update(delta time.Duration) error {
	s.wm.update()
	return nil
}
