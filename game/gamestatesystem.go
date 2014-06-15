package game

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/der-antikeks/mathgl/mgl32"
)

/*
	handle game-state, start loading/unloading entities, send update messages

	state string
	Update(delta time.Duration)
	SubscribeOnUpdate(chan time.Duration, prio int)
*/
type gameStateSystem struct {
	state string
	since time.Time

	messages chan interface{}
	timer    *time.Timer

	quit, update *Observer
}

var (
	stateInstance *gameStateSystem
	stateOnce     sync.Once
)

func GameStateSystem() *gameStateSystem {
	stateOnce.Do(func() {
		stateInstance = &gameStateSystem{
			messages: make(chan interface{}),

			quit:   NewObserver(),
			update: NewObserver(),
		}

		/*
			go func() {
				stateInstance.Restart()

				for event := range stateInstance.messages {
					switch event.(type) {
					case MessageKey,
						MessageTimeout:

						if err := stateInstance.Update(); err != nil {
							log.Fatal("could not update game state:", err)
						}
					}
				}
			}()
		*/
	})

	return stateInstance
}

func (s *gameStateSystem) Restart() {
	//s.update.Subscribe(s.messages, PriorityBeforeRender)
}

func (s *gameStateSystem) Stop() {
	GlContextSystem(nil).OnKey().Unsubscribe(s.messages)
	//s.update.Unsubscribe(s.messages)

	s.state = ""
}

func (s *gameStateSystem) init() {
	log.Println("initialize")

	w, h := GlContextSystem(nil).Size()
	aspect := float32(w) / float32(h) // 4.0 / 3.0
	// TODO: update aspect after wm.onResize

	var size float32 = 10.0
	c := EntitySystem().CreateOrthographicCamera(-size, size, size/aspect, -size/aspect, 1, 100)

	ec, err := EntitySystem().Get(c, TransformationType)
	if err != nil {
		log.Fatal("could not get transform of camera:", err)
	}
	t := ec.(Transformation)

	t.Position = mgl32.Vec3{0, 10, 0}
	t.Rotation = mgl32.QuatLookAtV(t.Position, mgl32.Vec3{0, 0, 0}, t.Up)

	if err := EntitySystem().Set(c, t); err != nil {
		log.Fatal("could not move camera:", err)
	}
}

func (s *gameStateSystem) Update(delta time.Duration) {
	if GlContextSystem(nil).IsKeyDown(KeyEscape) {
		log.Println("closing")

		GlContextSystem(nil).Close()
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

		EntitySystem().CreateSplashScreen()
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
		if time.Now().After(s.since.Add(5*time.Second)) || GlContextSystem(nil).AnyKeyDown() {
			log.Println("create main menu")

			//s.timer.Stop()

			// late key message subscription
			GlContextSystem(nil).OnKey().Subscribe(s.messages, PriorityBeforeRender)

			/*
				for _, e := range EntitySystem().Query() {
					if e == s.state {
						continue
					}
					EntitySystem().Delete(e) // ignoring errors
				}
			*/

			EntitySystem().CreateMainMenu()

			w, h := GlContextSystem(nil).Size()
			c := EntitySystem().CreatePerspectiveCamera(45.0, float32(w)/float32(h), 0.1, 200.0)
			EntitySystem().Set(c,
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
		if GlContextSystem(nil).IsKeyDown(KeyEnter) {
			log.Println("starting game")

			s.state = "playing"
			s.since = time.Now()
		}

	case "optionsmenu":
		if GlContextSystem(nil).IsKeyDown(KeyEscape) {
			log.Println("back to main menu")

			s.state = "mainmenu"
			s.since = time.Now()
		}

	case "playing":
		if GlContextSystem(nil).IsKeyDown(KeyPause) {
			log.Println("pausing")

			s.state = "pause"
			s.since = time.Now()
		}

	case "pause":
		if !GlContextSystem(nil).IsKeyDown(KeyPause) && GlContextSystem(nil).AnyKeyDown() {
			log.Println("restarting")

			s.state = "playing"
			s.since = time.Now()
		}
	}

	s.update.Publish(MessageUpdate{Delta: delta})
	return
}

func (s *gameStateSystem) OnQuit() *Observer   { return s.quit }
func (s *gameStateSystem) OnUpdate() *Observer { return s.update }
