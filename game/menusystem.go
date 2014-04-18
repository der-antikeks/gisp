package main

import (
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

type MenuSystem struct {
	im *InputManager

	buttons   *ecs.Collection
	gamestate *ecs.Collection
}

func NewMenuSystem(im *InputManager) ecs.System {
	return &MenuSystem{
		im: im,
	}
}

func (s *MenuSystem) AddedToEngine(e *ecs.Engine) error {
	/*
		s.buttons = e.Collection(MenuType, PositionType, MeshType)
		s.gamestate = e.Collection(GameStateType)
	*/
	return nil
}

func (s *MenuSystem) RemovedFromEngine(*ecs.Engine) error {
	/*
		s.buttons = nil
		s.gamestate = nil
	*/
	return nil
}

func (s *MenuSystem) Update(delta time.Duration) error {
	/*
		state := s.gamestate.First().Get(GameStateType).(*GameStateComponent)

		for _, en := range s.buttons.Entities() {
			x, y := s.im.MousePos()
			// TODO: cache old position & compare

			m := en.Get(MenuType).(*MenuComponent)
			p := en.Get(PositionType).(*PositionComponent).Position
			c := en.Get(MeshType).(*MeshComponent).Collider
			mp := Point{(x - 320) - p.X, (240 - y) - p.Y}

			if c.ContainsPoint(mp) {
				if s.im.IsMouseClick(glfw.MouseButton1) {
					log.Println("clicked", m.Name)

					switch m.Name {
					case "Exit":
						state.State = "exit"
					case "LoadSubmenu":
						state.State = "load-submenu"
					case "ReturnMenu":
						state.State = "load-menu"
					}
				}

				if s.im.IsMouseDown(glfw.MouseButton1) {
					en.ChangeState("pressed")
				} else {
					en.ChangeState("hover")
				}
			} else {
				en.ChangeState("normal")
			}
		}
	*/
	return nil
}
