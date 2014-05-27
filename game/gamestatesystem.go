package game

import (
	"log"
	m "math"
	"time"

	"github.com/der-antikeks/gisp/math"
)

type GameStateSystem struct {
	engine *Engine
	prio   SystemPriority
	em     *EntityManager
	im     *InputManager
	wm     *WindowManager

	messages    chan Message
	state       Entity
	timer       *time.Timer
	initialized bool
	quit        chan struct{}
}

func NewGameStateSystem(engine *Engine, em *EntityManager, im *InputManager, wm *WindowManager, quit chan struct{}) *GameStateSystem {
	s := &GameStateSystem{
		engine:   engine,
		prio:     PriorityBeforeRender,
		em:       em,
		im:       im,
		wm:       wm,
		messages: make(chan Message),
		state:    NoEntity,
		quit:     quit,
	}

	go func() {
		s.Restart()
		s.init()

		for event := range s.messages {
			switch e := event.(type) {
			case MessageEntityAdd:
				s.state = e.Added

				if err := s.Update(); err != nil {
					log.Fatal("could not update game state:", err)
				}
			case MessageEntityRemove:
				if s.state == e.Removed {
					s.state = NoEntity
				}

			case MessageEntityUpdate,
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
	s.engine.Subscribe(Filter{
		Types: []MessageType{TimeoutMessageType},
	}, s.prio, s.messages)

	s.engine.Subscribe(Filter{
		Types:  []MessageType{EntityAddMessageType, EntityUpdateMessageType, EntityRemoveMessageType},
		Aspect: []ComponentType{GameStateType},
	}, s.prio, s.messages)
}

func (s *GameStateSystem) Stop() {
	s.engine.Unsubscribe(Filter{
		Types: []MessageType{KeyMessageType, TimeoutMessageType},
	}, s.messages)

	s.engine.Unsubscribe(Filter{
		Types:  []MessageType{EntityAddMessageType, EntityUpdateMessageType, EntityRemoveMessageType},
		Aspect: []ComponentType{GameStateType},
	}, s.messages)

	s.state = NoEntity
}

func (s *GameStateSystem) init() {
	if s.initialized || s.state != NoEntity {
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
	t.Rotation = math.QuaternionLookAt(t.Position, math.Vector{0, 0, 0}, t.Up)

	if err := s.engine.Set(c, t); err != nil {
		log.Fatal("could not move camera:", err)
	}

	s.initialized = true
}

func (s *GameStateSystem) Update() error {
	if s.im.IsKeyDown(KeyEscape) {
		log.Println("closing")

		s.wm.Close()
		close(s.quit)
		// TODO: later replace with quit screen, closing initialized by gui-system
		return nil
	}

	if s.state == NoEntity {
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
			s.engine.Subscribe(Filter{
				Types: []MessageType{KeyMessageType, TimeoutMessageType},
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
			c := s.em.CreatePerspectiveCamera(45.0, float64(w)/float64(h), 0.1, 100.0)
			s.engine.Set(c,
				OrbitControl{
					MovementSpeed: 1.0,
					RotationSpeed: 0.01,
					ZoomSpeed:     1.0,

					Min:    5.0,
					Max:    m.Inf(1),
					Target: Entity(0), // TODO: proper target setting
				},
			)

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
