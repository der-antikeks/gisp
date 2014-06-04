package game

import (
	"log"
	"math"
	"time"

	"github.com/der-antikeks/mathgl/mgl32"
)

/*
	handle game-state, start loading/unloading entities, send update messages

	state string
	Update(delta time.Duration)
	SubscribeOnUpdate(chan time.Duration, prio int)
*/
type GameStateSystem struct {
	context *GlContextSystem
	ents    *EntitySystem

	state string
	since time.Time

	messages chan interface{}
	timer    *time.Timer

	quit, update *Observer
}

func NewGameStateSystem(context *GlContextSystem, ents *EntitySystem) *GameStateSystem {
	s := &GameStateSystem{
		context: context,
		ents:    ents,

		messages: make(chan interface{}),

		quit:   NewObserver(),
		update: NewObserver(),
	}

	/*
		go func() {
			s.Restart()

			for event := range s.messages {
				switch event.(type) {
				case MessageKey,
					MessageTimeout:

					if err := s.Update(); err != nil {
						log.Fatal("could not update game state:", err)
					}
				}
			}
		}()
	*/

	return s
}

func (s *GameStateSystem) Restart() {
	//s.update.Subscribe(s.messages, PriorityBeforeRender)
}

func (s *GameStateSystem) Stop() {
	s.context.OnKey().Unsubscribe(s.messages)
	//s.update.Unsubscribe(s.messages)

	s.state = ""
}

func (s *GameStateSystem) init() {
	log.Println("initialize")

	w, h := s.context.Size()
	aspect := float32(w) / float32(h) // 4.0 / 3.0
	// TODO: update aspect after wm.onResize

	var size float32 = 10.0
	c := s.ents.CreateOrthographicCamera(-size, size, size/aspect, -size/aspect, 1, 100)

	ec, err := s.ents.Get(c, TransformationType)
	if err != nil {
		log.Fatal("could not get transform of camera:", err)
	}
	t := ec.(Transformation)

	t.Position = mgl32.Vec3{0, 10, 0}
	t.Rotation = mgl32.QuatLookAtV(t.Position, mgl32.Vec3{0, 0, 0}, t.Up)

	if err := s.ents.Set(c, t); err != nil {
		log.Fatal("could not move camera:", err)
	}
}

func (s *GameStateSystem) Update(delta time.Duration) {
	if s.context.IsKeyDown(KeyEscape) {
		log.Println("closing")

		s.context.Close()
		s.quit.Publish(MessageQuit{})
		// TODO: later replace with quit screen, closing initialized by gui-system
		return
	}

	switch s.state {
	case "":
		s.init()
		s.state = "init"
		s.since = time.Now()

	case "init":
		log.Println("create splash screen")

		s.ents.CreateSplashScreen()
		s.state = "splash"
		s.since = time.Now()

		/*
			if s.timer != nil {
				s.timer.Stop()
			}
			s.timer = time.AfterFunc(5*time.Second+1, func() {
				log.Println("timeout!")
				s.messages <- MessageTimeout(s.since)
			})
		*/

	case "splash":
		if time.Now().After(s.since.Add(5*time.Second)) || s.context.AnyKeyDown() {
			log.Println("create main menu")

			//s.timer.Stop()

			// late key message subscription
			s.context.OnKey().Subscribe(s.messages, PriorityBeforeRender)

			/*
				for _, e := range s.ents.Query() {
					if e == s.state {
						continue
					}
					s.ents.Delete(e) // ignoring errors
				}
			*/

			s.ents.CreateMainMenu()

			w, h := s.context.Size()
			c := s.ents.CreatePerspectiveCamera(45.0, float32(w)/float32(h), 0.1, 100.0)
			s.ents.Set(c,
				OrbitControl{
					MovementSpeed: 1.0,
					RotationSpeed: 0.01,
					ZoomSpeed:     1.0,

					Min:    5.0,
					Max:    math.Inf(1),
					Target: Entity(0), // TODO: proper target setting
				},
			)

			s.state = "mainmenu"
			s.since = time.Now()
		}

	case "mainmenu":
		if s.context.IsKeyDown(KeyEnter) {
			log.Println("starting game")

			s.state = "playing"
			s.since = time.Now()
		}

	case "optionsmenu":
		if s.context.IsKeyDown(KeyEscape) {
			log.Println("back to main menu")

			s.state = "mainmenu"
			s.since = time.Now()
		}

	case "playing":
		if s.context.IsKeyDown(KeyPause) {
			log.Println("pausing")

			s.state = "pause"
			s.since = time.Now()
		}

	case "pause":
		if !s.context.IsKeyDown(KeyPause) && s.context.AnyKeyDown() {
			log.Println("restarting")

			s.state = "playing"
			s.since = time.Now()
		}
	}

	s.update.Publish(MessageUpdate{Delta: delta})
	return
}

func (s *GameStateSystem) OnQuit() *Observer   { return s.quit }
func (s *GameStateSystem) OnUpdate() *Observer { return s.update }
