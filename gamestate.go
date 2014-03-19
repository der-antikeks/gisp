package main

import (
	"log"
	"time"

	glfw "github.com/go-gl/glfw3"

	"github.com/der-antikeks/gisp/ecs"
)

type GameStateSystem struct {
	engine *ecs.Engine
	state  *ecs.Entity

	em *EntityManager
	im *InputManager
	wm *WindowManager
}

func NewGameStateSystem(em *EntityManager, im *InputManager, wm *WindowManager) ecs.System {
	return &GameStateSystem{
		em: em,
		im: im,
		wm: wm,
	}
}

func (s *GameStateSystem) AddedToEngine(e *ecs.Engine) error {
	s.engine = e

	e.Collection(GameStateType).Subscribe(s, func(en *ecs.Entity) {
		s.state = en
	}, func(en *ecs.Entity) {
		s.state = nil
	})

	return nil
}

func (s *GameStateSystem) RemovedFromEngine(e *ecs.Engine) error {
	e.Collection(GameStateType).Unsubscribe(s)
	return nil
}

func (s *GameStateSystem) Update(delta time.Duration) error {
	if s.im.IsKeyDown(glfw.KeyEscape) {
		s.engine.ShutDown()
		return nil
	}

	if s.state == nil {
		log.Println("initialize")
		s.em.Initalize()
		return nil
	}

	se := s.state.Get(GameStateType).(*GameStateComponent)
	switch se.State {
	case "init":
		log.Println("create splash screen")

		s.em.CreateSplashScreen()
		se.State = "splash"
		se.Since = time.Now()

	case "splash":
		if time.Now().After(se.Since.Add(5*time.Second)) || s.im.AnyKeyDown() {
			log.Println("create main menu")

			s.em.CreateMainMenu()
			se.State = "mainmenu"
			se.Since = time.Now()
		}

	case "mainmenu":
		if s.im.IsKeyDown(glfw.KeyEnter) {
			log.Println("starting game")

			se.State = "playing"
			se.Since = time.Now()
		}

	case "optionsmenu":
		if s.im.IsKeyDown(glfw.KeyEscape) {
			log.Println("back to main menu")

			se.State = "mainmenu"
			se.Since = time.Now()
		}

	case "playing":
		if s.im.IsKeyDown(glfw.KeyPause) {
			log.Println("pausing")

			se.State = "pause"
			se.Since = time.Now()
		}

	case "pause":
		if s.im.AnyKeyDown() {
			log.Println("restarting")

			se.State = "playing"
			se.Since = time.Now()
		}
	}

	return nil
}
