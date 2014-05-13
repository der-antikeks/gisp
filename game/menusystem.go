package game

import (
	"log"
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

type MenuSystem struct {
	engine *ecs.Engine
	prio   ecs.SystemPriority
	im     *InputManager

	messages chan ecs.Message

	buttons   []ecs.Entity
	gamestate ecs.Entity
}

func NewMenuSystem(engine *ecs.Engine, im *InputManager) *MenuSystem {
	s := &MenuSystem{
		engine:    engine,
		prio:      PriorityBeforeRender,
		im:        im,
		messages:  make(chan ecs.Message),
		gamestate: -1,
	}

	go func() {
		s.Restart()

		for event := range s.messages {
			switch e := event.(type) {
			default:
			case ecs.MessageUpdate:
				if err := s.Update(e.Delta); err != nil {
					log.Fatal("could not update menu:", err)
				}
			}
		}
	}()

	return s
}

func (s *MenuSystem) Restart() {
	s.engine.Subscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, s.prio, s.messages)

	s.engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{MenuType, PositionType, MeshType},
	}, s.prio, s.messages)

	s.engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{GameStateType},
	}, s.prio, s.messages)
}

func (s *MenuSystem) Stop() {
	s.engine.Unsubscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, s.messages)

	s.engine.Unsubscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{MenuType, PositionType, MeshType},
	}, s.messages)

	s.engine.Unsubscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{GameStateType},
	}, s.messages)

	s.buttons = []ecs.Entity{}
	s.gamestate = -1
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
