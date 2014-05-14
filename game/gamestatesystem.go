package game

import (
	"log"
	"time"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"
)

type GameStateSystem struct {
	engine *ecs.Engine
	prio   ecs.SystemPriority
	em     *EntityManager
	im     *InputManager
	wm     *WindowManager

	messages    chan ecs.Message
	state       ecs.Entity
	initialized bool
	timer       *time.Timer
}

func NewGameStateSystem(engine *ecs.Engine, em *EntityManager, im *InputManager, wm *WindowManager) *GameStateSystem {
	s := &GameStateSystem{
		engine:   engine,
		prio:     PriorityBeforeRender,
		em:       em,
		im:       im,
		wm:       wm,
		messages: make(chan ecs.Message),
		state:    -1,
	}

	go func() {
		s.Restart()
		s.init()

		for event := range s.messages {
			switch e := event.(type) {
			case ecs.MessageEntityAdd:
				s.state = e.Added

				if err := s.Update(); err != nil {
					log.Fatal("could not update game state:", err)
				}
			case ecs.MessageEntityRemove:
				if s.state == e.Removed {
					s.state = -1
				}

			case ecs.MessageEntityUpdate,
				MessageKey,
				MessageTimeout:

				if err := s.Update(); err != nil {
					log.Fatal("could not update game state:", err)
				}
			}
		}
	}()

	return s
}

func (s *GameStateSystem) Restart() {
	s.engine.Subscribe(ecs.Filter{
		Types: []ecs.MessageType{TimeoutMessageType},
	}, s.prio, s.messages)

	s.engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityUpdateMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{GameStateType},
	}, s.prio, s.messages)
}

func (s *GameStateSystem) Stop() {
	s.engine.Unsubscribe(ecs.Filter{
		Types: []ecs.MessageType{KeyMessageType, TimeoutMessageType},
	}, s.messages)

	s.engine.Unsubscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityUpdateMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{GameStateType},
	}, s.messages)

	s.state = -1
}

func (s *GameStateSystem) init() {
	if s.initialized || s.state != -1 {
		return
	}

	log.Println("initialize")
	s.em.Initalize()

	w, h := s.wm.Size()
	aspect := float64(w) / float64(h) // 4.0 / 3.0
	// TODO: update aspect after wm.onResize

	size := 10.0
	c := s.em.CreateOrthographicCamera(-size, size, size/aspect, -size/aspect, 1, 100)

	ec, err := s.engine.Get(c, TransformationType)
	if err != nil {
		log.Fatal("could not get transform of camera:", err)
	}
	t := ec.(Transformation)

	t.Position = math.Vector{0, 10, 0}
	t.Rotation = math.QuaternionFromRotationMatrix(math.LookAt(t.Position, math.Vector{0, 0, 0}, t.Up))

	if err := s.engine.Set(c, t); err != nil {
		log.Fatal("could not move camera:", err)
	}

	s.initialized = true
}

func (s *GameStateSystem) Update() error {
	if s.im.IsKeyDown(KeyEscape) {
		log.Println("closing")

		s.wm.Close()
		running = false
		// TODO: later replace with quit screen, closing initialized by gui-system
		return nil
	}

	if s.state == -1 {
		return nil
	}

	ec, err := s.engine.Get(s.state, GameStateType)
	if err != nil {
		return err
	}
	se := ec.(GameStateComponent)
	var update bool

	switch se.State {
	case "init":
		log.Println("create splash screen")

		s.em.CreateSplashScreen()
		se.State = "splash"
		se.Since = time.Now()
		update = true

		if s.timer != nil {
			s.timer.Stop()
		}
		s.timer = time.AfterFunc(5*time.Second+1, func() {
			log.Println("timeout!")
			s.messages <- MessageTimeout(se.Since)
		})

	case "splash":
		if time.Now().After(se.Since.Add(5*time.Second)) || s.im.AnyKeyDown() {
			log.Println("create main menu")

			s.timer.Stop()

			// late key message subscription
			s.engine.Subscribe(ecs.Filter{
				Types: []ecs.MessageType{KeyMessageType, TimeoutMessageType},
			}, s.prio, s.messages)

			/*
				for _, e := range s.engine.Query() {
					if e == s.state {
						continue
					}
					s.engine.Delete(e) // ignoring errors
				}
			*/

			s.em.CreateMainMenu()

			w, h := s.wm.Size()
			s.em.CreatePerspectiveCamera(45.0, float64(w)/float64(h), 0.1, 100.0)

			se.State = "mainmenu"
			se.Since = time.Now()
			update = true
		}

	case "mainmenu":
		if s.im.IsKeyDown(KeyEnter) {
			log.Println("starting game")

			se.State = "playing"
			se.Since = time.Now()
			update = true
		}

	case "optionsmenu":
		if s.im.IsKeyDown(KeyEscape) {
			log.Println("back to main menu")

			se.State = "mainmenu"
			se.Since = time.Now()
			update = true
		}

	case "playing":
		if s.im.IsKeyDown(KeyPause) {
			log.Println("pausing")

			se.State = "pause"
			se.Since = time.Now()
			update = true
		}

	case "pause":
		if !s.im.IsKeyDown(KeyPause) && s.im.AnyKeyDown() {
			log.Println("restarting")

			se.State = "playing"
			se.Since = time.Now()
			update = true
		}
	}

	if update {
		if err := s.engine.Set(s.state, se); err != nil {
			return err
		}
	}

	return nil
}
