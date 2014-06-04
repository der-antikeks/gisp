package game

import (
	//	"log"
	"time"
)

type MenuSystem struct {
	/*
		engine *Engine
		prio   Priority
		im     *InputManager

		messages chan Message

		buttons   []Entity
		gamestate Entity
	*/
}

func NewMenuSystem( /*engine *Engine, im *InputManager*/) /* *MenuSystem */ {
	/*
		s := &MenuSystem{
			engine:    engine,
			prio:      PriorityBeforeRender,
			im:        im,
			messages:  make(chan Message),
			gamestate: NoEntity,
		}

		go func() {
			s.Restart()

			for event := range s.messages {
				switch e := event.(type) {
				default:
				case MessageUpdate:
					if err := s.Update(e.Delta); err != nil {
						log.Fatal("could not update menu:", err)
					}
				}
			}
		}()

		return s
	*/
}

func (s *MenuSystem) Restart() {
	/*
		s.engine.Subscribe(Filter{
			Types: []MessageType{UpdateMessageType},
		}, s.prio, s.messages)

		s.engine.Subscribe(Filter{
			Types:  []MessageType{EntityAddMessageType, EntityRemoveMessageType},
			Aspect: []ComponentType{MenuType, PositionType, MeshType},
		}, s.prio, s.messages)

		s.engine.Subscribe(Filter{
			Types:  []MessageType{EntityAddMessageType, EntityRemoveMessageType},
			Aspect: []ComponentType{GameStateType},
		}, s.prio, s.messages)
	*/
}

func (s *MenuSystem) Stop() {
	/*
		s.engine.Unsubscribe(Filter{
			Types: []MessageType{UpdateMessageType},
		}, s.messages)

		s.engine.Unsubscribe(Filter{
			Types:  []MessageType{EntityAddMessageType, EntityRemoveMessageType},
			Aspect: []ComponentType{MenuType, PositionType, MeshType},
		}, s.messages)

		s.engine.Unsubscribe(Filter{
			Types:  []MessageType{EntityAddMessageType, EntityRemoveMessageType},
			Aspect: []ComponentType{GameStateType},
		}, s.messages)

		s.buttons = []Entity{}
		s.gamestate = NoEntity
	*/
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
